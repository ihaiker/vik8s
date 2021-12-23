package cmd

import (
	"github.com/ihaiker/cobrax"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/install/cri/docker"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
)

var criCmd = &cobra.Command{
	Use: "cri", Short: "defined kubernetes container runtime interface",
}

var dockerFlag = config.DefaultDockerConfiguration()
var dockerCmd = &cobra.Command{
	Use: "docker", Short: "defined kubernetes cni configure for docker",
	Example:  "vik8s docker --tls.enable",
	PreRunE:  configLoad(none),
	PostRunE: configDown(none),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := docker.Config(dockerFlag); err == nil {
			configure.Docker = dockerFlag
		}
		return nil
	},
}

func init() {
	err := cobrax.FlagsWith(dockerCmd, cobrax.GetFlags, dockerFlag, "", "VIK8S_DOCKER")
	utils.Panic(err, "setting flag error")
	dockerCmd.Flags().SortFlags = false

	criCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(criCmd)
}
