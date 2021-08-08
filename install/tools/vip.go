package tools

import (
	"github.com/ihaiker/vik8s/libs/utils"
	"net"
)

type VipFor int

const (
	Vik8sApiServer = iota + 1
	Vik8sCalicoETCD
)

func GetVip(cidr string, vipFor VipFor) string {
	_, ipnet, err := net.ParseCIDR(cidr)
	utils.Panic(err, "invalid %s cidr", cidr)

	_, max := utils.AddressRange(ipnet)
	for i := 0; i < int(vipFor); i++ {
		max = utils.PrevIP(max)
	}
	return max.String()
}

func AddressRange(cidr string) (net.IP, net.IP) {
	_, ipnet, err := net.ParseCIDR(cidr)
	utils.Panic(err, "invalid %s cidr", cidr)
	return utils.AddressRange(ipnet)
}
