package kube

import (
	"fmt"
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/volumes"
	"regexp"
	"strconv"
	"strings"
)

type (
	array []string

	port struct {
		name          string
		hostIP        string
		hostPort      string
		containerPort string
		protocol      string
	}
	device struct {
		name string
		path string
	}
	container struct {
		name            string
		image           string
		imagePullPolicy string
		command         array
		args            array
		ports           []*port
		envs            []*environment
		res             *resources
		volumeMounts    volumes.VolumeMounts
		devices         []*device
	}
)

var (
	port_pattern = regexp.MustCompile(`(((\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):)?(\d{1,5}):)?(\d{1,5})(\/(\S+))?`)
)

func (p port) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("- name: ", p.name)
	w.Tab().Line("containerPort: ", p.containerPort)
	if p.hostIP != "" {
		w.Tab().Line("hostIP: ", p.hostIP)
	}
	if p.hostPort != "" {
		w.Tab().Line("hostPort: ", p.hostPort)
	}
	if p.protocol != "" {
		w.Tab().Line("protocol: ", p.protocol)
	}
	return w.String()
}

func (ary array) ToYaml(kind string, indent int) string {
	//只有数字/bool的话需要添加 "
	for i, s := range ary {
		if s == "true" || s == "false" {
			ary[i] = "\"" + ary[i] + "\""
		} else if _, err := strconv.ParseFloat(s, 64); err == nil {
			ary[i] = "\"" + ary[i] + "\""
		}
	}

	w := config.Writer(indent)
	if len(ary) == 0 {
	} else if len(ary) <= 3 {
		w.Line(fmt.Sprintf("%s:", kind), "[", strings.Join(ary, ", "), "]")
	} else {
		w.Line(fmt.Sprintf("%s:", kind))
		for _, command := range ary {
			w.Tab().Line("-", command)
		}
	}
	return w.String()
}

func (c *container) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Line("- name:", c.name)
	w.Tab().Line("image:", c.image)
	if c.imagePullPolicy != "" {
		w.Tab().Line("imagePullPolicy:", c.imagePullPolicy)
	}
	w.Writer(c.command.ToYaml("command", indent+1))
	w.Writer(c.args.ToYaml("args", indent+1))

	if len(c.ports) > 0 {
		w.Tab().Line("ports:")
		for _, p := range c.ports {
			w.Writer(p.ToYaml(indent + 2))
		}
	}

	if len(c.envs) > 0 {
		w.Tab().Line("env:")
		for _, env := range c.envs {
			w.Writer(env.ToYaml(indent + 2))
		}
	}

	if c.res != nil {
		w.Writer(c.res.ToYaml(indent + 1))
	}

	if len(c.devices) > 0 {
		w.Indent(1).Line("volumeDevices:")
		for _, device := range c.devices {
			w.Indent(2).Line("- devicePath:", device.path)
			w.Indent(2).Line("  name:", device.name)
		}
	}

	if len(c.volumeMounts) > 0 {
		w.Writer(c.volumeMounts.ToYaml(indent + 1))
	}
	return w.String()
}

func portParse(name, portString string) *port {
	groups := port_pattern.FindStringSubmatch(portString)
	return &port{
		name:          name,
		containerPort: groups[5],
		hostIP:        groups[3],
		hostPort:      groups[4],
		protocol:      strings.ToUpper(groups[7]),
	}
}

func containerParse(d *config.Directive) (*container, volumes.Volumes) {
	asserts.ArgsRange(d, 2, 3)
	c := &container{
		name: d.Args[0], image: d.Args[1],
	}
	volumesDef := volumes.Volumes{}
	if len(d.Args) == 3 {
		c.imagePullPolicy = d.Args[2]
	}
	for _, body := range d.Body {
		switch body.Name {
		case "command":
			c.command = body.Args
		case "args":
			c.args = body.Args

		case "port":
			asserts.ArgsLen(body, 2)
			c.ports = append(c.ports, portParse(body.Args[0], body.Args[1]))
		case "ports":
			asserts.ArgsLen(body, 0)
			for _, p := range body.Body {
				asserts.ArgsLen(p, 1)
				c.ports = append(c.ports, portParse(p.Name, p.Args[0]))
			}
		case "env":
			asserts.ArgsMin(body, 2)
			c.envs = append(c.envs, NewEnv(body.Args[0], body.Args[1:]))
		case "envs":
			asserts.ArgsLen(body, 0)
			for _, env := range body.Body {
				c.envs = append(c.envs, NewEnv(env.Name, env.Args))
			}
		case "resources":
			asserts.ArgsLen(body, 0)
			c.res = resourceParse(body)
		case "device":
			asserts.ArgsLen(body, 2)
			c.devices = append(c.devices, &device{
				name: body.Args[0], path: body.Args[1],
			})
		case "mount":
			mountVolume, volumeDef := volumes.MountParse(body.Args, body.Body)
			c.volumeMounts = append(c.volumeMounts, mountVolume)
			if volumeDef != nil {
				volumesDef = append(volumesDef, volumeDef)
			}
		case "mounts":
			for _, directive := range body.Body {
				args := append([]string{directive.Name}, directive.Args...)
				mountVolume, volumeDef := volumes.MountParse(args, directive.Body)
				c.volumeMounts = append(c.volumeMounts, mountVolume)
				if volumeDef != nil {
					volumesDef = append(volumesDef, volumeDef)
				}
			}
		}
	}
	return c, volumesDef
}
