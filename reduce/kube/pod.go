package kube

import (
	"github.com/ihaiker/vik8s/reduce/asserts"
	"github.com/ihaiker/vik8s/reduce/config"
	"github.com/ihaiker/vik8s/reduce/kube/volumes"
	"strings"
)

type (
	Pod struct {
		*Entry
		*Properties
		nodeSelector   *Properties
		initContainers []*container
		containers     []*container
		volumes        volumes.Volumes
	}
)

func (c *Pod) ToYaml(indent int) string {
	w := config.Writer(indent)
	w.Writer(c.Entry.Yaml(indent))

	w.Line("spec:")
	w.Writer(c.Properties.ToYaml(indent + 1))
	if c.nodeSelector != nil {
		w.Writer(c.nodeSelector.ToYaml(indent + 1))
	}

	if len(c.initContainers) > 0 {
		w.Tab().Line("initContainers:")
		for _, initContainer := range c.initContainers {
			w.Writer(initContainer.ToYaml(indent + 2))
		}
	}

	w.Tab().Line("containers:")
	for _, container := range c.containers {
		w.Writer(container.ToYaml(indent + 2))
	}

	w.Writer(c.volumes.ToYaml(indent + 1))
	return w.String()
}

func podParse(d *config.Directive, kube *Kubernetes) {
	asserts.ArgsMin(d, 1)
	pod := &Pod{
		Entry: &Entry{
			Name:   d.Args[0],
			Labels: Labels(), Annotations: Annotations(),
		},
		Properties: Attributes(),
		volumes:    volumes.Volumes{},
	}
	entry(d, pod.Entry, func(body *config.Directive) {
		if podBodyParse(body, pod) {
			asserts.ArgsLen(body, 1)
			pod.Properties.Add(body.Name, strings.Join(body.Args, " "))
		}
	})
	kube.Add(pod)
}

func podBodyParse(d *config.Directive, pod *Pod) (escape bool) {
	switch d.Name {
	case "container":
		c, vs := containerParse(d)
		pod.containers = append(pod.containers, c)
		for _, v := range vs {
			pod.volumes.Add(v)
		}
	case "init_container":
		c, vs := containerParse(d)
		pod.initContainers = append(pod.initContainers, c)
		for _, v := range vs {
			pod.volumes.Add(v)
		}

	case "nodeSelector":
		pod.nodeSelector = NewProperties("nodeSelector")
		bodyLabels(pod.nodeSelector, d)

	case "volume":
		asserts.ArgsMin(d, 1)
		pod.volumes = append(pod.volumes, volumes.VolumeParse(d.Args, d.Body))
	case "volumes":
		for _, subDir := range d.Body {
			args := append([]string{subDir.Name}, subDir.Args...)
			pod.volumes = append(pod.volumes, volumes.VolumeParse(args, subDir.Body))
		}
	default:
		escape = true
	}
	return
}

func init() {
	kinds["pod"] = podParse
	kinds["Pod"] = podParse
}
