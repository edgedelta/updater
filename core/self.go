package core

import (
	"fmt"
	"os"
)

const (
	k8sNodeEnv      = "KUBE_NODE"
	k8sNamespaceEnv = "KUBE_NAMESPACE"
	k8sPodEnv       = "KUBE_POD"
)

type RuntimeInfo struct {
	Node      string
	Namespace string
	Pod       string
}

func GetSelfInfo() (*RuntimeInfo, error) {
	namespace := os.Getenv(k8sNamespaceEnv)
	if namespace == "" {
		namespace = getLocalResourceNamespace()
	}
	node := os.Getenv(k8sNodeEnv)
	if node == "" {
		var err error
		node, err = os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("os.Hostname: %v", err)
		}
	}
	return &RuntimeInfo{
		Node:      node,
		Namespace: namespace,
		Pod:       os.Getenv(k8sPodEnv),
	}, nil
}

func getLocalResourceNamespace() string {
	b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ""
	}
	return string(b)
}
