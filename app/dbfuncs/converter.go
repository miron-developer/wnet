/*
	converter convert interface <-> struct
*/

package dbfuncs

import (
	"reflect"
)

// struct -> []interface
func makeArrFromStruct(data interface{}) []interface{} {
	arr := []interface{}{}
	v := reflect.ValueOf(data)

	for i := 0; i < v.NumField(); i++ {
		arr = append(arr, v.Field(i).Interface())
	}
	return arr
}

// MapFromStructAndMatrix [][]interface{}{} + Struct sample -> []map[string]interface{}{}
func MapFromStructAndMatrix(data [][]interface{}, sampleStruct interface{}, joinArgs ...interface{}) []map[string]interface{} {
	structLen := reflect.ValueOf(sampleStruct).NumField()
	t := reflect.TypeOf(sampleStruct)
	result := []map[string]interface{}{}

	for _, currentRow := range data {
		// if len(currentRow) != structLen {
		// 	continue
		// }
		oneRow := map[string]interface{}{}
		for i := 0; i < structLen; i++ {
			jsonName := t.Field(i).Tag.Get("json")
			if jsonName == "" {
				continue
			}
			oneRow[jsonName] = currentRow[i]
		}
		for i, v := range joinArgs {
			oneRow[v.(string)] = currentRow[i+structLen]
		}
		result = append(result, oneRow)
	}
	return result
}

// FromINT64ToINT convert int64 -> int
func FromINT64ToINT(number interface{}) int {
	return int(number.(int64))
}
