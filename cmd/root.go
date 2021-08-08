package cmd

import (
	"errors"
	"fmt"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"gopkg.in/fatih/color.v1"
	"os"
	"runtime"
	"strings"
)

//configLoad load configuration
func configLoad(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := config.Load(paths.Vik8sConfiguration()); err != nil {
			return err
		}
		return fn(cmd, args)
	}
}

//hostsLoad load configuration
func hostsLoad(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		hosts.Load(paths.HostsConfiguration(), &hosts.Option{
			Port:       22,
			User:       "root",
			PrivateKey: "$HOME/.ssh/id_rsa",
		}, true)
		return fn(cmd, args)
	}
}

//configDown the configuration save it.
func configDown(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := config.Config.Write(); err != nil {
			return err
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
		paths.ConfigDir, "The folder where the configuration file is located")
	rootCmd.PersistentFlags().StringVarP(&paths.Cloud, "cloud", "c", paths.Cloud,
		"Multi-kubernetes cluster selection")
	rootCmd.PersistentFlags().BoolVar(&paths.China, "china", true, "Whether domestic network")

	rootCmd.AddCommand(hostsCmd, etcdCmd)
	rootCmd.AddCommand(initCmd, joinCmd, resetCmd, cleanCmd)
	//rootCmd.AddCommand(ingressRootCmd, sidecarsCmd)
	//rootCmd.AddCommand(completionCmd)
	//rootCmd.AddCommand(bashCmd)
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
