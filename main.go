package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var result map[string]interface{}

	json.NewDecoder(r.Body).Decode(&result)
	fmt.Println(result)
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No env file detected, make sure all secrets are loaded into the environment")
		// panic("Error loading .env file")
	}

	client := &http.Client{}

	token, err := AuthGithubApp(client)

	if err != nil {
		log.Println("failed to retrieve github access token: ", err.Error())
		return
	}

	log.Println(token)

	http.HandleFunc("POST /webhook", HandleWebhook)

	port := "8080"
	log.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func AuthGithubApp(client *http.Client) (string, error) {
	jwtToken := GetGithubJWTToken()

	installationInfoReq, err := http.NewRequest("GET", "https://api.github.com/orgs/ComputerScienceHouse/installation", nil)

	if err != nil {
		log.Println("error creating request:", err)
		return "", err
	}

	installationInfoReq.Header.Add("Accept", "application/vnd.github+json")
	installationInfoReq.Header.Add("Authorization", "Bearer "+jwtToken)
	installationInfoReq.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	installationInfoResp, err := client.Do(installationInfoReq)
	if err != nil {
		log.Println("error authenticating app: " + err.Error())
		return "", err
	}

	var installationInfoResult map[string]interface{}

	json.NewDecoder(installationInfoResp.Body).Decode(&installationInfoResult)

	tokenUrl, ok := installationInfoResult["access_tokens_url"].(string)

	if !ok {
		log.Println("error parsing access token url response")
		return "", errors.New("error parsing access token url response")
	}

	accessTokenReq, err := http.NewRequest("POST", tokenUrl, nil)

	if err != nil {
		log.Println("error creating request:", err)
		return "", err
	}

	accessTokenReq.Header.Add("Accept", "application/vnd.github+json")
	accessTokenReq.Header.Add("Authorization", "Bearer "+jwtToken)
	accessTokenReq.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	accessTokenResp, err := client.Do(accessTokenReq)
	if err != nil {
		log.Println("error authenticating app: " + err.Error())
		return "", err
	}

	var accessTokenResult map[string]interface{}

	json.NewDecoder(accessTokenResp.Body).Decode(&accessTokenResult)

	accessToken, ok := accessTokenResult["token"].(string)

	if !ok {
		return "", errors.New("Failed to parse token response")
	}

	return accessToken, nil
}

func GetGithubJWTToken() string {
	clientId := os.Getenv("GITHUB_APP_CLIENT_ID")
	keyString := os.Getenv("GITHUB_APP_PRIVATE_KEY")

	if len(clientId) == 0 {
		log.Fatal("Could not load github app client id")
	}

	if len(keyString) == 0 {
		log.Fatal("Could not load github app private key")
	}

	key, err := GetPrivateKeyFromStr(keyString)

	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	timeSeconds := time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": timeSeconds,
		"exp": timeSeconds + 600,
		"iss": clientId,
	})

	tokenString, err := token.SignedString(key)

	if err != nil {
		log.Fatalf("Error signing token: %s\n", err.Error())
	}

	return tokenString
}

func GetPrivateKeyFromStr(key string) (*rsa.PrivateKey, error) {
	keyBlock, _ := pem.Decode([]byte(key))

	if keyBlock == nil {
		log.Println("Failed to decode PEM block from private key")
		return nil, errors.New("Failed to decode PEM block from private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	return privateKey, err
}
