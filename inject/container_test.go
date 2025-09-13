package inject

import (
	"testing"
)

type IService interface{}

type Service struct {
	IService
}

func NewTestService(c *Container) IService {
	return &Service{}
}

func TestDefault(t *testing.T) {

	control := Default()
	Register[IService](control, NewTestService)

	t.Run("the container should be the same in both tests - 1", func(t *testing.T) {
		container := Default()
		if container == nil {
			t.Errorf("container should not be nil")
		}
		if container != control {
			t.Errorf("container should be the same as the control")
		}
		service, err := Resolve[IService](control)
		if err != nil {
			t.Errorf("error should be nil, got %v", err)
		}
		if service == nil {
			t.Errorf("service should not be nil")
		}
	})

	t.Run("the container should be the same in both tests - 2", func(t *testing.T) {
		container := Default()
		if container == nil {
			t.Errorf("container should not be nil")
		}
		if container != control {
			t.Errorf("container should be the same as the control")
		}
		service := Get[IService](container)
		if service == nil {
			t.Errorf("service should not be nil")
		}
	})
}

func TestNewContainer(t *testing.T) {
	container := NewContainer()

	if container == nil {
		t.Errorf("container should not be nil")
	}
}

func TestReset(t *testing.T) {
	container := NewContainer()
	Register[int](container, 1)
	container.Reset()

	if len(container.services) != 0 {
		t.Errorf("container.services should be empty")
	}
}

func TestRegister(t *testing.T) {
	container := NewContainer()
	Register[int](container, 1)

	service := Get[int](container)
	if service != 1 {
		t.Errorf("service should be 1, got %d", service)
	}
}

func TestRegisterNamed(t *testing.T) {
	container := NewContainer()
	RegisterNamed[int](container, "test", 1)

	service := GetNamed[int](container, "test")
	if service != 1 {
		t.Errorf("service should be 1, got %d", service)
	}
}
