package bootstrap

import (
	"log"
)

type Application struct {
	Logger *log.Logger
}

func NewInitializeBootsrap() Application {
	app := Application{}
	return app
}
