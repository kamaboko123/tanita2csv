package healthplanet

import (
	"fmt"
)

type InnerscanResponse struct {
    BirthDate string `json:"birth_date"`
    Height string `json:"height"`
    Sex string `json:"sex"`
    Data []struct{
        Date string `json:"date"`
        KeyData string `json:"keydata"`
        Tag string `json:"tag"`
        Model string `json:"model"`
    } `json:"data"`
}

func (d *InnerscanData) String() string {
    return fmt.Sprintf("(%s)Weight: %f, BodyFat: %f", d.Date, d.Weight, d.BodyFat)
}
