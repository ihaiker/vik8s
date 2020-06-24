package kubecerts

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/tools"
	"testing"
	"time"
)

func TestCreatePKIAssets(t *testing.T) {
	node := Node{
		Name:                "vm11",
		Host:                "10.24.1.11",
		ApiServer:           "vik8s-api-server",
		SvcCIDR:             "10.96.0.0/12",
		CertificateValidity: time.Hour * 24 * 365 * 10,
	}
	dir := tools.Join("kube", "pki")
	CreatePKIAssets(dir, node)
}

func TestKubeConfig(t *testing.T) {
	dir := tools.Join("kube")
	endpoint := fmt.Sprintf("https://%s:6443", "vik8s-api-server")
	files := CreateJoinControlPlaneKubeConfigFiles(dir, "vm10", endpoint, time.Hour)
	for k, v := range files {
		t.Log(k, "=", v)
	}
}
