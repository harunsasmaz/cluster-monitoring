# Cluster Monitoring

## Service Endpoints

#### Expose information on all pods in the cluster

Add an endpoint to the service that exposes all pods running in the cluster in a given namespace:

```
GET `/services/{namespace}`
[
  {
    "name": "first",
    "applicationGroup": "alpha",
    "runningPodsCount": 2
  },
  {
    "name": "second",
    "applicationGroup": "beta",
    "runningPodsCount": 1
  },
  ...
]
```

#### Expose information on a group of applications in the cluster

Create an endpoint in your service that exposes the pods in the cluster in a given namespace that are part of the same `applicationGroup`:

```
GET `/services/{namespace}/{applicationGroup}`
[
  {
    "name": "foobar",
    "applicationGroup": "<applicationGroup>",
    "runningPodsCount": 1
  },
  ...
]
```

## Creating a Cluster 

1. Apply Terraform definitions

```
cd infrastructure/terraform
terraform init
terraform apply
```

2. Authorize `kubectl`

```
gcloud container clusters get-credentials <CLUSTER_NAME>
```

3. Verify `kubectl`

```
kubectl cluster-info
```

4. Allow Kubernetes to pull images

```
kubectl create secret docker-registry gcr-access-token \
--docker-server=eu.gcr.io \
--docker-username=oauth3accesstoken \
--docker-password="$(gcloud auth print-access-token)" \
--docker-email=my@email.com
```

## Setup Knative & Istio

#### Upgrade cluster version to 1.22+

```
gcloud container clusters upgrade <CLUSTER_NAME> --master --latest
```

This may take a few minutes.

#### Install Knative Serving with YAML

First, you should check minimum system requirements for Knative to operate in Kubernetes cluster.

1. You should apply the following manifest to install required CRDs:

```
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.4.0/serving-crds.yaml
```

2. You should apply the following manifest to upload the core components: 

```
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.4.0/serving-core.yaml
```

3. Verify the installation

```
kubectl get pods -n knative-serving
```

Output should  be similar to the following, you should see all pods running

```
NAME                                      READY   STATUS    RESTARTS   AGE
3scale-kourier-control-54cc54cc58-mmdgq   1/1     Running   0          81s
activator-67656dcbbb-8mftq                1/1     Running   0          97s
autoscaler-df6856b64-5h4lc                1/1     Running   0          97s
controller-788796f49d-4x6pm               1/1     Running   0          97s
domain-mapping-65f58c79dc-9cw6d           1/1     Running   0          97s
domainmapping-webhook-cc646465c-jnwbz     1/1     Running   0          97s
webhook-859796bc7-8n5g2                   1/1     Running   0          96s
```

#### Install Istio with YAML

1. Install Istio on cluster

```
kubectl apply -l knative.dev/crd-install=true -f https://github.com/knative/net-istio/releases/download/knative-v1.4.0/istio.yaml
```

```
kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v1.4.0/istio.yaml
```

2. Install Knative Istio controller

```
kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v1.4.0/net-istio.yaml
```

3. Verify installation

```
kubectl get pods -n istio-system
```

Output should be similar to:

```
NAME                                    READY   STATUS    RESTARTS   AGE
istio-ingressgateway-666588bf64-7x66v   1/1     Running   0          120m
istio-ingressgateway-666588bf64-cvfnb   1/1     Running   0          120m
istio-ingressgateway-666588bf64-qh4gl   1/1     Running   0          120m
istiod-56967d8fcc-7w2n4                 1/1     Running   0          120m
istiod-56967d8fcc-lfmvk                 1/1     Running   0          120m
istiod-56967d8fcc-v44c7                 1/1     Running   0          120m
```

4. Fetch external IP of your ingress gateway

```
kubectl --namespace istio-system get service istio-ingressgateway
```

> External IP provided from this output will be used to configure custom domain and HTTPS configurations.

## Auto TLS and Custom Domain Configuration

#### Install Cert-Manager

1. [OPTIONAL] if you are using Google Kubernetes Engine, you should give permission to account of your GCP account by following:

```
kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account)
```

2. Install all cert-manager components 

```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
```

3. Verify installation

```
$ kubectl get pods --namespace cert-manager

NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-5c6866597-zw7kh               1/1     Running   0          2m
cert-manager-cainjector-577f6d9fd7-tr77l   1/1     Running   0          2m
cert-manager-webhook-787858fcdb-nlzsq      1/1     Running   0          2m
```

4. Install Knative Cert-Manager core components

```
kubectl apply -f https://github.com/knative/net-certmanager/releases/download/knative-v1.4.0/release.yaml
```

#### Custom Domain Configuration

For this project, I used my personal domain `harunsasmaz.com` which is registered on [Cloudflare](https://www.cloudflare.com/)

1. Open `config-domain` ConfigMap to edit

```
kubectl edit configmap config-domain -n knative-serving
```

2. Edit file to replace using your domain instad of example.com

First, delete all parts under `_example` key and then add your domain under `data` section as following. Note that, right side is `""` intentionally.

```yaml
apiVersion: v1
data:
  mydomain.com: ""
kind: ConfigMap
[...]
```

3. Publish your domain

First, you should visit your DNS provider dashboard. Then, create a **wildcard A record** with target IP as your cluster's external IP as we mentioned under `Istio` section.

```
*.default.mydomain.com   59     IN     A   <EXTERNAL_IP>
```

Here, 

- `*` is the wildcard, any URLs ending with `default.mydomain.com` will be redirected to provided IP.
- `default` means that your service running in `default` namespace.
- `A` is the record type.
- `EXTERNAP_IP` is the target address that your DNS provider will redirect incoming requests. 

#### Create a Cluster Issuer 

1. Create a YAML file using the following template

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-http01-issuer
spec:
  acme:
    privateKeySecretRef:
      name: letsencrypt
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
    - http01:
       ingress:
         class: istio

```

you may change names of private key secret and cluster issuer.

2. Apply the YAML file

```
kubectl apply -f <filename>.yaml
```

3. Verify ClusterIssuer

```
kubectl get clusterissuer <cluster-issuer-name>
```

You should see `READY` state is `True`

```
NAME                        READY   AGE
letsencrypt-http01-issuer   True    1m
```

#### Configure Cert-Manager ConfigMap

1. Open `config-certmanager` ConfigMap to edit

```
kubectl edit configmap config-certmanager --namespace knative-serving
```

2. Add issuer reference you created above with the data section

First, delete example section with the data section. Then, add the following:

```yaml
data:
  issuerRef: |
    kind: ClusterIssuer
    name: letsencrypt-http01-issuer
```

3. Verify that you updated file successfully

```
kubectl get configmap config-certmanager --namespace knative-serving --output yaml
```

#### Turn on Auto-TLS for HTTPS

1. Open `config-network` ConfigMap to edit

```
kubectl edit configmap config-network --namespace knative-serving
```

2. Enable Auto-TLS

First, delete example section with the data section. Then, add the following.

```yaml
data:
  auto-tls: Enabled
```

3. Verify that you updated file successfully

```
kubectl get configmap config-network --namespace knative-serving --output yaml
```

#### Verify Auto-TLS

1. Install a dummy Knative service

```
kubectl apply -f https://raw.githubusercontent.com/knative/docs/main/docs/serving/autoscaling/autoscale-go/service.yaml
```

2. Check for HTTPS

```
kubectl get ksvc autoscale-go
```

Output should be:

```
NAME           URL                                            LATESTCREATED        LATESTREADY          READY   REASON
autoscale-go   https://autoscale-go.default.harunsasmaz.com   autoscale-go-00001   autoscale-go-00001   True
```

Note: It might take a few minutes to provision a TLS certificate, until then you may see `http`.

* If you cannot get HTTPS connection, you can check the following for debugging:

```
kubectl describe certificate <CERTIFICATE_NAME>
```

## Testing

#### Create a new Knative service

```
make push-image
kubectl apply -f infrastructure/knative/service.yaml
```

You can check that HTTPS connection with your custom domain have been configured successfully by the following:

```
kubectl get ksvc
```

#### Deployment of test services

1. Give read permission to default system account

> The permission I gave described below is not encouraged, but applied for this sample project.

```
kubectl create clusterrolebinding admin-account \
  --clusterrole=cluster-admin \
  --group=system:serviceaccounts
```

2. Apply definitions

```
cd infrastructure/kubernetes
kubectl apply -f services.yaml
```

#### Testing via cURL

By adding `-v` flag, you will be able to see TLS handshakes and establishing a secure connection with your cluster.

`python -m json.tool` command provides pretty printing json responses.

```
$ curl -vX GET https://api-service.default.harunsasmaz.com/services/default | python -m json.tool 

> {
    "data": [
        {
            "applicationGroup": "beta",
            "name": "blissful-goodall",
            "runningPodsCount": 1
        },
        {
            "applicationGroup": "beta",
            "name": "confident-cartwright",
            "runningPodsCount": 1
        },
        {
            "applicationGroup": "",
            "name": "happy-colden",
            "runningPodsCount": 1
        },
        {
            "applicationGroup": "gamma",
            "name": "quirky-raman",
            "runningPodsCount": 1
        },
        {
            "applicationGroup": "alpha",
            "name": "stoic-sammet",
            "runningPodsCount": 2
        }
    ],
    "success": true
}
```

```
$ curl -vX GET https://api-service.default.harunsasmaz.com/services/default/beta | python -m json.tool

> {
    "data": [
        {
            "applicationGroup": "beta",
            "name": "blissful-goodall",
            "runningPodsCount": 1
        },
        {
            "applicationGroup": "beta",
            "name": "confident-cartwright",
            "runningPodsCount": 1
        }
    ],
    "success": true
}
```


