package kubecerts

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/libs/utils"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"path/filepath"
	"time"
)

type (
	clientCertAuth struct {
		CAKey         crypto.Signer
		Organizations []string
	}

	tokenAuth struct {
		Token string
	}

	kubeConfigSpec struct {
		CACert         *x509.Certificate
		APIServer      string
		ClientName     string
		TokenAuth      *tokenAuth
		ClientCertAuth *clientCertAuth
	}
)

const (
	AdminKubeConfigFileName             = "admin"
	ControllerManagerKubeConfigFileName = "controller-manager"
	KubeletKubeConfigFileName           = "kubelet"
	SchedulerKubeConfigFileName         = "scheduler"
	ClusterName                         = "kubernetes"
)

func CreateJoinControlPlaneKubeConfigFiles(dir, nodeName, controlPlaneEndpoint string, certificateValidity time.Duration) map[string]string {
	return createKubeConfigFiles(
		dir, ClusterName, nodeName, controlPlaneEndpoint, certificateValidity,
		AdminKubeConfigFileName, ControllerManagerKubeConfigFileName, KubeletKubeConfigFileName, SchedulerKubeConfigFileName,
	)
}

func CreateWorkerKubeConfigFile(dir, nodeName, controlPlaneEndpoint string, certificateValidity time.Duration) map[string]string {
	return createKubeConfigFiles(dir, ClusterName, nodeName, controlPlaneEndpoint, certificateValidity, KubeletKubeConfigFileName)
}

func createKubeConfigFiles(dir, clusterName, nodeName, controlPlaneEndpoint string, certificateValidity time.Duration, kubeConfigFileNames ...string) map[string]string {
	files := make(map[string]string)
	specs := getKubeConfigSpecs(dir, nodeName, controlPlaneEndpoint)
	for _, kubeConfigFileName := range kubeConfigFileNames {
		spec := specs[kubeConfigFileName]
		config := buildKubeConfigFromSpec(spec, clusterName, certificateValidity)
		kubeConfigFilePath := filepath.Join(dir, fmt.Sprintf("%s-%s.conf", kubeConfigFileName, nodeName))
		err := clientcmd.WriteToFile(*config, kubeConfigFilePath)
		utils.Panic(err, "write config %s", kubeConfigFileName)
		files[kubeConfigFileName] = kubeConfigFilePath
	}
	return files
}

func buildKubeConfigFromSpec(spec *kubeConfigSpec, clusterName string, certificateValidity time.Duration) *clientcmdapi.Config {
	if spec.TokenAuth != nil {
		return createWithToken(
			spec.APIServer,
			clusterName,
			spec.ClientName,
			certs.EncodeCertPEM(spec.CACert),
			spec.TokenAuth.Token,
		)
	}

	config := &certs.Config{
		Config: &certutil.Config{
			CommonName:   spec.ClientName,
			Organization: spec.ClientCertAuth.Organizations,
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
		CertificateValidity: certificateValidity,
	}

	clientCert, clientKey := certs.NewCertAndKey(spec.CACert, spec.ClientCertAuth.CAKey.(*rsa.PrivateKey), config)

	encodedClientKey, _ := keyutil.MarshalPrivateKeyToPEM(clientKey)

	return createWithCerts(
		spec.APIServer,
		clusterName,
		spec.ClientName,
		certs.EncodeCertPEM(spec.CACert),
		encodedClientKey,
		certs.EncodeCertPEM(clientCert),
	)
}

func createWithCerts(serverURL, clusterName, userName string, caCert []byte, clientKey []byte, clientCert []byte) *clientcmdapi.Config {
	config := createBasic(serverURL, clusterName, userName, caCert)
	config.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		ClientKeyData:         clientKey,
		ClientCertificateData: clientCert,
	}
	return config
}

func createWithToken(serverURL, clusterName, userName string, caCert []byte, token string) *clientcmdapi.Config {
	config := createBasic(serverURL, clusterName, userName, caCert)
	config.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		Token: token,
	}
	return config
}

func createBasic(serverURL, clusterName, userName string, caCert []byte) *clientcmdapi.Config {
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	return &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server:                   serverURL,
				CertificateAuthorityData: caCert,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{},
		CurrentContext: contextName,
	}
}

func getKubeConfigSpecs(dir, nodeName, controlPlaneEndpoint string) map[string]*kubeConfigSpec {
	caCert, caKey := certs.TryLoadCertAndKeyFromDisk(filepath.Join(dir, "pki"), "ca")

	var kubeConfigSpec = map[string]*kubeConfigSpec{

		AdminKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  controlPlaneEndpoint,
			ClientName: "kubernetes-admin",
			ClientCertAuth: &clientCertAuth{
				CAKey:         caKey,
				Organizations: []string{"system:masters"},
			},
		},

		ControllerManagerKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  controlPlaneEndpoint,
			ClientName: "system:kube-controller-manager",
			ClientCertAuth: &clientCertAuth{
				CAKey: caKey,
			},
		},

		SchedulerKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  controlPlaneEndpoint,
			ClientName: "system:kube-scheduler",
			ClientCertAuth: &clientCertAuth{
				CAKey: caKey,
			},
		},

		KubeletKubeConfigFileName: {
			CACert:     caCert,
			APIServer:  controlPlaneEndpoint,
			ClientName: fmt.Sprintf("system:node:%s", nodeName),
			ClientCertAuth: &clientCertAuth{
				CAKey:         caKey,
				Organizations: []string{"system:nodes"},
			},
		},
	}
	return kubeConfigSpec
}
