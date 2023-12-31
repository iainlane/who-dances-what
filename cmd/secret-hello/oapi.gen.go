// Package main provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.13.0 DO NOT EDIT.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

const (
	Jwt_authorizerScopes = "jwt_authorizer.Scopes"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Returns a secret greeting
	// (GET /secret-hello)
	GetSecretHello(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetSecretHello converts echo context to params.
func (w *ServerInterfaceWrapper) GetSecretHello(ctx echo.Context) error {
	var err error

	ctx.Set(Jwt_authorizerScopes, []string{"email"})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetSecretHello(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
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

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/secret-hello", wrapper.GetSecretHello)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/4RTzW7cPAx8FX9MgO/iv6Q334IAbXMrGrSXIAgYi7GVypRL0fVuFn73QvL+IFsUPVEi",
	"xeFwMNpB64fRM7EGaHYQqJ3E6va+7WmglHqdtcBJey/2jSRm/Eh8Z249M7X6TRw0cLmzIUwkS1XO5Fzx",
	"g/3MVXxoTdF6frHdJKjWM+Sg25GgeQ8DOWwKHPDNc4Gj7VBpxu3Z4NPtC26dR/PRy4D6nSRE5Aauyxpy",
	"sIZY4xJ+kjZOuhT6OVHQsic0JOXNHudA6HXW23cc0yxjiWP7A1zuDrcFHpfjAq+zwrIsSw6WX3xsUqsu",
	"VubeFwa5pVDMPWrcCHL4deR5VdZlDUueNIjFBj6UdXkFOYyofdK9CtQKadGTcwm8I43BUGjFjitPuMmC",
	"HUZHGbEZvWXNtEfNhHQSDhlmnRCp5S4LKjE8T5pZ/T9kKzwkDuvadwYa+ER6nyqf09wchMLoOaxmuK7r",
	"GFrPSpzoKG20Gh3apFqIvsF4og1GXtBAAsr387LZizP/nVyw0oIk4/lqZ9whvjk4FJqHP735ADSgdfC4",
	"POYQpmFA2UIDX49q7EkcgP/iOstK3ckKF0Iv0MBFdfor1T/aQuVweDYIOUxi0wcZp2dn26c1/4TCSzLP",
	"8jsAAP//NuQldYMDAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
