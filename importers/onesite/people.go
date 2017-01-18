package onesite

import (
	"encoding/csv"
	"os"
	"path"
	"reflect"
	"rentroll/importers/core"
	"rentroll/rlib"
	"strconv"
	"strings"
)

// CreatePeopleCSV create people csv temporarily
// write headers, used to load data from onesite csv
// return file pointer to call program
func CreatePeopleCSV(
	CSVStore string,
	timestamp string,
	peopleCSVStruct *core.PeopleCSV,
) (*os.File, *csv.Writer, bool) {

	var done = false

	// get path of people csv file
	filePrefix := prefixCSVFile["people"]
	fileName := filePrefix + timestamp + ".csv"
	peopleCSVFilePath := path.Join(CSVStore, fileName)

	// try to create file and return with error if occurs any
	peopleCSVFile, err := os.Create(peopleCSVFilePath)
	if err != nil {
		rlib.Ulog("Error <PEOPLE CSV>: %s\n", err.Error())
		return nil, nil, done
	}

	// create csv writer
	peopleCSVWriter := csv.NewWriter(peopleCSVFile)

	// parse headers of peopleCSV using reflect
	peopleCSVHeaders := []string{}
	peopleCSVHeaders, ok := core.GetStructFields(peopleCSVStruct)
	if !ok {
		rlib.Ulog("Error <PEOPLE CSV>: Unable to get struct fields for peopleCSV\n")
		return nil, nil, done
	}

	peopleCSVWriter.Write(peopleCSVHeaders)
	peopleCSVWriter.Flush()

	done = true

	return peopleCSVFile, peopleCSVWriter, done
}

// WritePeopleCSVData used to write the data to csv file
// with avoiding duplicate data
func WritePeopleCSVData(
	recordCount *int,
	rowIndex int,
	traceCSVData map[int]int,
	csvWriter *csv.Writer,
	csvRow *CSVRow,
	// avoidData *[]string,
	currentTimeFormat string,
	suppliedValues map[string]string,
	peopleStruct *core.PeopleCSV,
) {
	// TODO: need to decide how to avoid duplicate data

	// get csv row data
	ok, csvRowData := GetPeopleCSVRow(
		csvRow, peopleStruct,
		currentTimeFormat, suppliedValues,
		rowIndex,
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

// GetPeopleCSVRow used to create people
// csv row from onesite csv data
func GetPeopleCSVRow(
	oneSiteRow *CSVRow,
	fieldMap *core.PeopleCSV,
	timestamp string,
	DefaultValues map[string]string,
	rowIndex int,
) (bool, []string) {

	// take initial variable
	ok := false

	// ======================================
	// Load people's data from onesiterow data
	// ======================================
	reflectedOneSiteRow := reflect.ValueOf(oneSiteRow).Elem()
	reflectedPeopleFieldMap := reflect.ValueOf(fieldMap).Elem()

	// length of PeopleCSV
	pplLength := reflectedPeopleFieldMap.NumField()

	// return data array
	dataMap := make(map[int]string)

	for i := 0; i < pplLength; i++ {
		// get people field
		peopleField := reflectedPeopleFieldMap.Type().Field(i)

		// if peopleField value exist in DefaultValues map
		// then set it first
		suppliedValue, found := DefaultValues[peopleField.Name]
		if found {
			dataMap[i] = suppliedValue
		}

		// =========================================================
		// this condition has been put here because it's mapping field does not exist
		// =========================================================
		if peopleField.Name == "LastName" {
			nameSlice := strings.Split(oneSiteRow.Name, ",")
			dataMap[i] = strings.TrimSpace(nameSlice[0])
		}
		if peopleField.Name == "FirstName" {
			nameSlice := strings.Split(oneSiteRow.Name, ",")
			if len(nameSlice) > 1 {
				dataMap[i] = strings.TrimSpace(nameSlice[1])
			} else {
				dataMap[i] = ""
			}
		}
		// Special notes for people to get TCID in future with below value
		if peopleField.Name == "Notes" {
			dataMap[i] = onesiteNotesPrefix + strconv.Itoa(rowIndex)
		}

		// get mapping field
		MappedFieldName := reflectedPeopleFieldMap.FieldByName(peopleField.Name).Interface().(string)

		// if has not value then continue
		if !reflectedOneSiteRow.FieldByName(MappedFieldName).IsValid() {
			continue
		}

		// get field by mapping field name and then value
		OneSiteFieldValue := reflectedOneSiteRow.FieldByName(MappedFieldName).Interface()
		dataMap[i] = OneSiteFieldValue.(string)
	}

	dataArray := []string{}

	for i := 0; i < pplLength; i++ {
		dataArray = append(dataArray, dataMap[i])
	}
	ok = true
	return ok, dataArray
}
