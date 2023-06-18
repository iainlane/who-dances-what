package lambda

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
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

	// swagger.Server = nil

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = loggerMiddleware
	e.Use(loggerMiddleware.Hook())
	e.Use(middleware.OapiRequestValidator(swagger))

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
