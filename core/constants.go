package core

type K8sResourceKind string

const (
	K8sDaemonset   K8sResourceKind = "ds"
	K8sDeployment  K8sResourceKind = "deploy"
	K8sStatefulset K8sResourceKind = "sts"
)

var (
	SupportedK8sResourceKinds = map[K8sResourceKind]bool{
		K8sDaemonset:   true,
		K8sDeployment:  true,
		K8sStatefulset: true,
	}
)
