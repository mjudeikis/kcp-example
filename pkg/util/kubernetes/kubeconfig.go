package kubernetes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ghodss/yaml"
	"github.com/mjudeikis/kcp-example/pkg/util/file"
	"github.com/mjudeikis/kcp-example/pkg/util/validation"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

func GetRestConfigFromURL(url string) (*rest.Config, error) {
	var data []byte
	if validation.IsValidUrl(url) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("url [%s] not found", url)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("url [%s] failed to read", url)
		}
	} else {
		exists, _ := file.Exist(url)
		if !exists {
			return nil, fmt.Errorf("file [%s] not found", url)
		}
		var err error
		data, err = os.ReadFile(url)
		if err != nil {
			return nil, err
		}
	}

	var kubeconfig *clientcmdv1.Config
	err := yaml.Unmarshal(data, &kubeconfig)
	if err != nil {
		return nil, err
	}
	return RestConfigFromV1Config(kubeconfig)
}

// RestConfigFromV1Config takes a v1 config and returns a kubeconfig
func RestConfigFromV1Config(kc *clientcmdv1.Config) (*rest.Config, error) {
	var c clientcmdapi.Config
	err := latest.Scheme.Convert(kc, &c, nil)
	if err != nil {
		return nil, err
	}

	kubeconfig := clientcmd.NewDefaultClientConfig(c, &clientcmd.ConfigOverrides{})
	return kubeconfig.ClientConfig()
}

func MakeKubeconfig(server, namespace, token string) ([]byte, error) {
	return json.MarshalIndent(&clientcmdv1.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []clientcmdv1.NamedCluster{
			{
				Name: "cluster",
				Cluster: clientcmdv1.Cluster{
					Server:                server,
					InsecureSkipTLSVerify: true,
				},
			},
		},
		AuthInfos: []clientcmdv1.NamedAuthInfo{
			{
				Name: "user",
				AuthInfo: clientcmdv1.AuthInfo{
					Token: token,
				},
			},
		},
		Contexts: []clientcmdv1.NamedContext{
			{
				Name: "cluster",
				Context: clientcmdv1.Context{
					Cluster:   "cluster",
					Namespace: namespace,
					AuthInfo:  "user",
				},
			},
		},
		CurrentContext: "cluster",
	}, "", "    ")
}
