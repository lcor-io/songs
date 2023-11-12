package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	component := hello("Leo")

	http.Handle("/", templ.Handler(component))

	fmt.Println("Listening on port 42069")
	if err := http.ListenAndServe(":42069", nil); err != nil {
		panic(err)
	}
}
