package utils

import (
	"encoding/binary"
	"errors"
	"net"
	"sort"
	"strconv"

	"strings"

	"github.com/c-robinson/iplib"
)

// IPCheck 检查 ip 是否在特定的 ip 地址范围内
func IPCheck(thisip string, ips []string) bool {
	for _, ip := range ips {
		ip = strings.TrimRight(ip, "/")
		if strings.Contains(ip, "/") {
			if ipCheckMask(thisip, ip) {
				return true
			}
		} else if strings.Contains(ip, "-") {
			ipRange := strings.SplitN(ip, "-", 2)
			if ipCheckRange(thisip, ipRange[0], ipRange[1]) {
				return true
			}
		} else {
			if thisip == ip {
				return true
			}
		}
	}
	return false
}

func ipCheckRange(ip, ipStart, ipEnd string) bool {
	thisIP := net.ParseIP(ip)
	firstIP := net.ParseIP(ipStart)
	endIP := net.ParseIP(ipEnd)
	if thisIP.To4() == nil || firstIP.To4() == nil || endIP.To4() == nil {
		return false
	}
	firstIPNum := ipToInt(firstIP.To4())
	endIPNum := ipToInt(endIP.To4())
	thisIPNum := ipToInt(thisIP.To4())
	if thisIPNum >= firstIPNum && thisIPNum <= endIPNum {
		return true
	}
	return false
}

func ipCheckMask(ip, ipMask string) bool {
	_, subnet, _ := net.ParseCIDR(ipMask)

	thisIP := net.ParseIP(ip)
	return subnet.Contains(thisIP)
}

func ipToInt(ip net.IP) int32 {
	return int32(binary.BigEndian.Uint32(ip.To4()))
}

//检查一个CIDR（形如"192.168.0.1/24"）是否合法
func IsValidCIDR(network string) (isValid bool, err error) {
	_, _, err = iplib.ParseCIDR(network)
	if err != nil {
		isValid = false
	} else {
		isValid = true
	}
	return
}

func GetStartAndEndByCIDR(network string) (start uint32, end uint32, err error) {
	_, ipNet, err := iplib.ParseCIDR(network)
	if err != nil {
		return
	}
	start = iplib.IP4ToUint32(ipNet.FirstAddress())
	end = iplib.IP4ToUint32(ipNet.LastAddress())
	return
}

func GetCIDRByIpRange(start uint32, end uint32) (networks []string) {
	networks = getCIDRByIpRange(start, end, []string{})
	return
}

func getCIDRByIpRange(start uint32, end uint32, networks []string) []string {
	bits := 1
	var mask uint32 = 1
	for {
		if bits >= 32 {
			break
		}
		newip := start | mask
		if newip > end || ((start>>bits)<<bits) != start {
			bits = bits - 1
			mask = mask >> 1
			break
		}
		bits = bits + 1
		mask = (mask << 1) + 1
	}
	newip := start | mask
	bits = 32 - bits
	network := iplib.Uint32ToIP4(start).String() + "/" + strconv.Itoa(bits)
	networks = append(networks, network)
	if newip < end {
		return getCIDRByIpRange(newip+1, end, networks)
	}
	return networks
}

func SplitCIDR(network string, containNetwork string) (networks []string, err error) {
	_, ipna, err := iplib.ParseCIDR(network)
	if err != nil {
		return
	}
	_, ipnb, err := iplib.ParseCIDR(containNetwork)
	if err != nil {
		return
	}
	if !(ipna.ContainsNet(ipnb)) {
		err = errors.New("包含关系不存在")
		return
	}
	startA, endA, err := GetStartAndEndByCIDR(network)
	if err != nil {
		return
	}
	startB, endB, err := GetStartAndEndByCIDR(containNetwork)
	if err != nil {
		return
	}
	var leftNetworks []string
	var rightNetworks []string
	if startA < startB {
		leftNetworks = GetCIDRByIpRange(startA, startB-1)
	}
	if endA > endB {
		rightNetworks = GetCIDRByIpRange(endB+1, endA)
	}
	networks = append(leftNetworks, ipnb.String())
	networks = append(networks, rightNetworks...)
	return
}

type Block struct {
	start uint32
	end   uint32
}

func MergeCIDR(networks []string) (mergedNetworks []string, err error) {
	if len(networks) <= 1 {
		mergedNetworks = networks
		return
	}
	var ipNets []iplib.Net
	for _, network := range networks {
		_, ipNet, err0 := iplib.ParseCIDR(network)
		if err0 != nil {
			err = err0
			return
		}
		ipNets = append(ipNets, ipNet)
	}
	sort.Sort(iplib.ByNet(ipNets))

	var blocks []Block
	minStart, maxEnd, err := GetStartAndEndByCIDR(ipNets[0].String())
	if err != nil {
		return
	}
	for i, ipNet := range ipNets {
		split := false
		if i == 0 {
			continue
		}
		start, end, err0 := GetStartAndEndByCIDR(ipNet.String())
		if err0 != nil {
			err = err0
			return
		}
		if start > maxEnd+1 {
			block := Block{minStart, maxEnd}
			blocks = append(blocks, block)
			minStart = start
			maxEnd = end
		}
		if end > maxEnd {
			maxEnd = end
		}
		if i == len(ipNets)-1 && !split {
			block := Block{minStart, maxEnd}
			blocks = append(blocks, block)
			break
		}
	}

	for _, b := range blocks {
		mergedNetwork := GetCIDRByIpRange(b.start, b.end)
		mergedNetworks = append(mergedNetworks, mergedNetwork...)
	}
	return
}
