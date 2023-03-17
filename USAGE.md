# Kubescrub Operator

_This operator deploys the kubescrub application, wich consists of a React frontend and go backend developed over a few days, which makes educated predictions about whether resources are orphaned in Kubernetes._

This operator deploys:
- Ingress 
- Frontend Deployment
- Backend Deployment
- Frontend Service
- Backend Service
- Service Account for backend (Needs permissions to look at resources)
- ClusterRole for SA
- ClusterRoleBinding for SA

**TOC**


- [Usage](#usage)
- [Prereqs](#prereqs)
- [Tutorial](#tutorial)
- [Sites](#sites)
- [Clean Up](#clean-up)
## Usage

| Name         | Type   | Description                                                   | Values                                 | Optional                                               |
| ------------ | ------ | ------------------------------------------------------------- | -------------------------------------- | ------------------------------------------------------ |
| Poll         | string | Frontend polls backend for updates                            | true, false                            | true, default to true                                  |
| PollInterval | string | Interval that frontend queries backend for updates in seconds | "5", "60"                              | true, defaults to 5 seconds                            |
| Namespaces   | string | Namespaces to watch for orphaned resources                    | "default, kube-system, kube-public"    | true, defaults to all                                  |
| Resources    | string | types of resources to watch                                   | "ConfigMaps, ServiceAccounts, Secrets" | true, defaults to ConfigMaps, ServiceAccounts, Secrets |
| Theme        | string | Dark or light theme for the frontend                          | "dark", "light"                        | "dark" (defaults to dark)                              |

## Prereqs

_You need Kind for to run this demo_ (if you don't have kind you could just port-forward the backend (kubescrub) to localhost:8080 and port-forward the frontend (kubescrub-web) to another localhost port ). The frontend speaks to the backend at localhost:8080, to facilitate this, i have configured the kind cluster to map NodePort 31469 to hostPort 8080. NortPort 31469 is the nodePort of the `ingress-nginx-controller` service. Frontend -> Backend communicated leverages the ingress (as an operator dependency) to direct traffic. 

## Tutorial

_In this tutorial, we will deploy the Kubescrub operator and demonstrate the usage/configurations of the Kubescrub application._

Spin up the kind cluster:

```bash
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 31469 #NodePort of NGINX service
    hostPort: 8080
    protocol: TCP
EOF
```


_This will deploy the kubescrub operator and all of the depedencies, which include NGINX Ingress dependencies._

```bash
make deploy
```

Wait for the operator's controller-manager deployment to become ready:

```bash
kubectl wait --for=condition=Ready pod -l control-plane=controller-manager -n kubescrub-operator-system --timeout=180s

kubectl wait --for=condition=Ready pod -l app.kubernetes.io/component=controller -n ingress-nginx --timeout=180s
```
Set our context to the `kubescrub-operator-system` namespace

```bash
kubectl config set-context $(kubectl config current-context) --namespace=kubescrub-operator-system
```

You may choose to follow the logs in one terminal of the operator controller-manager:

```bash
k logs -l control-plane=controller-manager -f 
```


Create an instance of the `Kubescrub` operator running the `dark` theme, looking for `ConfigMaps, ServiceAccounts, and Secrets` in namespaces `default, ingress-nginx, and kube-system` that is polling for updates every 5 seconds.

```bash
kubectl apply -f -<<EOF
apiVersion: infra.caseywylie.io/v1alpha1
kind: Reaper
metadata:
  labels:
    app.kubernetes.io/name: reaper
    app.kubernetes.io/instance: reaper-sample
    app.kubernetes.io/part-of: kubescrub-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubescrub-operator
  name: carolina
spec:
  theme: dark
  resources: "ConfigMap,Secret,ServiceAccount"
  namespaces: "default,ingress-nginx,kube-system"
  poll: "true"
  pollInterval: "5"
EOF
```

Go to [localhost:8080](http://localhost:8080) and you can see the frontend in action

## Sites

-[NGINX Ingress](https://kubernetes.github.io/ingress-nginx/deploy/#quick-start)
