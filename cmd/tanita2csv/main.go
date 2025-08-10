package main

import (
    "fmt"
    "log/slog"
    "os"
    "gopkg.in/yaml.v3"
    "github.com/kamaboko123/tanita2csv/pkg/healthplanet"
    "time"
    "flag"
)

const Version = "1.0.0"

type Config struct {
    URL          string `yaml:"url"`
    TokenFile    string `yaml:"token_file"`
    ClientID     string `yaml:"client_id"`
    ClientSecret string `yaml:"client_secret"`
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


type DateValue struct {
    time.Time
}
type FromDateValue DateValue
type ToDateValue DateValue

func (d *DateValue) String() string {
    if d == nil{
        return ""
    }
    return d.Format("2006-01-02")
}

func (f *FromDateValue) Set(value string) error {
    date, err := time.Parse("2006-01-02", value)
    if err != nil {
        return fmt.Errorf("invalid date format for from date: %w", err)
    }
    f.Time = date
    return nil
}

func (t *ToDateValue) Set(value string) error {
    date, err := time.Parse("2006-01-02", value)
    if err != nil {
        return fmt.Errorf("invalid date format for to date: %w", err)
    }
    t.Time = date
    return nil
}


type RunOption struct {
    configFile string   // config file path
    mode string         // auth, dump
    from time.Time      // YYYYMMDD
    to   time.Time      // YYYYMMDD
    output string       // output file path
    debug bool          // debug mode
}

func getArgs() *RunOption {
    runOption := &RunOption{}

    c := flag.String("c", "config.yml", "Config file path")

    m := flag.String("m", "", "Mode to run: auth or dump")

    f := &FromDateValue{Time: time.Now().AddDate(0, 0, -90)} // Default is 3 months ago
    flag.Var(f, "f", "From date (YYYY-MM-DD) for dump mode. Default is 3 months ago from today.")

    t := &ToDateValue{Time: time.Now()} // Default is today
    flag.Var(t, "t", "To date (YYYY-MM-DD) for dump mode. Default is today.")

    o := flag.String("o", "", "Output file path")

    d := flag.Bool("v", false, "Debug mode")
    
    version := flag.Bool("version", false, "Show version information")

    flag.Parse()

    if *version {
        fmt.Printf("tanita2csv version %s\n", Version)
        os.Exit(0)
    }

    runOption.configFile = *c
    runOption.debug = *d
    runOption.mode = *m

    switch runOption.mode {
    case "dump":
        runOption.from = f.Time.Truncate(24 * time.Hour)
        runOption.to = t.Time.Truncate(24 * time.Hour)
        runOption.from.Truncate(24 * time.Hour)
        runOption.to.Truncate(24 * time.Hour)

        if runOption.from.After(runOption.to) {
            fmt.Println("From date cannot be after To date.")
            os.Exit(1)
        }
        if runOption.to.Sub(runOption.from) > 90*24*time.Hour{
            fmt.Println("The date range cannot exceed 90 days.")
            os.Exit(1)
        }
        runOption.output = *o
    case "auth":
    default:
        fmt.Println("Invalid mode. Use -m auth or -m dump")
        os.Exit(1)
    }

    return runOption
}

func main() {
    runOption := getArgs()

    loglevel := slog.LevelWarn
    if runOption.debug {
        loglevel = slog.LevelDebug
    }

    logger := slog.New(
        slog.NewTextHandler(
            os.Stderr,
            &slog.HandlerOptions{
                Level: loglevel,
            },
        ),
    )

    // Load config file
    config, err := loadConfig(runOption.configFile)
    if err != nil {
        logger.Error("Failed to load config", "error", err)
        return
    }

    // Initialize HealthPlanet Auth
    auth := healthplanet.NewSimpleAuth(config.URL, config.ClientID, config.ClientSecret, config.TokenFile, logger)

    // check token file existence
    if runOption.mode == "auth" {
        logger.Info("Starting authentication process")
        err = auth.Auth()
        if err != nil {
            logger.Error("Failed to authenticate with HealthPlanet, abort.", "error", err)
            os.Exit(10)
        }
        logger.Info("Authentication successful, token file created", "token_file", config.TokenFile)
        os.Exit(0)
    }

    if runOption.mode == "dump" {
        _, err = os.Stat(config.TokenFile); if err != nil {
            logger.Warn("Token file does not exist, please run in auth mode first.", "error", err)
            os.Exit(1)
        }

        err = auth.RefreshToken()
        if err != nil {
            logger.Error("Failed to refresh token, abort. Please reauthenticate with auth mode.", "error", err)
            os.Exit(11)
        }

        // Init HealthPlanet Client
        hpClient := healthplanet.NewClient(config.URL, auth, logger)

        // Get Innerscan Data
        // Note: between `from` and `to` is limited to 3 month maximum.
        innerscan, err := hpClient.GetInnerscanData(runOption.from, runOption.to)
        if err != nil {
            logger.Error("Failed to get Innerscan data", "error", err)
            return
        }
        logger.Info("Successfully retrieved Innerscan data", "data_count", len(innerscan.Data))
        if runOption.output != "" {
            // Write to output file
            file, err := os.Create(runOption.output)
            if err != nil {
                logger.Error("Failed to create output file", "error", err)
                return
            }
            defer file.Close()
            _, err = file.WriteString("Body\n")
            if err != nil {
                logger.Error("Failed to write header to output file", "error", err)
                return
            }
            _, err = file.WriteString(innerscan.ToCsv())
            if err != nil {
                logger.Error("Failed to write to output file", "error", err)
                return
            }
            logger.Info("Data written to output file", "output_file", runOption.output)
        } else {
            // Print to stdout
            fmt.Println("Body")
            fmt.Print(innerscan.ToCsv())
        }
    }
}
