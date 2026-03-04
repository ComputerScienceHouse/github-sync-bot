package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ComputerScienceHouse/github-sync-bot/ghauth"
	"github.com/joho/godotenv"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request, tokenInfo *ghauth.GhTokenInfo) {
	result := make(map[string]interface{})

	json.NewDecoder(r.Body).Decode(&result)

	webhookType := r.Header.Get("X-Github-Event")

	switch webhookType {
	case "pull_request_review":
		HandlePRReview(result, tokenInfo)
	}

	fmt.Println(result)
}

func HandlePRReview(result map[string]interface{}, tokenInfo *ghauth.GhTokenInfo) {
	// client := http.Client{}

	println(result["review"])

	_, err := http.NewRequest("GET", "https://api.github.com/repos/ComputerScienceHouse", nil)

	if err != nil {
		log.Println("error creating request")
		return
	}


}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No env file detected, make sure all secrets are loaded into the environment")
		// panic("Error loading .env file")
	}

	ghTokenInfo, err := ghauth.SetupGHAuth()

	if err != nil {
		log.Println("failed to retrieve github access token: ", err.Error())
		return
	}

	http.HandleFunc("POST /webhook", http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		HandleWebhook(w, r, ghTokenInfo)
	}))

	port := "8080"
	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

