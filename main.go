package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var result map[string]interface{}

	json.NewDecoder(r.Body).Decode(&result)
	fmt.Println(result["review"])
}

func main() {
	http.HandleFunc("POST /webhook", HandleWebhook)

	port := "8080"
	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
