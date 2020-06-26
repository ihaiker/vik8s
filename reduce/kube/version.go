package kube

type Version struct {
	Kubernetes string
}

func (v Version) Get(kind string) string {
	switch kind {
	case "Pod":
		return "v1"
	case "DaemonSet":
		return "v1"
	case "Deployment":
		return "apps/v1"
	}
	return "v1"
}
