package main

import "strings"

type ErrorHandling int8

const (
	Ignore ErrorHandling = iota + 1
	Warn
	Fatal
)

var errorHandlingStrings = [...]string{
	"Unknown",
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
	return 0 // Stands for "Unknown".
}

func (eh ErrorHandling) String() string {
	if eh < Ignore || eh > Fatal {
		return errorHandlingStrings[0]
	}
	return errorHandlingStrings[eh]
}

func (eh ErrorHandling) MarshalText() ([]byte, error) {
	return []byte(eh.String()), nil
}

func (eh *ErrorHandling) UnmarshalText(text []byte) error {
	*eh = ParseErrorHandling(string(text))
	return nil
}
