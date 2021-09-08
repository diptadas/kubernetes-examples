package apiserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"

	"github.com/diptadas/kubernetes-examples/extension-apiserver/apis/foocontroller/v1alpha1"
	"github.com/emicklei/go-restful"
	"github.com/google/go-github/github"
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
	"k8s.io/client-go/util/homedir"
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

	if PossiblyInCluster() {
		recommendedOptions.SecureServing.ServerCert.CertKey.CertFile = "/var/serving-cert/tls.crt"
		recommendedOptions.SecureServing.ServerCert.CertKey.KeyFile = "/var/serving-cert/tls.key"
	} else {
		kubeConfigPath := path.Join(homedir.HomeDir(), ".kube/config")
		recommendedOptions.Authentication.SkipInClusterLookup = true
		recommendedOptions.CoreAPI.CoreAPIKubeconfigPath = kubeConfigPath
		recommendedOptions.Authorization.RemoteKubeConfigFile = kubeConfigPath
		recommendedOptions.Authentication.RemoteKubeConfigFile = kubeConfigPath
	}

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

	{
		apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(v1alpha1.GroupName, registry, Scheme, metav1.ParameterCodec, Codecs)
		apiGroupInfo.GroupMeta.GroupVersion = v1alpha1.SchemeGroupVersion
		v1alpha1storage := map[string]rest.Storage{}
		v1alpha1storage["foos"] = NewREST()
		v1alpha1storage["bars"] = NewBarREST()
		apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

		if err := genericServer.InstallAPIGroup(&apiGroupInfo); err != nil {
			return err
		}
	}

	{
		// install web service
		wsPath := "/apis/foocontroller.k8s.io/v1alpha1/ws"
		genericServer.Handler.GoRestfulContainer.Add(getWebService(wsPath))
	}

	return genericServer.PrepareRun().Run(stopCh)
}

// https://github.com/cloud-ark/kubediscovery/blob/master/pkg/apiserver/apiserver.go
func getWebService(path string) *restful.WebService {
	log.Println("WS PATH:", path)

	ws := new(restful.WebService).Path(path)
	//ws.Consumes("*/*")
	//ws.Produces(restful.MIME_JSON, restful.MIME_XML)
	//ws.ApiVersion("foocontroller.k8s.io/v1alpha1")

	helloPath := "/hello"
	ws.Route(ws.GET(helloPath).To(helloHandler))

	echoPath := "/{message}/echo"
	ws.Route(ws.GET(echoPath).To(echoHandler))

	gitIssuePath := "/git/issue"
	ws.Route(ws.POST(gitIssuePath).To(gitIssueHandler))

	return ws
}

func helloHandler(request *restful.Request, response *restful.Response) {
	log.Println("Printing request...")
	log.Println(request.Request)
	response.Write([]byte("hello world"))
}

func echoHandler(request *restful.Request, response *restful.Response) {
	log.Println("Printing request...")
	log.Println(request.Request)
	message := request.PathParameter("message")
	response.Write([]byte(message))
}

func gitIssueHandler(request *restful.Request, response *restful.Response) {
	log.Println("Printing request...")

	eventType := request.Request.Header.Get("X-GitHub-Event")
	log.Println("Event:", eventType)

	if eventType == "issues" {
		var issueEvent github.IssueActivityEvent
		decoder := json.NewDecoder(request.Request.Body)
		if err := decoder.Decode(&issueEvent); err != nil {
			log.Println(err)
		}
		// oneliners.PrettyJson(issueEvent, "Issue Event")
	}
}

// github.com/appscode/kutil/meta
// PossiblyInCluster returns true if loading an inside-kubernetes-cluster is possible.
func PossiblyInCluster() bool {
	fi, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
		os.Getenv("KUBERNETES_SERVICE_PORT") != "" &&
		err == nil && !fi.IsDir()
}
