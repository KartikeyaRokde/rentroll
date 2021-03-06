package onesite

import (
	"encoding/csv"
	"os"
	"path"
	"reflect"
	"rentroll/importers/core"
	"rentroll/rlib"
)

// =========
// value type
// =========
// 0 - string, a collection of characters
// 1 - 64-bit integer
// 2 - 64-bit unsigned integer
// 3 - 64-bit floating point
// 4 - Date

// customAttributeMap holds the fields which needs to be extracted from onesite csv
// and for each field, need to create rows with multiple values.
// Key of this map should match exactly the column of onesite csv's custom attribute
// so this program can parse the value from this key field.
var customAttributeMap = map[string]map[string]string{
	"SQFT": {"Name": "Square Feet", "ValueType": "1", "Units": "sqft"},
}

// CreateCustomAttibutesCSV create rentabletype csv temporarily
// write headers, used to load data from onesite csv
// return file pointer to call program
func CreateCustomAttibutesCSV(
	CSVStore string,
	timestamp string,
	customAttributeStruct *core.CustomAttributeCSV,
) (*os.File, *csv.Writer, bool) {

	var done = false

	// get path of custom attribute csv file
	filePrefix := prefixCSVFile["custom_attribute"]
	fileName := filePrefix + timestamp + ".csv"
	customAttributeCSVFilePath := path.Join(CSVStore, fileName)

	// try to create file and return with error if occurs any
	customAttributeCSVFile, err := os.Create(customAttributeCSVFilePath)
	if err != nil {
		rlib.Ulog("Error <CUSTOM ATTRIBUTES CSV>: %s\n", err.Error())
		return nil, nil, done
	}

	// create csv writer
	customAttributeCSVWriter := csv.NewWriter(customAttributeCSVFile)

	// parse headers of customAttributeCSV using reflect
	customAttributeCSVHeaders := []string{}
	customAttributeCSVHeaders, ok := core.GetStructFields(customAttributeStruct)
	if !ok {
		rlib.Ulog("Error <CUSTOM ATTRIBUTES CSV>: Unable to get struct fields for customAttributeCSV\n")
		return nil, nil, done
	}

	customAttributeCSVWriter.Write(customAttributeCSVHeaders)
	customAttributeCSVWriter.Flush()

	done = true

	return customAttributeCSVFile, customAttributeCSVWriter, done
}

// WriteCustomAttributeData used to write the data to csv file
// with avoiding duplicate data
func WriteCustomAttributeData(
	recordCount *int,
	rowIndex int,
	traceCSVData map[int]int,
	csvWriter *csv.Writer,
	csvRow *CSVRow,
	avoidData map[string][]string,
	currentTimeFormat string,
	suppliedValues map[string]string,
	customAttributeStruct *core.CustomAttributeCSV,
) {

	for customAttributeField, customAttributeConfig := range customAttributeMap {

		reflectedOneSiteRow := reflect.ValueOf(csvRow).Elem()

		// get the value for key field from onesite row
		value := reflectedOneSiteRow.FieldByName(customAttributeField).Interface().(string)

		ValueFound := core.StringInSlice(value, avoidData[customAttributeField])
		// if value found then simplay continue to next
		if ValueFound {
			continue
		}
		avoidData[customAttributeField] = append(avoidData[customAttributeField], value)

		// csv row csvRowData used to write data it holds
		csvRowData := []string{}
		csvRowData = append(csvRowData, suppliedValues["BUD"])
		csvRowData = append(csvRowData, customAttributeConfig["Name"])
		csvRowData = append(csvRowData, customAttributeConfig["ValueType"])
		csvRowData = append(csvRowData, value)
		csvRowData = append(csvRowData, customAttributeConfig["Units"])

		csvWriter.Write(csvRowData)
		csvWriter.Flush()

		// after write operation to csv,
		// entry this rowindex with unit value in the map
		*recordCount = *recordCount + 1

		// need to map on next row index of temp csv as first row is header line
		// and recordCount initialized with 0 value
		traceCSVData[*recordCount+1] = rowIndex
	}
}
