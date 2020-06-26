package kube

import "github.com/ihaiker/vik8s/reduce/config"

type Properties struct {
	Kind  string
	Attrs map[string]string
}

func NewProperties(kind string) *Properties {
	return &Properties{
		Kind:  kind,
		Attrs: make(map[string]string),
	}
}
func Attributes() *Properties {
	return NewProperties("")
}
func Annotations() *Properties {
	return NewProperties("annotations")
}

func Labels() *Properties {
	return NewProperties("labels")
}

func (attrs *Properties) Add(label, value string) {
	attrs.Attrs[label] = value
}

func (props *Properties) Has() bool {
	return len(props.Attrs) > 0
}

func (props *Properties) ToYaml(indent int) string {
	if !props.Has() {
		return ""
	}
	add := 0
	outs := config.Writer(indent)
	if props.Kind != "" {
		add = 1
		outs.Format("%s:", props.Kind).Enter()
	}
	for k, v := range props.Attrs {
		outs.Indent(add).Format("%s: %s", k, v).Enter()
	}
	return outs.String()
}

func (an *Properties) String() string {
	return an.ToYaml(0)
}
