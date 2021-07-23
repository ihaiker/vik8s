package kubecerts

import (
	"fmt"
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
	dir := "../../bin"
	CreatePKIAssets(dir, node)
}

func TestKubeConfig(t *testing.T) {
	dir := "../../bin/test"
	endpoint := fmt.Sprintf("https://%s:6443", "vik8s-api-server")
	files := CreateWorkerKubeConfigFile(dir, "vm09", endpoint, time.Hour)

	t.Log("------------ worker -------------")
	for k, v := range files {
		t.Log(k, "=", v)
	}

	t.Log("------------ nodes -------------")
	files = CreateJoinControlPlaneKubeConfigFiles(dir, "vm10", endpoint, time.Hour)
	for k, v := range files {
		t.Log(k, "=", v)
	}
}
