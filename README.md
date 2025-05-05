# Demo of Kubernetes basic controllers

This repo contains a few basic [kubernetes controller](https://kubernetes.io/docs/concepts/architecture/controller/) samples for internal demo.
They are generated using [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) scaffolding tool.

This readme contains minimal explanations on what they do and how they were generated.

## Local minikube setup

- `minikube start`
- `cd 04-webhooks && make deploy-cert-manager`
- `cd pg-healthchecker && docker build -t pg-healthchecker . && minikube image load pg-healthchecker`
- For each controller project (except `01-empty-project`, I guess):
  - cd to project directory
  - `make docker-build`
  - `minikube image load {image name}`
    - `{image name}` can be found at the top of `Makefile` (e.g. `webhook-controller`)
  - `make deploy`

## Controller descriptions

- 01 Empty project structure
  - Empty project just to showcase the structure. Doesn't do anything by itself
  - Generated using:
    - `cd 01-empty-project && kubebuilder init --domain minds.co --repo minds.co/repo`
- 02 Namespace Labeler
  - Basic controller that labels namespaces whose name has `dev-` prefix with `env: dev` label
  - Generated using:
    - `cd 02-namespace-labeler && kubebuilder init --domain minds.co --repo minds.co/repo`
    - `kubebuilder create api --group "core" --kind "Namespace" --version "v1" --resource=false --controller=true`
    - Implemented `internal/controller/namespace_controller.go`
  - Test it:
    - `kubectl create ns dev-test`
    - `kubectl get ns dev-test -o yaml`
      - verify that `.metadata.labels` contains `env: dev`
- 03 CRD Reconciliation
  - Has it's own Custom Resource Definition (CRD)
  - CRD has single input - `image`
  - controller creates a `postgres` deployment, deployment with image from CRD and passes connection string for postgres to it
  - There is `pg-healthchecker` helper project that will continuously log whether it can reach the database for demo purposes
  - Generated using:
    - `cd 03-app-with-db-crd && kubebuilder init --domain minds.co --repo minds.co/repo`
    - `kubebuilder create api --group minds.co --version v1 --kind AppWithDb --namespaced=true --resource=true --controller-true`
    - Implemented `internal/controller/namespace_controller.go`
  - Test it:
    - `kubectl create ns dev-test`
    - `cd 04-app-with-db-crd/config/samples`
    - `kubectl apply -f minds.co_v1_appwithdb.yaml -n dev-test`
    - Verify that postgres and app deployments are created using `kubectl get deployments -n dev-test`
- 04 Validation Webhooks
  - Validates whether all created/updated pods have defined non-zero values for:
    - `.spec.containers[].resources.requests.cpu`
    - `.spec.containers[].resources.requests.memory.`
    - `.spec.containers[].resources.limits.cpu`
    - `.spec.containers[].resources.limits.memory`
  - Generated using:
    - `cd 04-webhooks && kubebuilder init --domain minds.co --repo minds.co/repo`
    - `kubebuilder create api --group core --version v1 --kind Pod --resource=false --controller=false`
      - a bit of a hack, because kubebuilder doesn't "fully" support scaffolding stuff for core resources
    - `kubebuilder create webhook --group core --version v1 --kind Pod --programmatic-validation`
    - Implemented `internal/webhook/v1/pod_webhook`
  - Test it:
    - `kubectl run nginx --image=nginx`
    - output should be a list of errors and pod should not be created

