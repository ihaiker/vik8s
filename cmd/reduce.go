package cmd

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/ihaiker/vik8s/reduce/kube"
	"github.com/ihaiker/vik8s/reduce/plugins"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var reduceCmd = &cobra.Command{
	Use: "reduce", Short: "Simplify kubernetes configure file",
	Args:    cobra.ExactArgs(1),
	PreRunE: configLoad(none),
	RunE: func(cmd *cobra.Command, args []string) error {
		if utils.NotExists(args[0]) {
			return fmt.Errorf("file not found: %s", args[0])
		}
		kube := kube.Reduce(args[0]).String()
		if output, _ := cmd.Flags().GetString("output"); output != "" {
			outputFile, _ := filepath.Abs(output)
			return ioutil.WriteFile(outputFile, []byte(kube), 0600)
		} else {
			fmt.Println(kube)
			return nil
		}
	},
}

var reduceDemoCmd = &cobra.Command{
	Use: "demo", Short: "show config demo",
	Args: cobra.ExactValidArgs(1), ValidArgs: []string{},
	PreRunE: configLoad(none),
	Run: func(cmd *cobra.Command, args []string) {
		for _, m := range plugins.Manager {
			for _, name := range m.Names {
				if name == args[0] {
					fmt.Println(m.Demo)
				}
			}
		}
		for _, kind := range kube.ReduceKinds {
			for _, name := range kind.Names {
				if name == args[0] {
					fmt.Println(kind.Demo)
				}
			}
		}
	},
}

func init() {
	reduceCmd.Flags().StringP("output", "o", "", "Output content to file")
	plugins.Load()
	for _, m := range plugins.Manager {
		reduceDemoCmd.ValidArgs = append(reduceDemoCmd.ValidArgs, m.Names...)
	}
	for _, kind := range kube.ReduceKinds {
		reduceDemoCmd.ValidArgs = append(reduceDemoCmd.ValidArgs, kind.Names...)
	}
	reduceDemoCmd.Long = "Args: \n  " + strings.Join(reduceDemoCmd.ValidArgs, ", ")
	reduceCmd.AddCommand(reduceDemoCmd)
	rootCmd.AddCommand(reduceCmd)
}
