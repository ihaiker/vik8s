package cmd

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/paths"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/kvz/logstreamer"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"gopkg.in/fatih/color.v1"
	"io"
	"log"
	"os"
	"strings"
)

var colors = []color.Attribute{
	color.FgRed, color.FgGreen, color.FgYellow,
	color.FgBlue, color.FgMagenta, color.FgCyan,
}
var colorsSize = len(colors)

func filterNode(node *ssh.Node, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if filter == node.Hostname {
			return true
		}
	}
	return false
}

func runCmd(sync bool, cmd string, nodes []*ssh.Node, filters ...string) {
	run := func(i int, node *ssh.Node) {
		if !filterNode(node, filters) {
			return
		}
		prefix := color.New(colors[i%colorsSize]).Sprintf("[%s] ", node.Hostname)
		out := logstreamer.NewLogstreamerForStdout(prefix)
		if err := node.CmdStd(cmd, out, true); err != nil {
			fmt.Println(err)
			_, _ = out.Write([]byte(fmt.Sprint(err)))
		}
	}
	if sync {
		ssh.Sync(nodes, run)
	} else {
		for i, node := range nodes {
			run(i, node)
		}
	}
}

var bashCmd = &cobra.Command{
	Use: "bash", Short: "Run commands uniformly in the cluster",
	//SilenceErrors: true, SilenceUsage: true,
	PersistentPreRunE: hostsLoad(none),
	Run: func(cmd *cobra.Command, args []string) {
		nodes := hosts.Nodes()
		utils.Assert(len(nodes) > 0, "not found any host, use `vik8s host <node>` to add.")

		if len(args) > 0 {
			runCmd(false, strings.Join(args, " "), nodes)
			return
		}

		sync := false
		filters := make([]string, 0)

		term := liner.NewLiner()
		defer func() { _ = term.Close() }()
		term.SetCtrlCAborts(true)

		historyFile := paths.Join("history")
		if f, err := os.Open(historyFile); err == nil {
			_, _ = term.ReadHistory(f)
			_ = f.Close()
		}
		defer func() {
			if f, err := os.Create(historyFile); err != nil {
				log.Print("Error writing history file: ", err)
			} else {
				_, _ = term.WriteHistory(f)
				_ = f.Close()
			}
		}()

		for {
			line, err := term.Prompt("vik8s> ")
			if err == io.EOF {
				break
			} else if err == liner.ErrPromptAborted || len(line) == 0 {
				continue
			}

			line = strings.TrimSpace(line)
			if line[0] == '@' {
				filters = strings.Split(strings.TrimSpace(line[1:]), " ")
				continue
			} else if line == "&" {
				sync = true
				continue
			} else if line == "-" {
				sync = false
				filters = filters[0:0]
				continue
			}

			switch line {
			case "":
			case "clear":
				_, _ = os.Stdout.Write([]byte("\x1b[2J\x1b[0;0H"))
			case "exit", "quit":
				return
			default:
				runCmd(sync, line, nodes, filters...)
				term.AppendHistory(line)
			}
		}
	},
}
