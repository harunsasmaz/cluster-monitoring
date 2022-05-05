package cluster

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
)

const (
	ServiceLabel          = "service"
	ApplicationGroupLabel = "applicationGroup"
)

type Client struct {
	manager *kubernetes.Clientset
}

func NewClient(config *rest.Config) (*Client, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{manager: clientSet}, nil
}

func (c *Client) deploymentsWithLabels(ctx context.Context, namespace string, params ...LabelParams) (*v1.DeploymentList, error) {
	deployments, err := c.manager.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: toLabelSelector(params...),
	})
	if err != nil {
		return nil, err
	}

	return deployments, nil
}

func (c *Client) runningPodsWithLabels(ctx context.Context, namespace string, params ...LabelParams) (*apiv1.PodList, error) {
	pods, err := c.manager.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: toLabelSelector(params...),
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func (c *Client) getServiceInfo(ctx context.Context, namespace string, params ...LabelParams) ([]*ServiceInfo, error) {
	deployments, err := c.deploymentsWithLabels(ctx, namespace, params...)
	if err != nil {
		return nil, err
	}

	var services []*ServiceInfo
	for _, deployment := range deployments.Items {
		name := deployment.Spec.Selector.MatchLabels[ServiceLabel]
		group := deployment.Labels[ApplicationGroupLabel]

		pods, err := c.runningPodsWithLabels(ctx, namespace, LabelParams{Label: ServiceLabel, Value: name})
		if err != nil {
			return nil, err
		}

		podCount := len(pods.Items)
		services = append(services, &ServiceInfo{
			Name:             name,
			ApplicationGroup: group,
			RunningPodsCount: podCount,
		})
	}

	return services, nil
}

func (c *Client) GetAllServices(ctx context.Context) ([]*ServiceInfo, error) {
	return c.getServiceInfo(ctx, "")
}

func (c *Client) GetServicesWithNamespace(ctx context.Context, namespace string) ([]*ServiceInfo, error) {
	return c.getServiceInfo(ctx, namespace)
}

func (c *Client) GetServicesWithLabels(ctx context.Context, namespace string, params ...LabelParams) ([]*ServiceInfo, error) {
	return c.getServiceInfo(ctx, namespace, params...)
}
