package onesite

import (
	"reflect"
	"rentroll/importers/core"
	"rentroll/rcsv"
)

// OneSiteMap is struct which contains several categories
// used to store the data from onesite to rentroll system
type OneSiteMap struct {
	RentableTypeCSV core.RentableTypeCSV
}

// OneSiteCSVRow contains fields which represents value
// exactly to the each raw of onesite input csv file
type OneSiteCSVRow struct {
	Unit            string
	FloorPlan       string
	UnitDesignation string
	Sqft            string
	UnitLeaseStatus string
	Name            string
	PhoneNumber     string
	Email           string
	MoveIn          string
	NoticeForDate   string
	MoveOut         string
	LeaseStart      string
	LeaseEnd        string
	MarketAddl      string
	DepOnHand       string
	Balance         string
	TotalCharges    string
	Rent            string
	WaterReImb      string
	Corp            string
	Discount        string
	Platinum        string
	Tax             string
	ElectricReImb   string
	Fire            string
	ConcSpecl       string
	WashDry         string
	EmplCred        string
	Short           string
	PetFee          string
	TrashReImb      string
	TermFee         string
	LakeView        string
	Utility         string
	Furn            string
	Mtom            string
	Referral        string
}

// LoadOneSiteCSVRow used to load data from slice
// into OneSiteCSVRow struct and return that struct
func LoadOneSiteCSVRow(csvCols []rcsv.CSVColumn, data []string) (bool, OneSiteCSVRow) {
	csvRow := reflect.New(reflect.TypeOf(OneSiteCSVRow{}))
	rowLoaded := false

	// fill data according to headers length
	for i := 0; i < len(csvCols); i++ {
		csvRow.Elem().Field(i).Set(reflect.ValueOf(data[i]))
	}

	// if blank data has not been passed then only need to return true
	if (OneSiteCSVRow{}) != csvRow.Elem().Interface().(OneSiteCSVRow) {
		rowLoaded = true
	}
	return rowLoaded, csvRow.Elem().Interface().(OneSiteCSVRow)
}
