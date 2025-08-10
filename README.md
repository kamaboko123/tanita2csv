# tanita2csv
Convert Tanita Health planet body data to garmin importable CSV

## Build
```bash
# build tanita2csv binary
make all

# clean
make clean
```

## Configuration
Create `config.yml` file:
```yaml
url: "https://www.healthplanet.jp"
client_id: "your_client_id"
client_secret: "your_client_secret"
token_file: "token.json"
timezone: "Asia/Tokyo"
```

## Usage

### Authentication
First, authenticate with HealthPlanet API:
```bash
./bin/tanita2csv -m auth
```

### Export data to CSV
Export body measurement data to CSV format:
```bash
# Export to stdout (default: last 90 days)
./bin/tanita2csv -m dump

# Export to file with custom date range
./bin/tanita2csv -m dump -f 2024-01-01 -t 2024-03-31 -o output.csv

# Debug mode
./bin/tanita2csv -m dump -v
```

### Options
- `-m`: Mode (`auth` or `dump`)
- `-f`: From date (YYYY-MM-DD, default: 90 days ago)
- `-t`: To date (YYYY-MM-DD, default: today)
- `-o`: Output file path (default: stdout)
- `-v`: Debug mode (verbose logging)

### CSV Format
The output CSV contains:
- Date: Measurement date (YYYY-MM-DD)
- Weight: Body weight (kg)
- BMI: Body Mass Index (calculated)
- Fat: Body fat percentage (%)

