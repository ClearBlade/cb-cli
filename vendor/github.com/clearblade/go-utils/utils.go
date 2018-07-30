package utils

import "fmt"

func ConvertFloatFieldToInt(theMap map[string]interface{}, key string) (int, error) {
	var fieldValue interface{}
	var code int
	var ok bool
	if fieldValue, ok = theMap[key]; !ok {
		return 0, fmt.Errorf("Supplied map does not contain key '%s'", key)
	}
	if val, ok := fieldValue.(float64); !ok {
		return 0, fmt.Errorf("Expected a float for field '%s' but got %T", key, theMap)
	} else {
		code = int(val)
	}
	return code, nil
}
