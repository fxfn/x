package main

import (
	"testing"

	"github.com/fxfn/x/inject"
)

type MyService interface {
	DoSomething(t *testing.T)
}

type MyServiceImpl struct {
	MyService
}

func (s *MyServiceImpl) DoSomething(t *testing.T) {
	t.Logf("Doing something!\n")
}

func NewMyService(c *inject.Container) MyService {
	return &MyServiceImpl{}
}

func TestWithBasic(t *testing.T) {

	t.Run("should resolve myService", func(t *testing.T) {
		container := inject.NewContainer()
		inject.Register[MyService](container, NewMyService)
		myService, err := inject.Resolve[MyService](container)
		if err != nil {
			t.Fatalf("failed to resolve myService: %v", err)
		}

		myService.DoSomething(t)
	})

	t.Run("should return an error if not registered", func(t *testing.T) {
		container := inject.NewContainer()
		_, err := inject.Resolve[MyService](container)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
	})

	t.Run("error returned when service is not registered should be ErrServiceNotFound", func(t *testing.T) {
		container := inject.NewContainer()
		_, err := inject.Resolve[MyService](container)
		if err != inject.ErrServiceNotFound {
			t.Fatalf("expected ErrServiceNotFound, got %v", err)
		}
	})

	t.Run("error returned when service is not of the correct type should be ErrInvalidServiceType", func(t *testing.T) {
		container := inject.NewContainer()
		inject.Register[MyService](container, NewMyService)
		_, err := inject.Resolve[int](container)
		if err != inject.ErrInvalidServiceType {
			t.Fatalf("expected ErrInvalidServiceType, got %v", err)
		}
	})
}
