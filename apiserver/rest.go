package apiserver

import (
	"errors"
	"log"

	"github.com/diptadas/k8s-extension-apiserver/apis/foocontroller/v1alpha1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
)

type REST struct{}

var _ rest.Creater = &REST{}
var _ rest.Getter = &REST{}
var _ rest.Lister = &REST{}
var _ rest.Watcher = &REST{}
var _ rest.GroupVersionKindProvider = &REST{}

func NewREST() *REST {
	return &REST{}
}

func (r *REST) Create(ctx apirequest.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, includeUninitialized bool) (runtime.Object, error) {
	log.Println("Create...")
	foo := obj.(*v1alpha1.Foo)
	log.Println(foo)
	return foo, nil
}

func (r *REST) New() runtime.Object {
	return &v1alpha1.Foo{}
}

func (r *REST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return v1alpha1.SchemeGroupVersion.WithKind("Foo")
}

func (r *REST) Get(ctx apirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	log.Println("Get...")

	ns, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, errors.New("missing namespace")
	}
	if len(name) == 0 {
		return nil, errors.New("missing search query")
	}

	resp := &v1alpha1.Foo{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "foocontroller.k8s.io/v1alpha1",
			Kind:       "Foo",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: "do-not-care",
	}

	return resp, nil
}

func (r *REST) NewList() runtime.Object {
	return &v1alpha1.FooList{}
}

func (r *REST) List(ctx apirequest.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	log.Println("List...")

	ns, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, errors.New("missing namespace")
	}

	resp := &v1alpha1.FooList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "foocontroller.k8s.io/v1alpha1",
			Kind:       "Foo",
		},
		Items: []v1alpha1.Foo{
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "foocontroller.k8s.io/v1alpha1",
					Kind:       "Foo",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-1",
					Namespace: ns,
				},
				Spec: "do-not-care",
			},
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "foocontroller.k8s.io/v1alpha1",
					Kind:       "Foo",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-2",
					Namespace: ns,
				},
				Spec: "do-not-care",
			},
		},
	}

	return resp, nil
}

func (r *REST) Watch(ctx apirequest.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	log.Println("Watch...")
	fw := NewFooWatcher()
	return fw, nil
}
