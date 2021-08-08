package dashboard

import (
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var tokenPrintCmd = &cobra.Command{
	Use: "token", Short: "print admin-user token",
	Long: `kubectl -n kubernetes-dashboard get  secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')  -o jsonpath={.data.token} | base64 --decode`,
	Run: func(cmd *cobra.Command, args []string) {
		master := hosts.Get(config.K8S().Masters[0])
		token, err := master.SudoCmdString(cmd.Long)
		utils.Panic(err, "Get dashboard")
		fmt.Println(token)
	},
}
