package core

import (
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

func GetSelfInfo() *RuntimeInfo {
	namespace := os.Getenv(k8sNamespaceEnv)
	if namespace == "" {
		namespace = getLocalResourceNamespace()
	}
	return &RuntimeInfo{
		Node:      os.Getenv(k8sNodeEnv),
		Namespace: namespace,
		Pod:       os.Getenv(k8sPodEnv),
	}
}

func getLocalResourceNamespace() string {
	b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ""
	}
	return string(b)
}
