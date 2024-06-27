package utils

import (
	"net"
	"testing"

	"github.com/c-robinson/iplib"
)

var CampusIPs = []string{
	"10.10.10.0/8",
	"192.168.0.0/16",
	"172.16.0.0/12",
	"49.52.0.0/19",
	"58.198.176.0/20",
	"59.78.176.0/20",
	"59.78.192.0/21",
	"202.120.80.0/20",
	"202.127.250.0/24",
	"219.228.56.0/21",
	"219.228.128.0/20",
	"219.228.144.0/21",
	"222.204.232.0/21",
	"222.204.240.0/20",
}

func Test_IpCheck(t *testing.T) {
	var ipCheckList = []struct {
		ip  string
		exp bool
	}{
		{"10.10.3.1", true},
		{"192.168.12.3", true},
		{"172.20.3.10", true},
		{"49.52.10.13", true},
		{"58.198.180.1", true},
		{"59.78.180.12", true},
		{"59.78.193.12", true},
		{"202.120.95.15", true},
		{"202.127.250.123", true},
		{"219.228.58.0", true},
		{"219.228.133.3", true},
		{"219.228.150.1", true},
		{"222.204.234.12", true},
		{"222.204.244.12", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
	}

	for _, r := range ipCheckList {
		out := IPCheck(r.ip, CampusIPs)
		if r.exp != out {
			t.Errorf("IsIllegal_character of %s expects %t, but got %t ", r.ip, r.exp, out)
		}
	}
}

func Test_IsValidCIDR(t *testing.T) {
	input := "192.168.0.1/24"
	isValid, _ := IsValidCIDR(input)
	t.Log(isValid)
	input = "192.168.0.1/36"
	isValid, _ = IsValidCIDR(input)
	t.Log(isValid)
	input = "192.168.0.256/24"
	isValid, _ = IsValidCIDR(input)
	t.Log(isValid)
}

func Test_GetStartAndEndByCIDR(t *testing.T) {
	input := "192.168.0.0/24"
	start, end, err := GetStartAndEndByCIDR(input)
	t.Log(input)
	t.Log(start)
	t.Log(end)
	input = "10.1.0.0/16"
	start, end, err = GetStartAndEndByCIDR(input)
	t.Log(input)
	t.Log(start)
	t.Log(end)
	if err != nil {
		t.Error(err)
	}
}

func Test_GetCIDRByIpRange(t *testing.T) {
	start := iplib.IP4ToUint32(net.ParseIP("10.1.0.0"))
	end := iplib.IP4ToUint32(net.ParseIP("10.255.255.255"))
	networks := GetCIDRByIpRange(start, end)
	for _, network := range networks {
		t.Log(network)
	}
}

func Test_SplitCIDR(t *testing.T) {
	network := "10.0.0.0/8"
	containNetwork := "10.20.1.128/16"
	networks, err := SplitCIDR(network, containNetwork)
	for _, result := range networks {
		t.Log(result)
	}
	if err != nil {
		t.Error(err)
	}
}

func Test_MergeCIDR(t *testing.T) {
	networks1 := []string{"10.0.0.0/8"}
	networks2 := []string{"10.0.0.0/9", "10.0.0.0/8"}
	networks3 := []string{"192.168.0.0/16", "10.0.0.0/9", "10.0.0.0/8"}
	networks4 := []string{"192.168.0.0/16", "10.0.0.0/9", "11.0.0.0/8", "10.0.0.0/8"}
	networks5 := []string{"192.168.0.0/16", "10.0.0.0/9", "11.0.0.0/8", "10.0.0.0/8", "192.0.0.0/8"}
	networks6 := []string{"192.168.0.0/16", "10.0.0.0/9", "11.0.0.0/8", "10.0.0.0/8", "192.0.0.0/8", "193.0.0.0/8",
		"196.0.0.0/6", "200.0.0.0/5", "208.0.0.0/4", "224.0.0.0/3"}
	networks7 := []string{"192.168.0.0/16", "10.0.0.0/9", "11.0.0.0/8", "10.0.0.0/8", "192.0.0.0/8", "193.0.0.0/8",
		"194.0.0.0/7", "196.0.0.0/6", "200.0.0.0/5", "208.0.0.0/4", "224.0.0.0/3"}
	networks := [][]string{networks1, networks2, networks3, networks4, networks5, networks6, networks7}
	for _, network := range networks {
		mergedNetworks, err := MergeCIDR(network)
		for _, result := range mergedNetworks {
			t.Log(result)
		}
		if err != nil {
			t.Error(err)
		}
		t.Log("=================")
	}

}
