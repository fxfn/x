package inject

import (
	"errors"
	"reflect"
)

var (
	ErrServiceNotFound    = errors.New("service not found")
	ErrInvalidServiceType = errors.New("invalid service type")
)

var container *Container

type Container struct {
	services map[any]interface{}
}

type RegistrationValue interface{}

func NewContainer() *Container {
	return &Container{
		services: make(map[any]interface{}),
	}
}

func (c *Container) Reset() *Container {
	(*c) = *NewContainer()
	return c
}

func (c *Container) CreateChild() *Container {
	return &Container{
		services: make(map[any]interface{}),
	}
}

func Default() *Container {
	if container == nil {
		container = &Container{
			services: make(map[any]interface{}),
		}
	}

	return container
}

func Register[T any](c *Container, factory RegistrationValue) {
	c.services[reflect.TypeOf((*T)(nil)).Elem()] = factory
}

func RegisterNamed[T any](c *Container, name interface{}, factory RegistrationValue) {
	// check if we already have a service with this name
	if existing, ok := c.services[name]; ok {
		// existing should be a slice of factories
		factories := existing.([]RegistrationValue)
		c.services[name] = append(factories, factory)
	} else {
		c.services[name] = []RegistrationValue{factory}
	}
}

func RegisterSingleton[T any](c *Container, factory RegistrationValue) {
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	// check if the factory is a function that takes a *Container parameter
	if factoryType.Kind() == reflect.Func &&
		factoryType.NumIn() == 1 &&
		factoryType.In(0) == reflect.TypeOf((*Container)(nil)) {

		// call the factory function with the container
		results := factoryValue.Call([]reflect.Value{reflect.ValueOf(c)})
		if len(results) > 0 {
			c.services[reflect.TypeOf((*T)(nil)).Elem()] = results[0].Interface()
		}
	} else {
		// store the value directly
		c.services[reflect.TypeOf((*T)(nil)).Elem()] = factory
	}
}

func Get[T any](c *Container) T {
	var zero T
	service, ok := c.services[reflect.TypeOf((*T)(nil)).Elem()]
	if !ok {
		return zero
	}

	// Check if it's a factory function (transient)
	if factory, ok := service.(func(c *Container) T); ok {
		return factory(c)
	}

	// otherwise, its a singleton instance
	result, ok := service.(T)
	if !ok {
		return zero
	}

	return result
}

func GetNamed[T any](c *Container, name interface{}) T {
	var zero T
	service, ok := c.services[name]
	if !ok {
		return zero
	}

	// Named services are stored as slices, get the first one
	if factories, ok := service.([]RegistrationValue); ok {
		if len(factories) == 0 {
			return zero
		}

		factory := factories[0]

		// Check if it's a factory function
		factoryValue := reflect.ValueOf(factory)
		factoryType := factoryValue.Type()

		if factoryType.Kind() == reflect.Func &&
			factoryType.NumIn() == 1 &&
			factoryType.In(0) == reflect.TypeOf((*Container)(nil)) {
			// call the factory function with the container
			results := factoryValue.Call([]reflect.Value{reflect.ValueOf(c)})
			if len(results) > 0 {
				if result, ok := results[0].Interface().(T); ok {
					return result
				}
			}
		} else {
			// store the value directly
			if result, ok := factory.(T); ok {
				return result
			}
		}
	}

	return zero
}

func GetAllNamed[T any](c *Container, name interface{}) []T {
	var result []T
	services, ok := c.services[name]
	if !ok {
		return []T{}
	}

	if factories, ok := services.([]RegistrationValue); ok {
		for _, factory := range factories {
			// Check if it's a factory function
			factoryValue := reflect.ValueOf(factory)
			factoryType := factoryValue.Type()

			if factoryType.Kind() == reflect.Func &&
				factoryType.NumIn() == 1 &&
				factoryType.In(0) == reflect.TypeOf((*Container)(nil)) {
				// call the factory function with the container
				results := factoryValue.Call([]reflect.Value{reflect.ValueOf(c)})
				if len(results) > 0 {
					if service, ok := results[0].Interface().(T); ok {
						result = append(result, service)
					}
				}
			} else {
				// store the value directly
				if service, ok := factory.(T); ok {
					result = append(result, service)
				}
			}
		}
		return result
	}

	return []T{}
}

func Resolve[T any](c *Container) (T, error) {
	var zero T
	requestedType := reflect.TypeOf((*T)(nil)).Elem()
	service, ok := c.services[requestedType]
	if !ok {
		// Check if any type-based services are registered (exclude named services)
		hasTypeBasedServices := false
		for key := range c.services {
			if _, isType := key.(reflect.Type); isType {
				hasTypeBasedServices = true
				break
			}
		}

		if !hasTypeBasedServices {
			return zero, ErrServiceNotFound
		}
		// Type-based services exist but not the requested type
		return zero, ErrInvalidServiceType
	}

	// Check if it's a factory function (transient)
	if factory, ok := service.(func(c *Container) T); ok {
		return factory(c), nil
	}

	// Otherwise, it's a singleton instance
	result, ok := service.(T)
	if !ok {
		return zero, ErrInvalidServiceType
	}

	return result, nil
}
