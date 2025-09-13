package main

import (
	"fmt"

	"github.com/fxfn/x/inject"
)

type Logger interface {
	Info(message string)
	Error(message string)
}

type ConsoleLogger struct {
}

func (l *ConsoleLogger) Info(message string) {
	fmt.Println("LOG", message)
}

func (l *ConsoleLogger) Error(message string) {
	fmt.Println("ERROR", message)
}

func NewConsoleLogger(c *inject.Container) Logger {
	return &ConsoleLogger{}
}

func main() {
	container := inject.NewContainer()
	inject.Register[Logger](container, NewConsoleLogger)

	logger, err := inject.Resolve[Logger](container)
	if err != nil {
		panic(err)
	}

	logger.Info("Hello, world!")
}
