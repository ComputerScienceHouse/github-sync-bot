package main

import (
	"fmt"
	"log"
	"net/http"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hi")

	fmt.Println(r.Body)
}

func main() {
	http.HandleFunc("POST /webhook", HandleWebhook)

	port := "8080"
	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
