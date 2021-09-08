package admission

import (
	"fmt"
	"net"

	clientset "github.com/diptadas/kubernetes-examples/admission-webhook/client/clientset/versioned"
	"github.com/openshift/generic-admission-server/pkg/apiserver"
	admission "k8s.io/api/admission/v1beta1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/admission.foocontroller.k8s.io"

func Run(stopCh <-chan struct{}, fooClient clientset.Interface) error {
	recommendedOptions := genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, apiserver.Codecs.LegacyCodec(admission.SchemeGroupVersion))
	recommendedOptions.Etcd = nil
	recommendedOptions.SecureServing.BindPort = 8443
	recommendedOptions.SecureServing.ServerCert.CertKey.CertFile = "/var/serving-cert/tls.crt"
	recommendedOptions.SecureServing.ServerCert.CertKey.KeyFile = "/var/serving-cert/tls.key"
	recommendedOptions.Authentication.SkipInClusterLookup = true

	if err := recommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)
	if err := recommendedOptions.ApplyTo(serverConfig, apiserver.Scheme); err != nil {
		return err
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig: apiserver.ExtraConfig{
			AdmissionHooks: []apiserver.AdmissionHook{
				&FooValidator{fooClient: fooClient},
				&FooMutator{},
			},
		},
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
