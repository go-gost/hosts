package gost

import (
	"net"
	"testing"
)

var hostsLookupTests = []struct {
	hosts []Host
	host  string
	ip    net.IP
}{
	{nil, "", nil},
	{nil, "example.com", nil},
	{[]Host{}, "", nil},
	{[]Host{}, "example.com", nil},
	{[]Host{NewHost(nil, "")}, "", nil},
	{[]Host{NewHost(nil, "example.com")}, "example.com", nil},
	{[]Host{NewHost(net.IPv4(192, 168, 1, 1), "")}, "", nil},
	{[]Host{NewHost(net.IPv4(192, 168, 1, 1), "example.com")}, "example.com", net.IPv4(192, 168, 1, 1)},
	{[]Host{NewHost(net.IPv4(192, 168, 1, 1), "example.com")}, "example", nil},
	{[]Host{NewHost(net.IPv4(192, 168, 1, 1), "example.com", "example", "examples")}, "example", net.IPv4(192, 168, 1, 1)},
	{[]Host{NewHost(net.IPv4(192, 168, 1, 1), "example.com", "example", "examples")}, "examples", net.IPv4(192, 168, 1, 1)},
}

func TestHostsLookup(t *testing.T) {
	for i, tc := range hostsLookupTests {
		hosts := NewHosts(tc.hosts...)
		ip := hosts.Lookup(tc.host)
		if !ip.Equal(tc.ip) {
			t.Errorf("#%d test failed: lookup should be %s, got %s", i, tc.ip, ip)
		}
	}
}
