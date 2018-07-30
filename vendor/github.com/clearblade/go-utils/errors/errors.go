package errors

import (
	"encoding/json"
	"fmt"

	utils "github.com/clearblade/go-utils"
	cc "github.com/clearblade/go-utils/errors/constants/category"
	uuid "github.com/clearblade/go-utils/uuid"
)

type Info struct {
	Id            uuid.Uuid           `json:"id"`
	Code          int                 `json:"code"`
	Level         int                 `json:"level"`
	Category      cc.CategoryConstant `json:"category"`
	Message       string              `json:"message"`
	Detail        string              `json:"detail"`
	LowLevelError error               `json:"lowLevelError"`
	Line          string              `json:"line"`
}

type Response struct {
	Info       Info `json:"error" mapstructure:"error"`
	StatusCode int  `json:"statusCode"`
}

func (e *Response) Error() string {
	marsha, _ := json.Marshal(e)
	return string(marsha)
}

func (e *Response) HTTPStatusCode() int {
	return e.StatusCode
}

// this func is used to convert a JSON decoded HTTP response into the Response object
// one tricky thing to note is that the JSON decoder assumes that all numbers are float64, but
// the Response object only uses ints for code, level, statusCode, etc.
// because of this we need to do some conversion between the two types
func (e *Response) fromMap(body interface{}) *Response {
	var imTheMap map[string]interface{}
	var ok bool
	if imTheMap, ok = body.(map[string]interface{}); !ok {
		return createEmptyErrorResponse(body)
	}

	var errorInfo map[string]interface{}

	if errorInfo, ok = imTheMap["error"].(map[string]interface{}); !ok {
		return createEmptyErrorResponse(body)
	}
	statusCode, err := utils.ConvertFloatFieldToInt(imTheMap, "statusCode")
	if err != nil {
		return createEmptyErrorResponse(body)
	}

	// alright, it must be the format we're expecting. let's grab the fields we need

	parsedId, err := uuid.ParseUuid(errorInfo["id"].(string))
	if err != nil {
		return createEmptyErrorResponse(body)
	}
	errorCode, err := utils.ConvertFloatFieldToInt(errorInfo, "code")
	if err != nil {
		return createEmptyErrorResponse(body)
	}

	errorLevel, err := utils.ConvertFloatFieldToInt(errorInfo, "level")
	if err != nil {
		return createEmptyErrorResponse(body)
	}

	info := Info{
		Id:            parsedId,
		Message:       errorInfo["message"].(string),
		Code:          errorCode,
		Level:         errorLevel,
		Category:      cc.CategoryConstant(errorInfo["category"].(string)),
		Detail:        errorInfo["detail"].(string),
		LowLevelError: fmt.Errorf("%+v", errorInfo["lowLevelError"]),
		Line:          errorInfo["line"].(string),
	}

	return &Response{
		Info:       info,
		StatusCode: statusCode,
	}
}

func createEmptyErrorResponse(body interface{}) *Response {
	info := Info{
		Message: fmt.Sprintf("%+v", body),
	}
	return &Response{
		Info:       info,
		StatusCode: 0,
	}
}

func CreateResponseFromMap(body interface{}) *Response {
	var errResp Response
	return errResp.fromMap(body)
}
