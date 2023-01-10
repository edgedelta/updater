package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/edgedelta/updater/core"
	"github.com/edgedelta/updater/log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

type NewClientOpt func(*Client)

func WithConfig(config *rest.Config) NewClientOpt {
	return func(c *Client) {
		c.config = config
	}
}

func NewClient(opts ...NewClientOpt) (*Client, error) {
	cli := &Client{}
	for _, o := range opts {
		o(cli)
	}
	var err error
	if cli.config == nil {
		cli.config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	cli.clientset, err = kubernetes.NewForConfig(cli.config)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (c *Client) SetResourceKeyValue(ctx context.Context, path core.K8sResourcePath, updateValue string) error {
	res, err := path.Parse()
	if err != nil {
		return fmt.Errorf("path.Parse: %v", err)
	}
	if _, ok := core.SupportedK8sResourceKinds[res.Kind]; !ok {
		return fmt.Errorf("k8s resource kind %q is not supported", res.Kind)
	}
	switch res.Kind {
	case core.K8sDaemonset:
		ds, err := c.clientset.AppsV1().DaemonSets(res.Namespace).Get(ctx, res.Name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("clientset.AppsV1.DaemonSets.Get: %v", err)
		}
		if ds == nil {
			return fmt.Errorf("no DaemonSet exists with name: %q, namespace: %q", res.Name, res.Namespace)
		}
		fieldSelectorPath := strings.Split(res.UpdateKeyPath, ".")
		old, updated, err := CompareAndUpdateStructField(ds, fieldSelectorPath, updateValue)
		if err != nil {
			return fmt.Errorf("k8s.CompareAndUpdateStructField: %v", err)
		}
		log.Info("Current daemonset image version is %s", old)
		if !updated {
			log.Info("Passing version update of resource with path %s to %s, older version is the same as the new one", path, updateValue)
			return nil
		}
		_, err = c.clientset.AppsV1().DaemonSets(res.Namespace).Update(ctx, ds, v1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("clientset.AppsV1.DaemonSets.Update: %v", err)
		}
		log.Info("Updated version of resource with path %s to %s", path, updateValue)
	case core.K8sDeployment:
		deploy, err := c.clientset.AppsV1().Deployments(res.Namespace).Get(ctx, res.Name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("clientset.AppsV1.Deployments.Get: %v", err)
		}
		if deploy == nil {
			return fmt.Errorf("no Deployment exists with name: %q, namespace: %q", res.Name, res.Namespace)
		}
		fieldSelectorPath := strings.Split(res.UpdateKeyPath, ".")
		old, updated, err := CompareAndUpdateStructField(deploy, fieldSelectorPath, updateValue)
		if err != nil {
			return fmt.Errorf("k8s.CompareAndUpdateStructField: %v", err)
		}
		log.Info("Current deployment image version is %s", old)
		if !updated {
			log.Info("Passing version update of resource with path %s to %s, older version is the same as the new one", path, updateValue)
			return nil
		}
		_, err = c.clientset.AppsV1().Deployments(res.Namespace).Update(ctx, deploy, v1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("clientset.AppsV1.Deployments.Update: %v", err)
		}
		log.Info("Updated version of resource with path %s to %s", path, updateValue)
	default:
		return fmt.Errorf("unsupported K8s resource kind: %q", res.Kind)
	}
	return nil
}

func (c *Client) GetSecret(ctx context.Context, namespace, name string) (string, error) {
	sc, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(sc.Data[name]), nil
}
