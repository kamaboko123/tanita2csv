package healthplanet

import (
	"fmt"
    "time"
    "strconv"
    "sort"
    "math"
)

// API response structures
type InnerscanDataResponse struct {
    Date string `json:"date"`
    KeyData string `json:"keydata"`
    Tag string `json:"tag"`
    Model string `json:"model"`
}

type InnerscanResponse struct {
    BirthDate string `json:"birth_date"`
    Height string `json:"height"`
    Sex string `json:"sex"`
    Data []InnerscanDataResponse `json:"data"`
}

type InnerscanData struct {
    Date time.Time
    Weight float64
    BodyFat float64
    BMI float64     // Calc from Weight and Height(it is defined in `Innerscan` struct)
}

func (d *InnerscanData) Validate() error {
    if d.Date.IsZero() {
        return fmt.Errorf("date is not set")
    }
    if d.Weight <= 0 {
        return fmt.Errorf("weight must be greater than 0")
    }
    if d.BodyFat < 0 || d.BodyFat > 100 {
        return fmt.Errorf("body fat must be between 0 and 100")
    }
    if d.BMI < 0 {
        return fmt.Errorf("BMI must be greater than or equal to 0")
    }
    return nil
}

// Decoded data
type Innerscan struct {
    BirthDate time.Time
    Hight float64
    Sex string
    Data []*InnerscanData
}

func (ir *InnerscanResponse) ToInnerscan(timezone *time.Location) (*Innerscan, error) {
    innerscan := &Innerscan{
        Data: make([]*InnerscanData, 0),
    }

    innerscan.Sex = ir.Sex

    birthDate, err := time.Parse("20060102", ir.BirthDate)
    if err != nil {
        return nil, fmt.Errorf("failed to parse birth date: %w", err)
    }
    innerscan.BirthDate = birthDate
    
    height, err := strconv.ParseFloat(ir.Height, 64)
    if err != nil {
        return nil, fmt.Errorf("failed to parse height: %w", err)
    }
    innerscan.Hight = height

    // convert data with map
    // key: date in "2006-01-02 15:04:05" format
    // Because a mesurement made is made up of multiple data, we need to merge them by date
    dates := make(map[string] *InnerscanData)
    for _, d := range ir.Data {
        date, err := time.Parse("200601021504", d.Date)
        key := date.In(timezone).Format("2006-01-02 15:04")
        if err != nil {
            return nil, fmt.Errorf("failed to parse date: %w", err)
        }
        // if mesurement data is not exist, add it
        if _, ok := dates[key]; !ok {
            dates[key] = &InnerscanData{
                Date: date,
                Weight: 0,
                BodyFat: 0,
                BMI: 0,
            }
        }

        switch d.Tag {
        case "6021": // Weight
            if d.KeyData != "" {
                weight, err := strconv.ParseFloat(d.KeyData, 64)
                if err != nil {
                    return nil, fmt.Errorf("failed to parse weight: %w", err)
                }
                dates[key].Weight = weight
            }
        case "6022": // Body Fat
            if d.KeyData != "" {
                bodyFat, err := strconv.ParseFloat(d.KeyData, 64)
                if err != nil {
                    return nil, fmt.Errorf("failed to parse body fat: %w", err)
                }
                dates[key].BodyFat = bodyFat
            }
        default:
            return nil, fmt.Errorf("unknown tag: %s", d.Tag)
        }

        if dates[key].Weight != 0 && innerscan.Hight != 0 {
            // Calculate BMI if weight and height are available
            dates[key].BMI = dates[key].Weight / math.Pow(innerscan.Hight/100, 2)
        }
    }

    // validate all data and if not valid, delete it
    for _, d := range dates {
        if err := d.Validate(); err != nil {
            delete(dates, d.Date.In(timezone).Format("2006-01-02 15:04"))
        }
    }

    // convert data with uniq by date
    // if there are multiple data for the same date, keep the latest one
    
    // sort keys 
    keys := make([]string, 0, len(dates))
    for k := range dates {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    // uniq by date(overwrite if same date)
    uniqueData := make(map[string]*InnerscanData)
    for _, k := range keys { // k is formatted date "2006-01-02 15:04"
        day := k[:10] // "2006-01-02"
        uniqueData[day] = dates[k]
    }

    for _, k := range keys {
        day := k[:10]
        innerscan.Data = append(innerscan.Data, uniqueData[day])
    }

    return innerscan, nil
}



// String conversion
func (d *InnerscanData) String() string {
    return fmt.Sprintf("(%s)Weight: %f, BMI: %f, BodyFat: %f", d.Date, d.Weight, d.BMI, d.BodyFat)
}

// CSV Conversion
func (i *Innerscan) CsvHeader() string {
    return "Date,Weight,BMI,Fat"
}
func (i *Innerscan) ToCsv() string {
    ret := i.CsvHeader() + "\n"
    for _, d := range i.Data {
        ret += fmt.Sprintf("%s,%f,%f,%f\n", d.Date.Format("2006-01-02"), d.Weight, d.BMI, d.BodyFat)
    }
    return ret
}
