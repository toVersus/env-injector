# env-injector

This is a sample implmentation of Mutating Admission Webhook based on [Knative webhook library](https://pkg.go.dev/github.com/knative/pkg/webhook).

## Prerequisite

- [Install Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- Clone this repository

## Usage

Bootstrap local cluster (v1.17)

```
$ kind create cluster
```

Deploy Knative's sample app, which returns an env value specified in the `TARGET` key if it exists. In this case, no env variable is defined in the PodTemplate, therefore it will send back the default hardcoded text.

```bash
$ cd env-injector
$ kubectl apply -f sample/helloworld.yaml
```

Forward local port to the sample app

```bash
$ kubectl port-forward deploy/helloworld-go 8080:8080
```

Check if the response body from our sample app contains the following keyword in advance:

```bash
$ curl http://localhost:8080
Hello World!
```

Delete the sample app

```bash
$ kubectl delete -f sample/helloworld.yaml
```

Create a namespace for env-injector resources to be installed and set label to it to avoid bootstrapping failure occurred in env-injector deployment itself.

```bash
$ kubectl create ns injector
$ kubectl label ns injector toversus.dev.env-injector.exclude="true"
```

Install the env-injector resources

```bash
$ kubectl apply -f install/
```

Check if the env-injector pod is ready or not:

```bash
$ watch kubectl get po -n injector
```

A lot of noisy error logs can be found in env-injector container, but ignore them!

```bash
$ kubectl logs -n injector -l app=env-injector
```

Check if the caBundle is properly reconciled by env-injector controller

```bash
$ ./hack/compare_cabundle.sh
```

Check if the power of env-injector

```bash
$ kubectl apply -f sample/helloworld.yaml
$ kubectl port-forward deploy/helloworld-go 8080:8080

$ curl http://localhost:8080
Hello Sample Injector v1!
```

You can see that the earlier respnse body is replaced to the value of `TARGET` env injected by env-injector!

```bash
$ kubectl get deploy -l app=helloworld-go -o jsonpath='{.items[].spec.template.spec.containers[0].env}'
[map[name:TARGET value:Sample Injector v1]]
```
