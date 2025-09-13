package main

import (
	"fmt"

	"github.com/fxfn/x/inject"
)

func main() {
	container := inject.Default()

	inject.RegisterNamed[string](container, "foo", "bar")
	foo := inject.GetNamed[string](container, "foo")

	inject.RegisterNamed[string](container, "bar", "baz")
	bar := inject.GetNamed[string](container, "bar")

	fmt.Printf("foo: %s, bar: %s\n", foo, bar)
}
