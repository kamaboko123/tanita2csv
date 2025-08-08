package healthplanet

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

/*
SimpleAuth is a simple authentication client for HealthPlanet API.

It supports OAuth2 authorization code flow and token refresh.
It is designed to be used in a command line application.

It saves the token to a file and loads it from the file,
And If the token is expired or needs to be refreshed, it will automatically refresh the token.

The clinet satisfies the HealthPlanetAuth interface.

Usage:
```
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
auth := NewSimpleAuth("https://www.healthplanet.jp", "your_client_id", "your_client_secret", "token.json", logger)

// Fist, We need to start the authentication process to ger the initial token.
err := auth.Auth()
if err != nil {
    logger.Error("Failed to start authentication", "error", err)
    return
}

// After that, we can use the token to access the API.
if !auth.IsTokenValid() {
    logger.Error("Token is not valid, please reauthenticate")
    return
}
token, err := auth.GetToken()
if err != nil {
    logger.Error("Failed to get token", "error", err)
    return
}
fmt.Println("Access token:", token)
*/

const TokenRefreshThreshold = 60 * 60 * 24 * 7 // 1 week

type Token struct {
    AccessToken string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn int64 `json:"expires_in"`
    
    // metadata
    CreateDate int64 `json:"create_date,omitempty"`
}

func (t *Token) IsTokenExpired() bool {
    return t.CreateDate + t.ExpiresIn < time.Now().Unix()
}

func (t *Token) IsTokenNeedRefresh() bool {
    return t.CreateDate + t.ExpiresIn - TokenRefreshThreshold < time.Now().Unix()
}



type SimpleAuth struct {
    Url string
    ClientId string
    clientSecret string
    token *Token
    TokenFile string
    Logger *slog.Logger
}

func (a *SimpleAuth) GetToken() (string, error) {
    var err error
    err = a.RefreshToken()
    if err != nil {
        return "", err
    }
    err = a.ValidateToken()
    if err != nil {
        return "", err
    }


    return a.token.AccessToken, nil
}


func NewSimpleAuth(Url string, ClientId string, ClientSecret string, TokenFile string, logger *slog.Logger) *SimpleAuth {
    auth := SimpleAuth{
        Url: Url,
        ClientId: ClientId,
        clientSecret: ClientSecret,
        TokenFile: TokenFile,
        token: nil,
        Logger: logger,
    }
    auth.token = &Token{CreateDate: 0}

    return &auth
}

func (a *SimpleAuth) SaveToken() error {
    data, err := json.MarshalIndent(a.token, "", "  ")
    if err != nil {
        return err
    }

    f, err := os.OpenFile(a.TokenFile, os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = f.Write(data)
    if err != nil {
        return err
    }

    return nil
}

func (a *SimpleAuth) LoadToken() error {
    data, err := os.ReadFile(a.TokenFile)
    if err != nil {
        return err
    }

    a.token = &Token{}
    err = json.Unmarshal(data, a.token)
    if err != nil {
        return err
    }

    return nil
}

func (a *SimpleAuth) ValidateToken() error {
    // TODO: implement!!!!!
    return nil
}

/*
func (t *Token) IsTokenValid() bool {
    if t.AccessToken == "" {
        return false
    }
    if t.CreateDate == 0 {
        return false
    }
    // TODO: implement token validation by actual API call

    return !t.IsTokenExpired()
}*/

func (a *SimpleAuth) BuildAuthURL() (string, error) {
    u, err := url.Parse(a.Url)
    if err != nil {
        return "", err
    }

    u.Path = "/oauth/auth"

    q := u.Query()
    q.Set("client_id", a.ClientId)
    q.Set("client_secret", a.clientSecret)
    q.Set("redirect_uri", "http://localhost")
    q.Set("response_type", "code")
    q.Set("scope", "innerscan,sphygmomanometer,pedometer,smug")
    u.RawQuery = q.Encode()

    return u.String(), nil
}


func (a *SimpleAuth) GetTokenWithCode(code string) (*Token, error) {
    u, err := url.Parse(a.Url)
    if err != nil {
        return nil, err
    }

    u.Path = "/oauth/token"
    q := u.Query()
    q.Set("client_id", a.ClientId)
    q.Set("client_secret", a.clientSecret)
    q.Set("redirect_uri", "http://localhost")
    q.Set("grant_type", "authorization_code")
    q.Set("code", code)
    u.RawQuery = q.Encode()

    req, err := http.NewRequest("POST", u.String(), nil)
    if err != nil {
        return nil, err
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode != 200 {
        return nil, errors.New("[HealthPlanet]Failed to get token")
    }

    body, _ := io.ReadAll(resp.Body)
    token := Token{}
    err = json.Unmarshal(body, &token)
    if err != nil {
        return nil, err
    }
    a.token = &token
    a.token.CreateDate = time.Now().Unix()
    

    return &token, nil
}

func (a *SimpleAuth) RefreshToken() error{
    err := a.LoadToken()
    if err != nil {
        return err
    }

    if !a.token.IsTokenNeedRefresh(){
        return nil
    }
    fmt.Println("Token is need to refresh")

    if a.token == nil {
        return errors.New("[HealthPlanet]Token is not initialized")
    }

    u, err := url.Parse(a.Url)
    if err != nil {
        return err
    }

    u.Path = "/oauth/token"
    q := u.Query()
    q.Set("client_id", a.ClientId)
    q.Set("client_secret", a.clientSecret)
    q.Set("redirect_uri", "http://localhost")
    q.Set("grant_type", "refresh_token")
    q.Set("refresh_token", a.token.RefreshToken)
    u.RawQuery = q.Encode()
    fmt.Println(u.String())

    req, err := http.NewRequest("POST", u.String(), nil)
    if err != nil {
        return err
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    if resp.StatusCode != 200 {
        return errors.New("[HealthPlanet]Failed to refresh token")
    }

    body, _ := io.ReadAll(resp.Body)
    err = json.Unmarshal(body, a.token)
    if err != nil {
        return err
    }
    fmt.Println("Success to refresh token")

    err = a.SaveToken()
    if err != nil {
        return err
    }

    return nil
}


func (a *SimpleAuth) Auth() error {
    // check dump file exists
    _, err := os.Stat(a.TokenFile)
    if err == nil {
        return errors.New("[HealthPlanet]Token file already exists. If you want to reinitilize, please remove the file")
    }

    url, err := a.BuildAuthURL()
    if err != nil {
        return err
    }
    fmt.Printf("Access to follwing URL with browser: %s\n", url)

    fmt.Printf("And enter code:")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    code := scanner.Text()

    _, err = a.GetTokenWithCode(code)
    if err != nil {
        return err
    }
    
    err = a.SaveToken()
    if err != nil {
        return err
    }

    fmt.Println("Success to Authenticate!")
    return nil
}

