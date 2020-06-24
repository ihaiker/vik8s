package hosts

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"testing"
)

func TestIPS(t *testing.T) {
	ips := []string{
		"root:jianchi2313!23#@172.16.100.10:22",
		"ri_123:teset@172.16.100.10:22",
		"giir-12312:bbb@172.16.100.10-172.16.100.15",
		"root:$HOME/.ssh/id_rsa@172.16.100.10-172.16.100.15:22",
		"172.16.100.10-172.16.100.15:22",
		"172.16.100.10",
		"172.16.100.15:22",
		"vm10",
	}

	columes := make([][]string, 0)
	for _, ip := range ips {
		groups := pattern.FindStringSubmatch(ip)
		if pattern.MatchString(ip) {
			columes = append(columes, []string{groups[2], groups[3], groups[5], groups[6], groups[8]})
		} else {
			fmt.Println("hostname: ", ip)
		}
	}

	out := bytes.NewBufferString("")
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"user", "password", "start", "end", "port"})
	table.SetBorder(true)
	table.SetCenterSeparator("*")
	table.SetRowSeparator("-")
	table.AppendBulk(columes)
	table.Render()
	fmt.Println(out.String())
}

func TestAdd(t *testing.T) {
	cfg := SSH{
		Password: "", PkFile: "$HOME/.ssh/id_rsa", Port: 22,
	}
	nodes := Add(cfg, "root:$HOME/.ssh/id_rsa@172.16.100.10-172.16.100.15:22")
	for _, node := range nodes {
		t.Log(node.Hostname)
	}
}
