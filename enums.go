package main

import (
	"encoding/json"
	"strings"
)

type ErrorHandling int8

const (
	Ignore ErrorHandling = iota
	Warn
	Fatal
)

var errorHandlingStrings = [...]string{
	"Ignore",
	"Warn",
	"Fatal",
}

func ParseErrorHandling(s string) ErrorHandling {
	for i := range errorHandlingStrings {
		if strings.EqualFold(s, errorHandlingStrings[i]) {
			return ErrorHandling(i)
		}
	}
	return -1 // Stands for "Unknown".
}

func (eh ErrorHandling) String() string {
	if eh < Ignore || eh > Fatal {
		return "Unknown"
	}
	return errorHandlingStrings[eh]
}

func (eh ErrorHandling) MarshalJSON() ([]byte, error) {
	s := eh.String()
	return json.Marshal(s)
}

func (eh *ErrorHandling) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*eh = ParseErrorHandling(s)
	return nil
}
