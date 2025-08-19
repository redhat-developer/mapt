package network

import (
	"net"

	"github.com/redhat-developer/mapt/pkg/util"
)

func RandomIp(cidr string) (*string, error) {
	cidrIps, err := ips(cidr)
	if err != nil {
		return nil, err
	}
	rIp := util.RandomItemFromArray(cidrIps)
	return &rIp, nil
}

func ips(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
