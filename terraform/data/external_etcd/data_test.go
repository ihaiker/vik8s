package external_etcd_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tf "github.com/ihaiker/vik8s/terraform"
	"os"
	"testing"
)

func providerFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"vik8s": func() (*schema.Provider, error) {
			return tf.Provider(), nil
		},
	}
}

func testAccPreCheck(t *testing.T) {}

func init() {
	_ = os.Setenv("TF_ACC", "true")
	_ = os.Setenv("TF_LOG", "debug")
}

func TestExternalEtcd(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories(),
		CheckDestroy:      resource.ComposeTestCheckFunc(),
		Steps: []resource.TestStep{
			{
				Config: `
data "vik8s_external_etcd" "etcd" {
	endpoints = ["https://10.24.2.10"]
	ca_raw = "raw"
	cert_raw = "raw"
	key_raw = "raw"
}
`,
				Check: resource.ComposeTestCheckFunc(),
			},
		},
	})
}
