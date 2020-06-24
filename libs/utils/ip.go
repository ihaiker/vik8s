package utils

import (
	"fmt"
	"math/big"
	"net"
	"strings"
)

func RangeIP(start, end string) []string {
	from := net.ParseIP(start).To4()
	to := net.ParseIP(end).To4()
	Assert(from != nil, "Invalid IP %s", start)
	Assert(to != nil, "Invalid IP %s", end)

	ips := make([]string, 0)
	for ComposeIP(from, to) <= 0 {
		ips = append(ips, from.String())
		from = NextIP(from)
	}
	return ips
}

func ParseIPS(nodes []string) []string {
	outs := make([]string, 0)
	for _, node := range nodes {
		if strings.Contains(node, "-") { // ip-ip
			startAndEnd := strings.SplitN(node, "-", 2)
			outs = append(outs, RangeIP(startAndEnd[0], startAndEnd[1])...)
		} else {
			outs = append(outs, node)
		}
	}
	return outs
}

func ComposeIP(a, b net.IP) int {
	aa, _ := ipToInt(a)
	bb, _ := ipToInt(b)
	return aa.Cmp(bb)
}

func NextIP(ip net.IP) net.IP {
	a := big.NewInt(0).SetBytes(ip)
	b := a.Add(a, big.NewInt(1))
	return net.IP(b.Bytes())
}

func PrevIP(ip net.IP) net.IP {
	a := big.NewInt(0).SetBytes(ip)
	b := a.Sub(a, big.NewInt(1))
	return net.IP(b.Bytes())
}

// AddressRange returns the first and last addresses in the given CIDR range.
func AddressRange(network *net.IPNet) (net.IP, net.IP) {
	// the first IP is easy
	firstIP := network.IP

	// the last IP is the network address OR NOT the mask address
	prefixLen, bits := network.Mask.Size()
	if prefixLen == bits {
		// Easy!
		// But make sure that our two slices are distinct, since they
		// would be in all other cases.
		lastIP := make([]byte, len(firstIP))
		copy(lastIP, firstIP)
		return firstIP, lastIP
	}

	firstIPInt, bits := ipToInt(firstIP)
	hostLen := uint(bits) - uint(prefixLen)
	lastIPInt := big.NewInt(1)
	lastIPInt.Lsh(lastIPInt, hostLen)
	lastIPInt.Sub(lastIPInt, big.NewInt(1))
	lastIPInt.Or(lastIPInt, firstIPInt)

	return firstIP, intToIP(lastIPInt, bits)
}

func ipToInt(ip net.IP) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes([]byte(ip))
	if len(ip) == net.IPv4len {
		return val, 32
	} else if len(ip) == net.IPv6len {
		return val, 128
	} else {
		panic(fmt.Errorf("Unsupported address length %d", len(ip)))
	}
}

func intToIP(ipInt *big.Int, bits int) net.IP {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)
	// Pack our IP bytes into the end of the return array,
	// since big.Int.Bytes() removes front zero padding.
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}
	return net.IP(ret)
}
