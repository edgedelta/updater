package core

import (
	"fmt"
	"strings"
)

type Errors struct {
	errors []string
}

func NewErrors() *Errors {
	return &Errors{
		errors: make([]string, 0),
	}
}

func (e *Errors) Addf(format string, args ...any) {
	format = "- " + format
	e.errors = append(e.errors, fmt.Sprintf(format, args...))
}

func (e *Errors) ErrorOrNil() error {
	if len(e.errors) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(e.errors, "\n"))
}
