package cbjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func GetJSONFile(filename string) (map[string]interface{}, string, error) {
	jsonStr, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, "", err
	}
	return GetJSONFromString(string(jsonStr))
}

func GetJSONFromString(jsonStr string) (map[string]interface{}, string, error) {
	p := NewPreprocessor([]byte(jsonStr))
	jsonBuf, err := p.preprocess()
	if err != nil {
		return nil, "", err
	}
	parsed := map[string]interface{}{}
	jsonBytes := jsonBuf.Bytes()
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		return nil, "", fmt.Errorf("%s at line %d\n", err.Error(), errorLineNumber(jsonBytes, err))
	}
	return parsed, jsonStr, nil
}

func errorLineNumber(jsonStr []byte, err error) int {
	if jsonErr, ok := err.(*json.SyntaxError); ok {
		return bytes.Count(jsonStr[:jsonErr.Offset], []byte("\n")) + 1
	}
	return -1
}
