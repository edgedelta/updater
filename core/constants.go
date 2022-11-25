package core

type K8sResourceKind string

const (
	K8sDaemonset K8sResourceKind = "ds"
)

var (
	SupportedK8sResourceKinds = map[K8sResourceKind]bool{
		K8sDaemonset: true,
	}
)
