package kubecerts

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testdata = "_testdata"

func TestKubernetesCerts(t *testing.T) {
	config := config.DefaultK8SConfiguration()
	defer os.RemoveAll(testdata)

	node := Node{
		Name:                "vm11",
		Host:                "10.24.1.11",
		ApiServer:           "vik8s-api-server",
		SvcCIDR:             "10.96.0.0/12",
		CertificateValidity: time.Hour * 24 * 365 * 10,
	}
	CreatePKIAssets(config.ApiServer, filepath.Join(testdata, "pki"), node)

	endpoint := fmt.Sprintf("https://%s:6443", "vik8s-api-server")
	files := CreateWorkerKubeConfigFile(testdata, "vm09", endpoint, time.Hour)

	t.Log("------------ worker -------------")
	for k, v := range files {
		t.Log(k, "=", v)
	}

	t.Log("------------ nodes -------------")
	files = CreateJoinControlPlaneKubeConfigFiles(testdata, "vm10", endpoint, time.Hour)
	for k, v := range files {
		t.Log(k, "=", v)
	}
}
