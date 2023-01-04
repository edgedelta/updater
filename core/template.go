package core

import (
	"bytes"
	"fmt"
	"text/template"
)

type contextualVariableTemplate struct {
	Vars map[string]string
}

func EvaluateContextualTemplate(raw string, vars map[string]string) (string, error) {
	if vars == nil {
		return raw, nil
	}
	t, err := template.New("ctx-template").Parse(raw)
	if err != nil {
		return "", fmt.Errorf("template.New: %v", err)
	}
	b := new(bytes.Buffer)
	if err := t.Execute(b, &contextualVariableTemplate{
		Vars: vars,
	}); err != nil {
		return "", fmt.Errorf("template.Execute: %v", err)
	}
	return b.String(), nil
}
