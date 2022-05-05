package cluster

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var client *Client

func init() {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	if err != nil {
		log.Fatalf("cannot build config: %v", err)
	}
	client, _ = NewClient(config)
}

func TestClient_GetServices(t *testing.T) {
	type fields struct {
		manager *kubernetes.Clientset
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "on_success",
			fields: fields{
				manager: client.manager,
			},
			args: args{
				ctx: context.Background(),
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				manager: tt.fields.manager,
			}
			got, err := c.GetAllServices(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < 1 && tt.want {
				t.Errorf("GetAllServices() no items found")
				return
			}
		})
	}
}

func TestClient_GetServicesWithNamespace(t *testing.T) {
	type fields struct {
		manager *kubernetes.Clientset
	}
	type args struct {
		ctx       context.Context
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "on_success",
			fields: fields{
				manager: client.manager,
			},
			args: args{
				ctx:       context.Background(),
				namespace: apiv1.NamespaceDefault,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "on_fail",
			fields: fields{
				manager: client.manager,
			},
			args: args{
				ctx:       context.Background(),
				namespace: "does not exist",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				manager: tt.fields.manager,
			}
			got, err := c.GetServicesWithNamespace(tt.args.ctx, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServicesWithNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < 1 && tt.want {
				t.Errorf("GetServices() no items found")
				return
			}
		})
	}
}

func TestClient_GetServicesWithLabels(t *testing.T) {
	type fields struct {
		manager *kubernetes.Clientset
	}
	type args struct {
		ctx       context.Context
		namespace string
		params    []LabelParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "on_success_alpha",
			fields: fields{
				manager: client.manager,
			},
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				params: []LabelParams{
					{
						Label: "applicationGroup",
						Value: "alpha",
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "on_success_beta",
			fields: fields{
				manager: client.manager,
			},
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				params: []LabelParams{
					{
						Label: "applicationGroup",
						Value: "beta",
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "on_success_multiple",
			fields: fields{
				manager: client.manager,
			},
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				params: []LabelParams{
					{
						Label: "applicationGroup",
						Value: "beta",
					},
					{
						Label: "service",
						Value: "blissful-goodall",
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "on_fail",
			fields: fields{
				manager: client.manager,
			},
			args: args{
				ctx:       context.Background(),
				namespace: "default",
				params: []LabelParams{
					{
						Label: "nonexist",
						Value: "beta",
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				manager: tt.fields.manager,
			}
			got, err := c.GetServicesWithLabels(tt.args.ctx, tt.args.namespace, tt.args.params...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServicesWithLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < 1 && tt.want {
				t.Errorf("GetServices() no items found")
				return
			}
		})
	}
}
