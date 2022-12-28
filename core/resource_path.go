package core

import (
	"errors"
	"strings"
)

type K8sResourcePath string

type K8sResourceIdentifier struct {
	Namespace     string
	Kind          K8sResourceKind
	Name          string
	UpdateKeyPath string
}

func (rp K8sResourcePath) Parse() (*K8sResourceIdentifier, error) {
	sp := strings.Split(string(rp), ":")
	if len(sp) != 3 {
		return nil, errors.New("invalid schema, wrong number of semicolon-separated items")
	}
	ri := &K8sResourceIdentifier{
		Namespace:     sp[0],
		UpdateKeyPath: sp[2],
	}
	sp = strings.Split(sp[1], "/")
	if len(sp) != 2 {
		return nil, errors.New("invalid schema, wrong number of slash-separated items")
	}
	ri.Kind = K8sResourceKind(sp[0])
	ri.Name = sp[1]
	return ri, nil
}
