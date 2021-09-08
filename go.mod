module github.com/diptadas/kubernetes-examples

go 1.16

require (
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	metacontroller v0.0.0-00010101000000-000000000000
)

replace metacontroller => github.com/metacontroller/metacontroller v1.5.20
