package healthplanet

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "encoding/json"
    "errors"
    "time"
    "strconv"
    "log/slog"
)


type Client struct {
    url string
    auth HealthPlanetAuth
    Logger *slog.Logger
    Timezone *time.Location
}


func NewClient(url string, auth HealthPlanetAuth, logger *slog.Logger, timezone *time.Location) *Client{
    return &Client{url: url, auth: auth, Logger: logger, Timezone: timezone}
}



type InnerscanData struct {
    Date time.Time
    Weight float64
    BodyFat float64
}
type InnerscanDataMap map[string]*InnerscanData

func (c *Client) GetInnerscanData(from time.Time) (InnerscanDataMap, error){
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
    q.Set("from", from.Format("20060102150405"))
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
    resp_data := InnerscanResponse{}
    err = json.Unmarshal(body, &resp_data)
    if err != nil {
        return nil, err
    }

    return(resp_data.GetInnerscanDataMap(c.Timezone))
}

func (resp *InnerscanResponse) GetInnerscanDataMap(timezone *time.Location) (InnerscanDataMap, error) {
    ret := make(InnerscanDataMap)
    
    for _, d := range resp.Data {
        if _, ok := ret[d.Date]; !ok {
            ret[d.Date] = &InnerscanData{}
        }
        date, err := time.ParseInLocation("200601021504", d.Date, timezone)
        if err != nil {
            return nil, err
        }
        ret[d.Date].Date = date

        if d.Tag == "6021" {
            value, err := strconv.ParseFloat(d.KeyData, 64)
            if err != nil {
                return nil, err
            }
            ret[d.Date].Weight = value
        }
        if d.Tag == "6022" {
            value, err := strconv.ParseFloat(d.KeyData, 64)
            if err != nil {
                return nil, err
            }
            ret[d.Date].BodyFat = value
        }
    }

    return ret, nil
}

