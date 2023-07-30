package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// permitted strings for BOOL and NULL data types
var falseString = []string{"0", "f", "F", "FALSE", "false", "False"}
var trueString = []string{"1", "t", "T", "TRUE", "true", "True"}

// assert interface as a string, trim whitespaces then check for match in list
// returns 'true' if value was found
func trimmedInterfaceContains(list []string, i interface{}) bool {
	s := strings.Trim(i.(string), " ")
	for _, a := range list {
		if a == s {
			return true
		}
	}
	return false
}

// assert interface as appropriate type, then apply formatting rules
// returns formatted value (string, float64, bool, map, list) and success (bool)
func cleanInterface(array map[string]interface{}) (any, bool) {
	for key, element := range array {
		switch k := strings.Trim(key, " "); k {
		case "S":
			trimmedString := strings.Trim(element.(string), " ")
			if len(trimmedString) > 0 {
				//determine if string is datetime
				ut, err := time.Parse("2006-01-02T15:04:05Z07:00", trimmedString)
				if err == nil {
					return ut.Unix(), true
				}
				return trimmedString, true
			}
		case "N":
			//assert interface as string then convert to floating point number
			str := element.(string)
			if s, err := strconv.ParseFloat(str, 32); err == nil {
				return s, true
			}
		case "M":
			// create new map, assert element as map then iterate over each
			// entry, adding it to the new map if valid
			cleaned := make(map[string]any)

			se, _ := element.(map[string]interface{})
			for key, sub := range se {
				if reflect.TypeOf(sub).Kind() == reflect.Map && len(key) > 0 {
					if str, ok := sub.(map[string]interface{}); ok {
						c, success := cleanInterface(str)
						if success {
							cleaned[key] = c
						}
					}
				}
			}
			if len(cleaned) > 0 {
				return cleaned, true
			}
		case "BOOL":
			// check for TRUE/FALSE values
			if trimmedInterfaceContains(falseString, element) {
				return false, true
			}
			if trimmedInterfaceContains(trueString, element) {
				return true, true
			}
		case "NULL":
			// check for TRUE values only, invalid and FALSE dropped
			if trimmedInterfaceContains(trueString, element) {
				return nil, true
			}

		case "L":
			if (reflect.TypeOf(element).Kind()) == reflect.Slice {
				// create a new list, assert element as list of any types
				// then add valid entries to new list
				anyList := make([]any, 0)
				elAsAny := element.([]any)
				for ele := range elAsAny {
					if (reflect.TypeOf(elAsAny[ele]).Kind()) == reflect.Map {
						jo := elAsAny[ele].(map[string]interface{})
						c, success := cleanInterface(jo)
						if success {
							anyList = append(anyList, c)
						}

					}
				}
				if len(anyList) > 0 {
					return anyList, true
				}
			}
		}
	}
	// entry is invalid
	return false, false
}

// load a json file, apply data cleaning rules then output to STDOUT
func main() {
	start := time.Now()
	// Open input json file
	jsonFile, err := os.Open("input.json")
	// if error opening file, print it
	if err != nil {
		fmt.Println(err)
	}
	byteValue, _ := io.ReadAll(jsonFile)
	var result map[string]interface{}
	//create an empty map to hold cleaned values
	cleaned := make(map[string]any)
	if err := json.Unmarshal(byteValue, &result); err != nil {
		fmt.Println(err)
	}
	//iterate over entries in json file and format, if valid add to new map
	for key, element := range result {
		if reflect.TypeOf(element).Kind() == reflect.Map && len(key) > 0 {
			if str, ok := element.(map[string]interface{}); ok {
				c, success := cleanInterface(str)
				if success {
					cleaned[key] = c
				}
			}
		}
	}
	var output []any
	output = append(output, cleaned)
	// format cleaned map as json and output to STDOUT
	b, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(b) + "\n")
	fmt.Printf("duration: %v\n", time.Since(start))
}
