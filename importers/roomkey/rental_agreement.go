package roomkey

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"reflect"
	"rentroll/importers/core"
	"rentroll/rlib"
	"strings"
	"time"
)

// CreateRentalAgreementCSV create rental agreement csv temporarily
// write headers, used to load data from onesite csv
// return file pointer to call program
func CreateRentalAgreementCSV(
	CSVStore string,
	timestamp string,
	rentalAgreementStruct *core.RentalAgreementCSV,
) (*os.File, *csv.Writer, bool) {

	var done = false

	// get path of rentalAgreement csv file
	filePrefix := prefixCSVFile["rental_agreement"]
	fileName := filePrefix + timestamp + ".csv"
	rentalAgreementCSVFilePath := path.Join(CSVStore, fileName)

	// try to create file and return with error if occurs any
	rentalAgreementCSVFile, err := os.Create(rentalAgreementCSVFilePath)
	if err != nil {
		rlib.Ulog("Error <RENTAL AGREEMENT CSV>: %s\n", err.Error())
		return nil, nil, done
	}

	// create csv writer
	rentalAgreementCSVWriter := csv.NewWriter(rentalAgreementCSVFile)

	// parse headers of rentalAgreementCSV using reflect
	rentalAgreementCSVHeaders := []string{}
	rentalAgreementCSVHeaders, ok := core.GetStructFields(rentalAgreementStruct)
	if !ok {
		rlib.Ulog("Error <RENTAL AGREEMENT CSV>: Unable to get struct fields for rentalAgreementCSV\n")
		return nil, nil, done
	}

	rentalAgreementCSVWriter.Write(rentalAgreementCSVHeaders)
	rentalAgreementCSVWriter.Flush()

	done = true

	return rentalAgreementCSVFile, rentalAgreementCSVWriter, done
}

// WriteRentalAgreementData used to write the data to csv file
// with avoiding duplicate data
func WriteRentalAgreementData(
	recordCount *int,
	rowIndex int,
	traceCSVData map[int]int,
	csvWriter *csv.Writer,
	csvRow *CSVRow,
	currentTime time.Time,
	currentTimeFormat string,
	suppliedValues map[string]string,
	rentalAgreementStruct *core.RentalAgreementCSV,
	traceTCIDMap map[int]string,
	csvErrors map[int][]string,
) {
	// TODO: need to decide how to avoid data
	// checkRentableStyle := csvRow.FloorPlan
	// Stylefound := core.StringInSlice(checkRentableStyle, *avoidData)

	// // if style found then simplay return otherwise continue
	// if Stylefound {
	//  return
	// }

	// *avoidData = append(*avoidData, checkRentableStyle)

	// TODO: generate error here for RentableTypeRef
	// to let endusers know that least start/end dates don't exists so we are taking
	// defaults
	currentYear, currentMonth, currentDate := currentTime.Date()
	DtStart := fmt.Sprintf("%d/%d/%d", currentMonth, currentDate, currentYear)
	// DtStart := fmt.Sprintf("%02d/%02d/%04d", currentMonth, currentDate, currentYear)
	DtStop := "12/31/9999" // no end date

	// make rentable data from userSuppliedValues and defaultValues
	rentableDefaultData := map[string]string{}
	for k, v := range suppliedValues {
		rentableDefaultData[k] = v
	}

	// flag warning that we are taking default values for least start, end dates
	// as they don't exists
	if csvRow.Empty3 == "" {
		warnPrefix := "W:<" + core.DBTypeMapStrings[core.DBRentable] + ">:"
		csvErrors[rowIndex] = append(csvErrors[rowIndex],
			warnPrefix+"No lease start date found. Using default value: "+DtStart,
		)
	}
	if csvRow.DateOut == "" {
		warnPrefix := "W:<" + core.DBTypeMapStrings[core.DBRentable] + ">:"
		csvErrors[rowIndex] = append(csvErrors[rowIndex],
			warnPrefix+"No lease start date found. Using default value: "+DtStop,
		)
	}

	dateIn := getFormattedDate(csvRow.Empty3)
	dateOut := getFormattedDate(csvRow.DateOut)

	rentableDefaultData["DtStart"] = dateIn
	rentableDefaultData["DtStop"] = dateOut
	rentableDefaultData["TCID"] = traceTCIDMap[rowIndex]

	// get csv row data
	ok, csvRowData := GetRentalAgreementCSVRow(
		csvRow, rentalAgreementStruct,
		currentTimeFormat, rentableDefaultData,
	)

	if ok {
		csvWriter.Write(csvRowData)
		csvWriter.Flush()

		// after write operation to csv,
		// entry this rowindex with unit value in the map
		*recordCount = *recordCount + 1
		traceCSVData[*recordCount] = rowIndex
	}
}

// GetRentalAgreementCSVRow used to create RentalAgreement
// csv row from roomkey csv
func GetRentalAgreementCSVRow(
	roomKeyRow *CSVRow,
	fieldMap *core.RentalAgreementCSV,
	timestamp string,
	DefaultValues map[string]string,
) (bool, []string) {

	// take initial variable
	ok := false

	// ======================================
	// Load rentalAgreement's data from onesiterow data
	// ======================================
	reflectedOneSiteRow := reflect.ValueOf(roomKeyRow).Elem()
	reflectedRentalAgreementFieldMap := reflect.ValueOf(fieldMap).Elem()

	// length of RentalAgreementCSV
	rRTLength := reflectedRentalAgreementFieldMap.NumField()

	// return data array
	dataMap := make(map[int]string)

	for i := 0; i < rRTLength; i++ {
		// get rentalAgreement field
		rentalAgreementField := reflectedRentalAgreementFieldMap.Type().Field(i)

		// if rentalAgreementField value exist in DefaultValues map
		// then set it first
		suppliedValue, found := DefaultValues[rentalAgreementField.Name]
		if found {
			dataMap[i] = suppliedValue
		}

		// =========================================================
		// this condition has been put here because it's mapping field does not exist
		// =========================================================
		if rentalAgreementField.Name == "PayorSpec" {
			payorSpec, ok := getPayorSpec(roomKeyRow, DefaultValues)
			if ok {
				dataMap[i] = payorSpec
			} else {
				// TODO: verify that what to do in false case
				dataMap[i] = payorSpec
			}
		}
		if rentalAgreementField.Name == "UserSpec" {
			userSpec, ok := getUserSpec(roomKeyRow, DefaultValues)
			if ok {
				dataMap[i] = userSpec
			} else {
				// TODO: verify that what to do in false case
				dataMap[i] = userSpec
			}
		}
		if rentalAgreementField.Name == "RentableSpec" {
			rentableSpec, ok := getRentableSpec(roomKeyRow)
			if ok {
				dataMap[i] = rentableSpec
			} else {
				// TODO: verify that what to do in false case
				dataMap[i] = rentableSpec
			}
		}

		// get mapping field
		MappedFieldName := reflectedRentalAgreementFieldMap.FieldByName(rentalAgreementField.Name).Interface().(string)

		// if has not value then continue
		if !reflectedOneSiteRow.FieldByName(MappedFieldName).IsValid() {
			continue
		}

		// get field by mapping field name and then value
		OneSiteFieldValue := reflectedOneSiteRow.FieldByName(MappedFieldName).Interface()
		dataMap[i] = OneSiteFieldValue.(string)

		// Formatting dates to RentRoll importable format
		if rentalAgreementField.Name == "AgreementStart" {
			dataMap[i] = getFormattedDate(roomKeyRow.DateRes)
		}
		if rentalAgreementField.Name == "PossessionStart" ||
			rentalAgreementField.Name == "RentStart" {
			dataMap[i] = getFormattedDate(roomKeyRow.Empty3)
		}
		if rentalAgreementField.Name == "AgreementStop" ||
			rentalAgreementField.Name == "PossessionStop" ||
			rentalAgreementField.Name == "RentStop" {
			dataMap[i] = getFormattedDate(roomKeyRow.DateOut)
		}

	}

	dataArray := []string{}

	for i := 0; i < rRTLength; i++ {
		dataArray = append(dataArray, dataMap[i])
	}
	ok = true
	return ok, dataArray
}

// getPayorSpec used to get payor spec in format of rentroll system
func getPayorSpec(
	csvRow *CSVRow,
	defaults map[string]string,
) (string, bool) {

	// TODO: verify if validation required here
	ok := false

	orderedFields := []string{}

	// append TCID for user identification
	orderedFields = append(orderedFields, defaults["TCID"])

	if defaults["TCID"] != "" {
		// append rent start
		if csvRow.Empty3 == "" {
			orderedFields = append(orderedFields, defaults["DtStart"])
		} else {
			orderedFields = append(orderedFields, csvRow.Empty3)
		}

		// append date out
		if csvRow.DateOut == "" {
			orderedFields = append(orderedFields, defaults["DtStop"])
		} else {
			orderedFields = append(orderedFields, csvRow.DateOut)
		}
	} else {
		fmt.Println(csvRow)
		fmt.Println(defaults)
	}

	ok = true
	if ok {
		return strings.Join(orderedFields, ","), ok
	}

	return ",,", ok
}

// getUserSpec used to get user spec in format of rentroll system
func getUserSpec(
	csvRow *CSVRow,
	defaults map[string]string,
) (string, bool) {

	// TODO: verify if validation required here
	ok := false

	orderedFields := []string{}

	orderedFields = append(orderedFields, defaults["TCID"])

	if defaults["TCID"] != "" {
		// append rent start
		if csvRow.Empty3 == "" {
			orderedFields = append(orderedFields, defaults["DtStart"])
		} else {
			orderedFields = append(orderedFields, csvRow.Empty3)
		}

		// append date out
		if csvRow.DateOut == "" {
			orderedFields = append(orderedFields, defaults["DtStop"])
		} else {
			orderedFields = append(orderedFields, csvRow.DateOut)
		}
	}

	ok = true
	if ok {
		return strings.Join(orderedFields, ","), ok
	}

	return ",,", ok
}

// getRentableSpec used to get rentable spec in format of rentroll system
func getRentableSpec(
	csvRow *CSVRow,
) (string, bool) {

	// TODO: verify if validation required here
	ok := false

	orderedFields := []string{}

	// append rentable
	orderedFields = append(orderedFields, csvRow.Room)
	// append contractrent
	rent := csvRow.Rate
	rent = strings.Replace(rent, "$", "", -1)
	rent = strings.Replace(rent, ".", "", -1)
	orderedFields = append(orderedFields, rent)

	ok = true
	if ok {
		return strings.Join(orderedFields, ","), ok
	}

	return ",", ok
}
