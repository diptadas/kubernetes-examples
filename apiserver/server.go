package apiserver

import (
	"fmt"
	"net"

	"k8s-admission-webhook/apis/foocontroller/v1alpha1"

	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/foocontroller.k8s.io"

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	Scheme               = runtime.NewScheme()
	Codecs               = serializer.NewCodecFactory(Scheme)
)

func init() {
	v1alpha1.Install(groupFactoryRegistry, registry, Scheme)
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

func Run(stopCh <-chan struct{}) error {
	recommendedOptions := genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion))
	recommendedOptions.Etcd = nil
	recommendedOptions.SecureServing.BindPort = 8443
	recommendedOptions.SecureServing.ServerCert.CertKey.CertFile = "/var/serving-cert/tls.crt"
	recommendedOptions.SecureServing.ServerCert.CertKey.KeyFile = "/var/serving-cert/tls.key"

	if err := recommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(Codecs)
	if err := recommendedOptions.ApplyTo(serverConfig, Scheme); err != nil {
		return err
	}

	serverConfig.Version = &version.Info{
		Major: "1",
		Minor: "1",
	}

	genericServer, err := serverConfig.Complete().New("foo-apiserver", genericapiserver.EmptyDelegate)
	if err != nil {
		return err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(v1alpha1.GroupName, registry, Scheme, metav1.ParameterCodec, Codecs)
	apiGroupInfo.GroupMeta.GroupVersion = v1alpha1.SchemeGroupVersion
	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage["foos"] = NewREST()
	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

	if err := genericServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return err
	}

	return genericServer.PrepareRun().Run(stopCh)
}
