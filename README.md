# tanita2csv
Convert Tanita Health planet body data to garmin importable CSV

## Build
To build the `tanita2csv` binary, you need to have Go installed on your system. Follow these steps:
1. Clone the repository:
```bash
git clone https://github.com/kamaboko123/tanita2csv.git
cd tanita2csv
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the binary:
```bash
make
```

4. The binary will be created in the `bin` directory:
```bash
ls bin/tanita2csv
```

(Optional) Clean up build artifacts:
```bash
make clean
```

## Configuration
Create `config.yml` file:
```bash
cp config.example.yml config.yml
```

Edit `config.yml` with your HealthPlanet API credentials:
```yaml
url: "https://www.healthplanet.jp"
client_id: "your_client_id"
client_secret: "your_client_secret"
token_file: "token.json"
```

If you don't have client_id and client_secret, you need to register your application on HealthPlanet API.

### Register the application for HealthPlanet API
1. Go to [HealthPlanet API registration page](https://www.healthplanet.jp/apis_account.do)
2. Register a new application
3. Fill in the fields:
    - **Service Name**: Your app name(e.g., "tanita2csv")
    - **Email Address**: Your email address
    - **Description**: Brief description of your app(e.g., "Tanita body data exporter")
    - **Application Type**: Client Application


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


### Reauthentication
If you need to reauthenticate when the token does not work by any reason, you can remove the `token.json` file and run the authentication command again:
```bash
rm token.json
./bin/tanita2csv -m auth
```
