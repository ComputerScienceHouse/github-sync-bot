package ghauth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GhTokenInfo struct {
	token   string
	expTime time.Time
}

func (tokenInfo *GhTokenInfo) IsValid() bool {
	return tokenInfo.expTime.After(time.Now().Add(time.Duration(-30*1000)))
}

func (tokenInfo *GhTokenInfo) GetToken() string {
	if tokenInfo.IsValid() {
		return tokenInfo.token
	}

	tokenInfo.Update()
	return tokenInfo.GetToken()
}

func (tokenInfo *GhTokenInfo) Update() error {
	jwtToken := GetGithubJWTToken()
	client := &http.Client{}

	installationInfoReq, err := http.NewRequest("GET", "https://api.github.com/orgs/ComputerScienceHouse/installation", nil)

	if err != nil {
		log.Println("error creating request:", err)
		return err
	}

	installationInfoReq.Header.Add("Accept", "application/vnd.github+json")
	installationInfoReq.Header.Add("Authorization", "Bearer "+jwtToken)
	installationInfoReq.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	installationInfoResp, err := client.Do(installationInfoReq)
	if err != nil {
		log.Println("error authenticating app: " + err.Error())
		return err
	}

	installationInfoResult := make(map[string]interface{})

	json.NewDecoder(installationInfoResp.Body).Decode(&installationInfoResult)

	tokenUrl, ok := installationInfoResult["access_tokens_url"].(string)

	if !ok {
		log.Println("error parsing access token url response")
		return errors.New("error parsing access token url response")
	}

	accessTokenReq, err := http.NewRequest("POST", tokenUrl, nil)

	if err != nil {
		log.Println("error creating request:", err)
		return err
	}

	accessTokenReq.Header.Add("Accept", "application/vnd.github+json")
	accessTokenReq.Header.Add("Authorization", "Bearer "+jwtToken)
	accessTokenReq.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	accessTokenResp, err := client.Do(accessTokenReq)
	if err != nil {
		log.Println("error authenticating app: " + err.Error())
		return err
	}

	accessTokenResult := make(map[string]interface{})

	json.NewDecoder(accessTokenResp.Body).Decode(&accessTokenResult)

	accessToken, ok := accessTokenResult["token"].(string)

	if !ok {
		return errors.New("Failed to parse token response")
	}

	timeStr := accessTokenResult["expires_at"].(string)
	expTime, err := time.Parse(time.RFC3339, timeStr)

	if !ok {
		log.Println("Failed to parse expiration time:", err.Error())
		return err
	}

	
	tokenInfo.token = accessToken
	tokenInfo.expTime = expTime

	return nil
}

var tokenInfo GhTokenInfo

func SetupGHAuth() (*GhTokenInfo, error) {
	tokenInfo = GhTokenInfo{}

	err := tokenInfo.Update()

	if err != nil {
		return nil, err
	}

	return &tokenInfo, nil
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
	println(key)
	keyBlock, _ := pem.Decode([]byte(key))

	if keyBlock == nil {
		log.Println("Failed to decode PEM block from private key")
		return nil, errors.New("Failed to decode PEM block from private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	return privateKey, err
}
