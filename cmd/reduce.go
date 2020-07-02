package cmd

import (
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/kube"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
)

var reduceCmd = &cobra.Command{
	Use: "reduce", Short: "Simplify kubernetes configuration file",
	Args: cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		defer utils.Catch(func(err error) {})
		k8s.Config.Load()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		outputFile, _ := filepath.Abs(output)
		kube := kube.Parse(args[0]).String()
		return ioutil.WriteFile(outputFile, []byte(kube), 0600)
	},
}

func init() {
	reduceCmd.Flags().StringP("output", "o", "", "Output content to file")
	rootCmd.AddCommand(reduceCmd)
}
