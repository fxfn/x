package main

import "github.com/fxfn/x/schema"

func main() {
	router := schema.NewRouter()

	openApi := schema.OpenAPI(router.Engine, &schema.OpenAPIOpts{
		Title:       "Basic API",
		Version:     "1.0.0",
		Description: "Basic API",
		Contact:     "John Doe",
		License:     "MIT",
		OutputFile:  "openapi.json",
	})

	router.GET("/swagger.json", openApi.HandleGetSwagger)
	router.Run(":8080")
}
