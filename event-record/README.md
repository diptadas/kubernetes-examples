# Event Record

```shell
$ go run main.go

2021/09/08 05:23:46 Creating new configmap as reference object
2021/09/08 05:23:46 Recording events using broadcaster
2021/09/08 05:23:46 Event recorded: example-configmap.16a2d1bd5d136870
2021/09/08 05:23:47 Creating events directly
2021/09/08 05:23:47 Event recorded: example-configmap.16a2d1bd98bebcd0
```

```shell
$ kubectl describe configmap example-configmap

Name:         example-configmap
Namespace:    default
Labels:       <none>
Annotations:  <none>

Data
====

BinaryData
====

Events:
  Type    Reason        Age   From             Message
  ----    ------        ----  ----             -------
  Normal  event-test-1  12s   golang-examples  new event is recorded
  Normal  event-test-2  11s   golang-examples  new event is recorded
```

```shell
$ kubectl get events

LAST SEEN   TYPE     REASON         OBJECT                        MESSAGE
28s         Normal   event-test-1   configmap/example-configmap   new event is recorded
27s         Normal   event-test-2   configmap/example-configmap   new event is recorded
```