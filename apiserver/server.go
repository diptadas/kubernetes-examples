package apiserver

import (
	"fmt"
	"net"

	"log"

	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apimachinery"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	restclient "k8s.io/client-go/rest"
)

const defaultEtcdPathPrefix = "/registry/admission.foocontroller.k8s.io"

func Run(kubeClientConfig *restclient.Config, stopCh <-chan struct{}) error {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)

	recommendedOptions := genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, codecs.LegacyCodec(admission.SchemeGroupVersion))
	recommendedOptions.Etcd = nil

	if err := recommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("192.168.99.100", nil, []net.IP{net.ParseIP("192.168.99.100")}); err != nil {
		log.Println("31...")
		return fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(codecs)
	if err := recommendedOptions.ApplyTo(serverConfig, scheme); err != nil {
		log.Println("36...")
		return err
	}

	genericServer, err := serverConfig.Complete().New("foo-apiserver", genericapiserver.EmptyDelegate)
	if err != nil {
		log.Println("43...")
		return err
	}

	// no idea whats going down here

	accessor := meta.NewAccessor()
	versionInterfaces := &meta.VersionInterfaces{
		ObjectConvertor:  scheme,
		MetadataAccessor: accessor,
	}
	interfacesFor := func(version schema.GroupVersion) (*meta.VersionInterfaces, error) {
		if version != admission.SchemeGroupVersion {
			return nil, fmt.Errorf("unexpected version %v", version)
		}
		return versionInterfaces, nil
	}
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{admission.SchemeGroupVersion}, interfacesFor)
	apiGroupInfo := genericapiserver.APIGroupInfo{
		GroupMeta: apimachinery.GroupMeta{
			SelfLinker:    runtime.SelfLinker(accessor),
			RESTMapper:    restMapper,
			InterfacesFor: interfacesFor,
			InterfacesByVersion: map[schema.GroupVersion]*meta.VersionInterfaces{
				admission.SchemeGroupVersion: versionInterfaces,
			},
		},
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{},
		OptionsExternalVersion:       &schema.GroupVersion{Version: "v1"},
		Scheme:                       scheme,
		ParameterCodec:               metav1.ParameterCodec,
		NegotiatedSerializer:         codecs,
	}

	admissionResource, singularResourceType := Resource()
	admissionVersion := admissionResource.GroupVersion()

	restMapper.AddSpecific(
		admission.SchemeGroupVersion.WithKind("AdmissionReview"),
		admissionResource,
		admissionVersion.WithResource(singularResourceType),
		meta.RESTScopeRoot)

	apiGroupInfo.GroupMeta.GroupVersions = appendUniqueGroupVersion(apiGroupInfo.GroupMeta.GroupVersions, admissionVersion)

	admissionReview := NewREST(Admit)
	v1alpha1storage := map[string]rest.Storage{
		admissionResource.Resource: admissionReview,
	}
	apiGroupInfo.VersionedResourcesStorageMap[admissionVersion.Version] = v1alpha1storage

	// just prefer the first one in the list for consistency
	apiGroupInfo.GroupMeta.GroupVersion = apiGroupInfo.GroupMeta.GroupVersions[0]

	if err := genericServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		log.Println("98...")
		return err
	}

	postStartName := fmt.Sprintf("admit-%s.%s.%s-init", admissionResource.Resource, admissionResource.Version, admissionResource.Group)
	genericServer.AddPostStartHookOrDie(postStartName,
		func(context genericapiserver.PostStartHookContext) error {
			return Initialize(kubeClientConfig, context.StopCh)
		},
	)

	return genericServer.PrepareRun().Run(stopCh)
}

func appendUniqueGroupVersion(slice []schema.GroupVersion, elems ...schema.GroupVersion) []schema.GroupVersion {
	m := map[schema.GroupVersion]bool{}
	for _, gv := range slice {
		m[gv] = true
	}
	for _, e := range elems {
		m[e] = true
	}
	out := make([]schema.GroupVersion, 0, len(m))
	for gv := range m {
		out = append(out, gv)
	}
	return out
}
