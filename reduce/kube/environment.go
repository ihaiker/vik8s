package kube

import (
	"fmt"
	"github.com/ihaiker/vik8s/reduce/config"
)

type (
	environment struct {
		name  string
		from  string
		value []string
	}
)

func (e environment) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("- name:", e.name)
	switch e.from {
	case "":
		w.Tab().Line("value:", e.value[0])
	case "field":
		w.Indent(1).Line("valueFrom:")
		w.Indent(2).Line("fieldRef:")
		w.Indent(3).Line("fieldPath:", e.value[0])
	case "configMap", "config":
		w.Indent(1).Line("valueFrom:")
		w.Indent(2).Line("configMapKeyRef:")
		w.Indent(3).Line("name:", e.value[0])
		w.Indent(3).Line("key:", e.value[1])
	case "secret":
		w.Indent(1).Line("valueFrom:")
		w.Indent(2).Line("secretKeyRef:")
		w.Indent(3).Line("name:", e.value[0])
		w.Indent(3).Line("key:", e.value[1])
	case "resource", "res":
		w.Indent(1).Line("valueFrom:")
		w.Indent(2).Line("resourceFieldRef:")
		w.Indent(3).Line("containerName:", e.value[0])
		w.Indent(3).Line("resource:", e.value[1])
	}
	return w.String()
}

func NewEnv(name string, args []string) *environment {
	env := &environment{
		name: name,
	}
	if len(args) == 1 {
		env.value = args
	} else {
		env.from = args[0]
		if !(env.from == "field" ||
			env.from == "configMap" || env.from == "config" ||
			env.from == "secret" ||
			env.from == "resource" || env.from == "res") {
			panic(fmt.Sprintf("env not support: %s", env.from))
		}
		env.value = args[1:]
	}
	return env
}
