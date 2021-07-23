package docker

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGenerateCertificate(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tmpDir)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if _, err := GenerateCertificate(tmpDir); err != nil {
		t.Fatal(err)
	}
}
