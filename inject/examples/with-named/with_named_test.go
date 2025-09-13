package main

import (
	"testing"

	"github.com/fxfn/x/inject"
)

func TestWithNamed(t *testing.T) {
	container := inject.Default()
	inject.RegisterNamed[string](container, "foo", "bar")
	foo := inject.GetNamed[string](container, "foo")

	inject.RegisterNamed[string](container, "bar", "baz")
	bar := inject.GetNamed[string](container, "bar")

	t.Run("should have a foo", func(t *testing.T) {
		if foo != "bar" {
			t.Errorf("expected foo to be bar, got %s", foo)
		}
	})

	t.Run("should have a bar", func(t *testing.T) {
		if bar != "baz" {
			t.Errorf("expected bar to be baz, got %s", bar)
		}
	})
}
