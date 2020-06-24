package dashboard

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/spf13/cobra"
)

var tokenPrintCmd = &cobra.Command{
	Use: "token", Short: "print admin-user token",
	Long: `kubectl -n kubernetes-dashboard get  secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')  -o jsonpath={.data.token} | base64 --decode`,
	Run: func(cmd *cobra.Command, args []string) {
		master := k8s.Config.Master()
		token := master.MustCmd2String(cmd.Long)
		fmt.Println(token)
	},
}
