package core

import (
	"fmt"
	yamls "github.com/ihaiker/vik8s/yaml"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use: "config", Short: "Show yaml file used by vik8s deployment cluster",
	Args: cobra.ExactValidArgs(1), ValidArgs: yamls.AssetNames(),
	Example: "vik8s config all",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(string(yamls.MustAsset(args[0])))
	},
}
var configNamesCmd = &cobra.Command{
	Use: "names", Short: "show file names",
	Run: func(cmd *cobra.Command, args []string) {
		for _, name := range yamls.AssetNames() {
			fmt.Println(name)
		}
	},
}

func init() {
	configCmd.AddCommand(configNamesCmd)
	rootCmd.AddCommand(configCmd)
}
