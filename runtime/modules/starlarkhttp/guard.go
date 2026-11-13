package starlarkhttp

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"

	"go.uber.org/atomic"
)

// BlockPrivateIPs controls whether applets are prevented from connecting to
// private, loopback, link-local, or otherwise non-public addresses.
var BlockPrivateIPs = atomic.NewBool(false)

var ErrPrivateIP = errors.New("connections to private IP addresses are not allowed")

// Non-public ranges that netip.Addr's helper methods do not cover.
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // "this" network
	netip.MustParsePrefix("100.64.0.0/10"),   // carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),    // documentation
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
	netip.MustParsePrefix("198.51.100.0/24"), // documentation
	netip.MustParsePrefix("203.0.113.0/24"),  // documentation
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved, including broadcast
	netip.MustParsePrefix("64:ff9b:1::/48"),  // local-use NAT64
	netip.MustParsePrefix("2001:db8::/32"),   // documentation
	netip.MustParsePrefix("fec0::/10"),       // deprecated site-local
}

// IPv6 ranges that embed an IPv4 address, which is checked instead.
var (
	nat64Prefix     = netip.MustParsePrefix("64:ff9b::/96")
	sixToFourPrefix = netip.MustParsePrefix("2002::/16")
)

// NewTransport returns an http.Transport that refuses to dial non-public
// addresses when BlockPrivateIPs is set. The check runs against the resolved
// IP at connect time, so it also covers redirects and DNS names that resolve
// to private addresses.
func NewTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   dialControl,
	}
	t.DialContext = dialer.DialContext
	return t
}

func dialControl(_, address string, _ syscall.RawConn) error {
	if !BlockPrivateIPs.Load() {
		return nil
	}

	addrPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return fmt.Errorf("parsing dial address %q: %w", address, err)
	}

	if !IsPublicAddr(addrPort.Addr()) {
		return ErrPrivateIP
	}
	return nil
}

// IsPublicAddr reports whether addr is a publicly routable address.
func IsPublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsValid() ||
		addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified() {
		return false
	}
	for _, p := range blockedPrefixes {
		if p.Contains(addr) {
			return false
		}
	}

	b := addr.As16()
	switch {
	case nat64Prefix.Contains(addr):
		return IsPublicAddr(netip.AddrFrom4([4]byte(b[12:16])))
	case sixToFourPrefix.Contains(addr):
		return IsPublicAddr(netip.AddrFrom4([4]byte(b[2:6])))
	}
	return true
}
