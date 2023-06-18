package main

import (
	"github.com/labstack/echo/v4"
)

type helloAPI struct {
}

func newAPI() *helloAPI {
	return &helloAPI{}
}

func (api *helloAPI) GetHello(ctx echo.Context) error {
	return ctx.JSON(200, "Hello, world!")
}
