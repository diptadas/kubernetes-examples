# k8s extension apiserver

## API

```go
type Foo struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec string `json:"spec"
}
```

## Minikube walk-through

Setup everything:

```console
$ cd hack
$ ./build.sh
```

Check apiservice status:

```console
$ kubectl get apiservice v1alpha1.foocontroller.k8s.io -o yaml

status:
  conditions:
  - lastTransitionTime: 2018-03-02T08:27:15Z
    message: all checks passed
    reason: Passed
    status: "True"
    type: Available
```

Get Foo object named `foo-one`:

```yaml
$ kubectl get foo foo-one -o yaml

apiVersion: foocontroller.k8s.io/v1alpha1
kind: Foo
metadata:
  creationTimestamp: null
  name: foo-one
  namespace: default
  selfLink: /apis/foocontroller.k8s.io/v1alpha1/namespaces/default/foos/foo-one
spec: do-not-care
```
