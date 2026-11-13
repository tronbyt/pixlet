package starlarkhttp

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPublicAddr(t *testing.T) {
	tests := map[string]bool{
		"1.1.1.1":          true,
		"8.8.8.8":          true,
		"2606:4700::1111":  true,
		"127.0.0.1":        false,
		"10.0.0.1":         false,
		"172.16.0.1":       false,
		"192.168.1.1":      false,
		"169.254.169.254":  false,
		"100.64.0.1":       false,
		"0.0.0.0":          false,
		"255.255.255.255":  false,
		"224.0.0.1":        false,
		"::1":              false,
		"::":               false,
		"fd00::1":          false,
		"fe80::1":          false,
		"::ffff:127.0.0.1": false,
		"::ffff:1.1.1.1":   true,
		"192.0.2.1":        false,
		"198.51.100.1":     false,
		"203.0.113.1":      false,
		"2001:db8::1":      false,
		"fec0::1":          false,
		// NAT64 and 6to4 are judged by the embedded IPv4 address.
		"64:ff9b::c0a8:101":     false,
		"64:ff9b:1:a00:0:100::": false,
		"2002:c0a8:101::1":      false,
		"64:ff9b::808:808":      true,
		"2002:808:808::1":       true,
		"198.51.101.1":          true,
		"203.0.114.1":           true,
		"2001:db9::1":           true,
	}

	for addr, want := range tests {
		t.Run(addr, func(t *testing.T) {
			assert.Equal(t, want, IsPublicAddr(netip.MustParseAddr(addr)))
		})
	}
}

func TestDialControl(t *testing.T) {
	require.NoError(t, dialControl("tcp", "192.168.1.1:80", nil))

	BlockPrivateIPs.Store(true)
	t.Cleanup(func() { BlockPrivateIPs.Store(false) })
	require.ErrorIs(t, dialControl("tcp", "192.168.1.1:80", nil), ErrPrivateIP)
	require.ErrorIs(t, dialControl("tcp6", "[::1]:443", nil), ErrPrivateIP)
	assert.NoError(t, dialControl("tcp", "1.1.1.1:443", nil))
}
