package orm

import (
	"reflect"
)

// MakeArrFromStruct struct -> []interface
func MakeArrFromStruct(data interface{}) []interface{} {
	arr := []interface{}{}
	v := reflect.ValueOf(data)

	for i := 0; i < v.NumField(); i++ {
		arr = append(arr, v.Field(i).Interface())
	}
	return arr
}

// MapFromStructAndMatrix [][]interface{}{} + Struct sample -> []map[string]interface{}{}
func MapFromStructAndMatrix(data [][]interface{}, sampleStruct interface{}, additionalFields ...string) []map[string]interface{} {
	structLen := reflect.ValueOf(sampleStruct).NumField()
	t := reflect.TypeOf(sampleStruct)
	result := []map[string]interface{}{}

	for _, currentRow := range data {
		oneRow := map[string]interface{}{}
		for i := 0; i < structLen; i++ {
			jsonName := t.Field(i).Tag.Get("json")
			if jsonName == "" {
				continue
			}
			oneRow[jsonName] = currentRow[i]
		}

		// add additional fields for result
		for i, v := range additionalFields {
			oneRow[v] = currentRow[i+structLen]
		}
		result = append(result, oneRow)
	}
	return result
}

// FromINT64ToINT convert int64 -> int
func FromINT64ToINT(number interface{}) int {
	return int(number.(int64))
}
