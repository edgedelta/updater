package k8s

import (
	"context"
	"fmt"

	"github.com/edgedelta/updater/core"

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
		return fmt.Errorf("failed to parse resource path, err: %v", err)
	}
	if _, ok := core.SupportedK8sResourceKinds[res.Kind]; !ok {
		return fmt.Errorf("K8s resource kind %s is not supported", res.Kind)
	}
	switch res.Kind {
	case core.K8sDaemonset:
		ds, err := c.clientset.AppsV1().DaemonSets(res.Namespace).Get(ctx, res.Name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get DaemonSet %s in namespace %s, err: %v", res.Name, res.Namespace, err)
		}
		if ds == nil {
			return fmt.Errorf("no DaemonSet exists with name %s in namespace %s", res.Name, res.Namespace)
		}

		// TODO: Make this generic by using SetStructFieldValue()
		ds.Spec.Template.Spec.Containers[0].Image = updateValue

		_, err = c.clientset.AppsV1().DaemonSets(res.Namespace).Update(ctx, ds, v1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update the DaemonSet with new value %+v, err: %v", ds, err)
		}
	default:
		return fmt.Errorf("no handlers available for resource kind %s", res.Kind)
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
