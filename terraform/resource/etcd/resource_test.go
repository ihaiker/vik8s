package etcd_test

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

func testAccPreCheck(t *testing.T) {

}

func init() {
	_ = os.Setenv("TF_ACC", "true")
	_ = os.Setenv("TF_LOG", "debug")
}

func TestVik8sETCD(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories(),
		CheckDestroy:      resource.ComposeTestCheckFunc(),
		Steps: []resource.TestStep{
			{
				Config: `
resource "vik8s_etcd" "etcd" {
  nodes = ["192.168.10.176", "192.168.11.160", "192.168.11.152"]
}
`,
				Check: resource.ComposeTestCheckFunc(
				//resource.TestCheckResourceAttr("vik8s_hosts.master", "id", "/job/tf-acc-test-"+randString+"/job/subfolder/test-ssh"),
				),
			},
		},
	})
}
