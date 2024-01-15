package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("golang-boilerplate starting...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
