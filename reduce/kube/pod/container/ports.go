package container

import (
	v1 "k8s.io/api/core/v1"
	"regexp"
	"strconv"
	"strings"
)

var (
	//ip:hostPort:containerPort/protocol
	portPattern = regexp.MustCompile(`(((\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):)?(\d{1,5}):)?(\d{1,5})(\/(\S+))?`)
)

func portParse(name, portString string) v1.ContainerPort {
	groups := portPattern.FindStringSubmatch(portString)
	containerPort, _ := strconv.Atoi(groups[5])
	hostPort, _ := strconv.Atoi(groups[4])
	return v1.ContainerPort{
		Name: name, Protocol: v1.Protocol(strings.ToUpper(groups[7])),
		ContainerPort: int32(containerPort),
		HostPort:      int32(hostPort), HostIP: groups[3],
	}
}
