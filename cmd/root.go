package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/ihaiker/vik8s/config"
	hs "github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strings"
)

var configure *config.Configuration

//configLoad load configure
func configLoad(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		if configure, err = config.Load(paths.Vik8sConfiguration()); err != nil {
			return
		}
		if configure.Hosts, err = hs.New(paths.HostsConfiguration(), hs.Option{
			Port:       22,
			User:       "root",
			PrivateKey: "$HOME/.ssh/id_rsa",
		}); err != nil {
			return
		}
		return fn(cmd, args)
	}
}

//configDown the configure save it.
func configDown(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := configure.Write(); err != nil {
			return utils.Wrap(err, "write configuration error")
		}
		return fn(cmd, args)
	}
}

func none(cmd *cobra.Command, args []string) error {
	return nil
}

var rootCmd = &cobra.Command{
	Use: "vik8s", Short: "very easy install HA k8s",
	Long: "very easy install k8s。Build: %s, Go: %s, GitLog: %s",
}

var completionCmd = &cobra.Command{
	Use: "completion", Short: "generates completion scripts",
	Args: cobra.ExactArgs(1), ValidArgs: []string{"bash", "zsh", "powershell", "fish"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		default:
			return errors.New("not support")
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&paths.ConfigDir, "config", "f",
		paths.ConfigDir, "The folder where the configure file is located")
	rootCmd.PersistentFlags().StringVarP(&paths.Cloud, "cloud", "c", paths.Cloud,
		"Multi-kubernetes cluster selection")
	rootCmd.PersistentFlags().BoolVar(&paths.China, "china", true, "Whether domestic network")

	rootCmd.AddCommand(hostsCmd, etcdCmd)
	rootCmd.AddCommand(initCmd, joinCmd, resetCmd, cleanCmd)
	rootCmd.AddCommand(ingressRootCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(bashCmd)
	rootCmd.Flags().SortFlags = false
}

func Execute(version, buildTime, gitTag string) {
	rootCmd.Version = version
	rootCmd.Long = fmt.Sprintf(rootCmd.Long, buildTime, runtime.Version(), gitTag)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	defer utils.Catch(func(err error) {
		if serr, match := err.(*utils.WrapError); match {
			_, _ = color.New(color.FgRed).Println(serr.Error())
		} else {
			color.Red(err.Error())
			color.Red(utils.Stack())
		}
		os.Exit(1)
	})

	if runCommand, args, err := rootCmd.Find(os.Args[1:]); err == nil {
		if runCommand.Name() == rootCmd.Name() {
			for i, arg := range args {
				if arg == "--" {
					param := strings.Join(args[i+1:], " ")
					os.Args = append(os.Args[0:i+1], "bash", param)
					break
				}
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.HiRedString(err.Error()))
		os.Exit(1)
	}
}
