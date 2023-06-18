package main

import (
	"github.com/labstack/echo/v4"
)

type secretHelloAPI struct {
}

func newAPI() *secretHelloAPI {
	return &secretHelloAPI{}
}

func (api *secretHelloAPI) GetSecretHello(ctx echo.Context) error {
	return ctx.JSON(200, "Hello, secret world!")
}
