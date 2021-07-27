package cmd

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/cni"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"io"
	"math/rand"
	"strings"
)

var cleanCmd = &cobra.Command{
	Use: "clean", Hidden: true,
	Short:   "This command is used to deeply clean up the environment, Use very carefully。Use very carefully。Use very carefully。",
	Example: `vik8s clean or vik8s clean 172.16.100.10`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			if !importantConfirmation() {
				fmt.Println("Verification code error")
				return
			}
		}
		k8s.Clean(hosts.Nodes(), cni.Plugins.Clean)
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
