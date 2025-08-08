package main

import (
    "fmt"
    "log/slog"
    "os"
    "gopkg.in/yaml.v3"
    "github.com/kamaboko123/tanita2csv/pkg/healthplanet"
    "time"
)

type Config struct {
    URL          string `yaml:"url"`
    TokenFile    string `yaml:"token_file"`
    ClientID     string `yaml:"client_id"`
    ClientSecret string `yaml:"client_secret"`
    Timezone     string `yaml:"timezone"`
}

func loadConfig(filePath string) (*Config, error) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    var config Config
    err = yaml.Unmarshal(content, &config)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    return &config, nil
}

func main() {
    fmt.Println("Hello, World!")
    
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))

    // Load config file
    config, err := loadConfig("config.yml")
    if err != nil {
        logger.Error("Failed to load config", "error", err)
        return
    }
    timezone, err := time.LoadLocation(config.Timezone)
    if err != nil {
        logger.Error("Failed to load timezone. Timezone must be a valid IANA timezone.(e.g, Asia/Tokyo, America/New_York and so on)")
        return
    }

    // Initialize HealthPlanet Auth
    auth := healthplanet.NewSimpleAuth(config.URL, config.ClientID, config.ClientSecret, config.TokenFile, logger)

    // check token file existence
    _, err = os.Stat(config.TokenFile)
    if err != nil{
        logger.Warn("Token file does not exist, starting authentication process")

        err = auth.Auth()
        if err != nil {
            logger.Error("Failed to authenticate with HealthPlanet, abort.", "error", err)
            os.Exit(10)
        }
    }

    err = auth.RefreshToken()
    if err != nil {
        logger.Error("Failed to refresh token, abort.", "error", err)
        os.Exit(11)
    }

    // Init HealthPlanet Client
    hpClient := healthplanet.NewClient(config.URL, auth, logger, timezone)
    
    // Get Innerscan Data
    from := time.Now().Add(-24 * 365 * time.Hour)
    innerscanData, err := hpClient.GetInnerscanData(from)
    if err != nil {
        logger.Error("Failed to get Innerscan data", "error", err)
        return
    }
    logger.Info("Innerscan Data Retrieved", "data_count", len(innerscanData))
    for date, data := range innerscanData {
        logger.Info("Innerscan Data", "date", date, "data", data.String())
    }
}
