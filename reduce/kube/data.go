package kube

import "github.com/ihaiker/vik8s/reduce/config"

type (
	Data map[string]string
)

func (an *Data) Has() bool {
	return len(*an) > 0
}

func (data *Data) ToYaml(indent int) string {
	outs := config.Writer(indent)
	outs.Line("data:")
	for k, v := range *data {
		if v[0] == '\'' {
			v = "|" + v[1:len(v)-1]
		}
		outs.Tab().Format("%s: %s", k, v).Enter()
	}
	return outs.String()
}
func (an *Data) String() string {
	return an.ToYaml(0)
}
