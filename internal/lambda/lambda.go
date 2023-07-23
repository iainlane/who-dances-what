package lambda

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sirupsen/logrus"
	echologrus "github.com/spirosoik/echo-logrus"
)

type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

func createLoggingAdapter(e *echoadapter.EchoLambdaV2, logger *logrus.Logger) func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		body := []byte(req.Body)
		if req.IsBase64Encoded {
			base64Body, err := base64.StdEncoding.DecodeString(req.Body)
			if err != nil {
				return core.GatewayTimeoutV2(), err
			}
			body = base64Body
		}

		logger.WithFields(logrus.Fields{
			"routeKey":              req.RouteKey,
			"stage":                 req.RequestContext.Stage,
			"contextHTTPMethod":     req.RequestContext.HTTP.Method,
			"contextHTTPPath":       req.RequestContext.HTTP.Path,
			"contextHTTProtocol":    req.RequestContext.HTTP.Protocol,
			"rawPath":               req.RawPath,
			"queryString":           req.RawQueryString,
			"queryStringParameters": req.QueryStringParameters,
			"body":                  string(body),
		}).Debug("API called")

		return e.ProxyWithContext(ctx, req)
	}
}

type jwksCache struct {
	*jwk.Cache
}

// use envconfig to get issuer and client_id
type oidcConfig struct {
	Issuer   string `envconfig:"ISSUER" required:"true"`
	ClientID string `envconfig:"CLIENT_ID" required:"true"`
}

func authFunc(refreshCtx, reqCtx context.Context, jwksCache *jwksCache, input *openapi3filter.AuthenticationInput) error {
	var oidcConfig oidcConfig
	err := envconfig.Process("", &oidcConfig)
	if err != nil {
		return fmt.Errorf("failed to load required config from environment: %w", err)
	}
	if oidcConfig.Issuer == "" {
		return fmt.Errorf("required environment variable ISSUER is set to an empty string")
	}
	if oidcConfig.ClientID == "" {
		return fmt.Errorf("required environment variable CLIENT_ID is set to an empty string")
	}

	headers := input.RequestValidationInput.Request.Header

	//get the jwks from the issuer
	oidcDiscoveryURL := input.SecurityScheme.OpenIdConnectUrl
	oidcDiscoveryURL = strings.Replace(oidcDiscoveryURL, "${issuer}", oidcConfig.Issuer, 1)

	// fetch that with a retry using go-retryablehttp
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 5 * time.Second
	resp, err := retryClient.Get(oidcDiscoveryURL)
	if err != nil {
		return fmt.Errorf("failed to fetch oidc discovery url: %w", err)
	}
	defer resp.Body.Close()

	// parse the response
	var discoveryResp struct {
		JwksURI string `json:"jwks_uri"`
	}
	err = json.NewDecoder(resp.Body).Decode(&discoveryResp)
	if err != nil {
		return fmt.Errorf("failed to decode oidc discovery response: %w", err)
	}

	if !jwksCache.IsRegistered(discoveryResp.JwksURI) {
		err = jwksCache.Register(discoveryResp.JwksURI)
		if err != nil {
			return fmt.Errorf("failed to register URI in jwks cache: %w", err)
		}

		_, err = jwksCache.Refresh(refreshCtx, discoveryResp.JwksURI)
		if err != nil {
			return fmt.Errorf("failed to fetch jwks: %w", err)
		}
	}

	keyset, err := jwk.Fetch(reqCtx, discoveryResp.JwksURI)
	if err != nil {
		return fmt.Errorf("failed to fetch jwks: %w", err)
	}

	token, err := jwt.ParseHeader(
		headers,
		"Authorization",
		jwt.WithKeySet(keyset),
		jwt.WithValidate(true),
		jwt.WithIssuer(oidcConfig.Issuer),
		jwt.WithClaimValue("client_id", oidcConfig.ClientID),
	)
	if err != nil {
		return fmt.Errorf("failed to parse jwt: %w", err)
	}

	if len(input.Scopes) == 0 {
		return nil
	}

	scopes, found := token.Get("scope")
	if !found {
		wantScopes := strings.Join(input.Scopes, " ")
		return fmt.Errorf("jwt has no scope claim, required scopes: %s", wantScopes)
	}

	scopesString, ok := scopes.(string)
	if !ok {
		return fmt.Errorf("jwt scope claim is not a string")
	}

	scopeSet := make(map[string]struct{})
	for _, scope := range strings.Split(scopesString, " ") {
		scopeSet[scope] = struct{}{}
	}

	// finally check each of the `input.Scopes` are in the jwt
	for _, scope := range input.Scopes {
		if _, ok := scopeSet[scope]; !ok {
			return fmt.Errorf("jwt does not have required scope: %s", scope)
		}
	}

	return nil
}

func StartServer[T any](
	getSwaggerFunc func() (*openapi3.T, error),
	registerHandlersFunc func(router EchoRouter, handler T),
	handler T,
) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	defer logger.Info("Shutting down")

	logger.SetFormatter(&logrus.JSONFormatter{})
	loggerMiddleware := echologrus.NewLoggerMiddleware(logger)

	swagger, err := getSwaggerFunc()
	if err != nil {
		logger.Fatalf("Error loading swagger spec\n: %s", err)
	}

	serverCtx := context.Background()
	serverCtx, cancel := context.WithCancel(serverCtx)
	defer cancel()

	jwksCache := &jwksCache{
		Cache: jwk.NewCache(serverCtx),
	}

	// swagger.Server = nil
	options := &middleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: func(reqCtx context.Context, input *openapi3filter.AuthenticationInput) error {
				return authFunc(serverCtx, reqCtx, jwksCache, input)
			},
		},
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = loggerMiddleware
	e.Use(loggerMiddleware.Hook())
	e.Use(middleware.OapiRequestValidatorWithOptions(swagger, options))

	registerHandlersFunc(e, handler)

	addr := ":8080"
	logger.WithField("address", addr).Info("Starting server")

	// are we running in AWS Lambda?
	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		logger.Info("Running in AWS Lambda")
		echoLambda := echoadapter.NewV2(e)
		loggingProxy := createLoggingAdapter(echoLambda, logger)
		lambda.Start(loggingProxy)
		return
	}

	e.Logger.Fatal(e.Start(addr))
}
