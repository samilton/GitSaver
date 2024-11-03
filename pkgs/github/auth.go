package github

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type InstallationToken struct {
    Token        string    `json:"token"`
    ExpiresAt    string    `json:"expires_at"`
}

type GitHubApp struct {
    AppID          int64
    InstallationID int64
    PrivateKey     []byte
}

func NewGitHubApp(appID int64, installationID int64, privateKeyPath string) (*GitHubApp, error) {
    privateKey, err := ioutil.ReadFile(privateKeyPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read private key: %v", err)
    }

    return &GitHubApp{
        AppID:          appID,
        InstallationID: installationID,
        PrivateKey:     privateKey,
    }, nil
}

func (app *GitHubApp) generateJWT() (string, error) {
    // Parse the private key
    block, _ := pem.Decode(app.PrivateKey)
    if block == nil {
        return "", fmt.Errorf("failed to parse PEM block containing private key")
    }

    privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(app.PrivateKey)
    if err != nil {
        return "", fmt.Errorf("failed to parse private key: %v", err)
    }

    // Create the JWT claims
    now := time.Now()
    claims := jwt.RegisteredClaims{
        IssuedAt:  jwt.NewNumericDate(now),
        ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
        Issuer:    fmt.Sprintf("%d", app.AppID),
    }

    // Create the token
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

    // Sign the token
    signedToken, err := token.SignedString(privateKey)
    if err != nil {
        return "", fmt.Errorf("failed to sign token: %v", err)
    }

    return signedToken, nil
}

func (app *GitHubApp) GetInstallationToken() (string, error) {
    jwt, err := app.generateJWT()
    if err != nil {
        return "", err
    }

    // Create the request to get an installation token
    url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", app.InstallationID)
    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        return "", fmt.Errorf("failed to create request: %v", err)
    }

    // Add required headers
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
    req.Header.Set("Accept", "application/vnd.github.v3+json")

    // Send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to send request: %v", err)
    }
    defer resp.Body.Close()

    // Read and parse the response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response: %v", err)
    }

    if resp.StatusCode != http.StatusCreated {
        return "", fmt.Errorf("failed to get installation token: %s", string(body))
    }

    var token InstallationToken
    if err := json.Unmarshal(body, &token); err != nil {
        return "", fmt.Errorf("failed to parse response: %v", err)
    }

    return token.Token, nil
}
