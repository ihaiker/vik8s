package cmd

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/cni"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"gopkg.in/fatih/color.v1"
	"io"
	"math/rand"
	"strings"
)

var cleanCmd = &cobra.Command{
	Use: "clean", Hidden: true, Args: cobra.MinimumNArgs(1),
	Short:   color.New(color.FgHiRed).Sprintf("This command is used to deeply clean up the environment. %s", strings.Repeat("Use very carefully", 3)),
	Example: `vik8s clean or vik8s clean 10.24.0.1`,
	PreRunE: configLoad(hostsLoad(none)),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			if !importantConfirmation() {
				fmt.Println("Verification code error")
				return
			}
		}
		var nodes []*ssh.Node
		if len(args) == 1 && args[0] == "all" {
			nodes = hosts.Nodes()
		} else {
			nodes = hosts.MustGets(args)
		}
		k8s.Clean(nodes, cni.Plugins.Clean)
	},
}

func importantConfirmation() bool {
	term := liner.NewLiner()
	defer func() { _ = term.Close() }()
	term.SetCtrlCAborts(true)

	code := fmt.Sprintf("%04d", rand.Intn(10000))
	for i := 0; i < 3; i++ {
		line, err := term.Prompt(fmt.Sprintf("Enter confirmation code [%s]> ", code))
		if err == io.EOF {
			break
		} else if err != nil && strings.Contains(err.Error(), "control-c break") {
			break
		}
		if code == line {
			return true
		}
	}
	return false
}

func init() {
	cleanCmd.Flags().Bool("force", false, "Clean the node without prompting for confirmation")
}
