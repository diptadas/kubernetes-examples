# k8s admission webhook

## API

```go
type Foo struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec FooSpec `json:"spec"`
}

type FooSpec struct {
    ConfigMapName string   `json:"configMapName"`
    Data          []string `json:"data"`
}
```

## Mutation

- Path: `/apis/admission.foocontroller.k8s.io/v1alpha1/mutations`
- Operations:
  - CREATE: Allow if `configMapName` not empty and add `deny-delete=true` annotation.

## Validation

- Path: `/apis/admission.foocontroller.k8s.io/v1alpha1/validations`
- Operations:
  - UPDATE: Don't allow if `configMapName` changed.
  - DELETE: Don't allow if annotation `deny-delete=true`.

## Controller

Creates a configmap with name specified in `configMapName`.

## Minikube walk-through

Setup everything:

```console
$ cd hack
$ ./build.sh
```

Check apiservice status:

```console
$ kubectl get apiservice v1alpha1.admission.foocontroller.k8s.io -o yaml

status:
  conditions:
  - lastTransitionTime: 2018-03-02T06:36:51Z
    message: all checks passed
    reason: Passed
    status: "True"
    type: Available
```

Try to create a Foo object without `configMapName`, it should fail:

```yaml
apiVersion: foocontroller.k8s.io/v1alpha1
kind: Foo
metadata:
  name: foo-one
spec:
  data:
  - hello
  - world
```

```console
$ kubectl apply -f foo-one-inv.yaml
Error from server (InternalError): error when creating "foo-one-inv.yaml": Internal error occurred: admission webhook "mutation.foocontroller.k8s.io" denied the request: configMapName not specified
```

Create a valid Foo object:

```yaml
apiVersion: foocontroller.k8s.io/v1alpha1
kind: Foo
metadata:
  name: foo-one
spec:
  configMapName: foo-one
  data:
  - hello
  - world
```

```console
$  kc apply -f foo-one.yaml
foo "foo-one" created
```

The mutation webhook will add `deny-delete=true` annotation:

```yaml
$ kubectl get foo foo-one -o yaml

apiVersion: foocontroller.k8s.io/v1alpha1
kind: Foo
metadata:
  annotations:
    deny-delete: "true"
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"foocontroller.k8s.io/v1alpha1","kind":"Foo","metadata":{"annotations":{},"name":"foo-one","namespace":"default"},"spec":{"configMapName":"foo-one","data":["hello","world"]}}
  clusterName: ""
  creationTimestamp: 2018-03-02T06:46:23Z
  name: foo-one
  namespace: default
  resourceVersion: "4935"
  selfLink: /apis/foocontroller.k8s.io/v1alpha1/namespaces/default/foos/foo-one
  uid: 6c03c5a9-1de5-11e8-9f57-080027fe67be
spec:
  configMapName: foo-one
  data:
  - hello
  - world
```

And the controller will create a configmap named `foo-one`:

```yaml
$ kc get configmap foo-one -o yaml

apiVersion: v1
data:
  foo-one: hello,world
kind: ConfigMap
metadata:
  creationTimestamp: 2018-03-02T06:46:23Z
  name: foo-one
  namespace: default
  ownerReferences:
  - apiVersion: v1alpha1
    kind: Foo
    name: foo-one
    uid: 6c03c5a9-1de5-11e8-9f57-080027fe67be
  resourceVersion: "4936"
  selfLink: /api/v1/namespaces/default/configmaps/foo-one
  uid: 6c05a637-1de5-11e8-9f57-080027fe67be
```

Now try to update `configMapName` of `foo-one`, it should fail:

```yaml
apiVersion: foocontroller.k8s.io/v1alpha1
kind: Foo
metadata:
  name: foo-one
spec:
  configMapName: foo-one-new
  data:
  - hello
  - world
```

```console
$ kubectl apply -f foo-one-new.yaml

Error from server (BadRequest): error when applying patch:
{"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"foocontroller.k8s.io/v1alpha1\",\"kind\":\"Foo\",\"metadata\":{\"annotations\":{},\"name\":\"foo-one\",\"namespace\":\"default\"},\"spec\":{\"configMapName\":\"foo-one-new\",\"data\":[\"hello\",\"world\"]}}\n"}},"spec":{"configMapName":"foo-one-new"}}
to:
&{0xc4200fab40 0xc420342bd0 default foo-one foo-one-new.yaml 0xc42000ccc8 4935 false}
for: "foo-one-new.yaml": admission webhook "validation.foocontroller.k8s.io" denied the request: invalid configMapName
```

Now try to delete `foo-one`, it should fail since annotation `deny-delete=true`:

```console
$ kubectl delete foo foo-one

Error from server (BadRequest): admission webhook "validation.foocontroller.k8s.io" denied the request: force denied delete
```

Remove the annotation and try to delete again:

```console
$ kubectl annotate foo foo-one deny-delete-

foo "foo-one" annotated
```

```console
$ kubectl delete foo foo-one

foo "foo-one" deleted
```

It will also delete the `foo-one` configmap.