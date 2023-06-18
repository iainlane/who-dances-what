package main

import (
	"github.com/iainlane/who-dances-what/internal/lambda"
)

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=../../api/types.cfg.yaml ../../api/secret-hello/api.yaml

func RegisterHandlersAdapter(router lambda.EchoRouter, handler ServerInterface) {
	RegisterHandlers(router, handler)
}

func main() {
	lambda.StartServer[ServerInterface](GetSwagger, RegisterHandlersAdapter, newAPI())
}
