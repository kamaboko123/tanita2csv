package healthplanet

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "encoding/json"
    "errors"
    "time"
    "log/slog"
)


type Client struct {
    url string
    auth HealthPlanetAuth
    Logger *slog.Logger
}


func NewClient(url string, auth HealthPlanetAuth, logger *slog.Logger) *Client{
    return &Client{url: url, auth: auth, Logger: logger}
}

func (c *Client) GetInnerscanData(from time.Time, to time.Time) (*Innerscan, error){
    c.Logger.Debug(fmt.Sprintf("GetInnerscanData called with from: %s, to: %s", from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))

    u, err := url.Parse(c.url)
    if err != nil {
        return nil, err
    }
    
    token, err := c.auth.GetToken()
    if err != nil {
        return nil, fmt.Errorf("Failed to get access token: %w", err)
    }

    u.Path = "/status/innerscan.json"
    q := u.Query()
    q.Set("access_token", token)
    q.Set("date", "1") // from, to date type: "1" means mesurement date
    q.Set("from", from.Format("20060102150405"))
    q.Set("to", to.Format("20060102150405"))
    q.Set("tag", "6021,6022") // 6021: Weight, 6022: Body Fat

    u.RawQuery = q.Encode()

    c.Logger.Debug(fmt.Sprintf("Access to: %s", u.String()))

    req, err := http.NewRequest("GET", u.String(), nil)
    if err != nil {
        return nil, err
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode != 200 {
        return nil, errors.New("[HealthPlanet]Failed to get innerscan data")
    }

    body, _ := io.ReadAll(resp.Body)
    c.Logger.Debug(fmt.Sprintf("Response: %s", body))
    respData := InnerscanResponse{}
    err = json.Unmarshal(body, &respData)
    if err != nil {
        return nil, err
    }

    innerscan, err := respData.ToInnerscan()
    if err != nil {
        return nil, fmt.Errorf("Failed to convert response data to Innerscan: %w", err)
    }

    return innerscan, nil
}

