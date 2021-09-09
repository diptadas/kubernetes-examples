# Dynamic Informer

K8s dynamic informer using [metacontroller](https://github.com/metacontroller/metacontroller).

```shell
$ kubectl apply -f crd.yaml
$ go run main.go
$ kubectl apply -f foo.yaml
```

```shell
$ kubectl delete -f foo.yaml
$ kubectl delete -f crd.yaml
```
