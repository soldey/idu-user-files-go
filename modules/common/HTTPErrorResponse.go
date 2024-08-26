package common

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ErrFields map[string]string // Error field-value pair type

type ResponseError struct {
	Msg    string    `json:"message"` // Error message
	Status int       `json:"status"`  // Http status code
	Data   ErrFields // For extra error fields e.g. reason, details, etc.
}

type ErrList []ResponseError // Multiple http errors type

// AddErrField adds a new field to the response error with given key and value
func (err *ResponseError) AddErrField(key, value string) {
	if err.Data == nil {
		err.Data = make(ErrFields)
	}
	err.Data[key] = value
}

// RemoveErrField removes existing field matching given key from response error
func (err *ResponseError) RemoveErrField(key string) {
	delete(err.Data, key)
}

// MarshalJSON marshals the response error into json
func (err *ResponseError) MarshalJSON() ([]byte, error) {
	// Determine json field name for error message
	errType := reflect.TypeOf(*err)
	msgField, ok := errType.FieldByName("Msg")
	msgJsonName := "message"
	if ok {
		msgJsonTag := msgField.Tag.Get("json")
		if msgJsonTag != "" {
			msgJsonName = msgJsonTag
		}
	}
	// Determine json field name for error status code
	statusField, ok := errType.FieldByName("Status")
	statusJsonName := "status"
	if ok {
		statusJsonTag := statusField.Tag.Get("json")
		if statusJsonTag != "" {
			statusJsonName = statusJsonTag
		}
	}
	fieldMap := make(map[string]string)
	fieldMap[msgJsonName] = err.Msg
	fieldMap[statusJsonName] = fmt.Sprintf("%d", err.Status)
	for key, value := range err.Data {
		fieldMap[key] = value
	}
	return json.Marshal(fieldMap)
}

// SerializeJSON converts response error into serialized json string
func (resErr *ResponseError) SerializeJSON() (string, error) {
	value, err := json.Marshal(resErr)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// SerializeJSON converts error list into serialized json string
func (errList ErrList) SerializeJSON() (string, error) {
	value, err := json.Marshal(errList)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// Error returns a general response error
func Error(msg string, status int) ResponseError {
	return ResponseError{msg, status, nil}
}

func (err *ResponseError) Error() string {
	return err.Msg
}
