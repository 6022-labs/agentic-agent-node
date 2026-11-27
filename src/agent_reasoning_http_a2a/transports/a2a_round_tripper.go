package transports

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/settings"
)

var ErrA2AAddressRefused = errors.New("a2a calls refuse a non-public address")

// ErrA2AInsecureScheme refuses plain http, which would hand PAYMENT-SIGNATURE to anyone on the path.
var ErrA2AInsecureScheme = errors.New("a2a calls refuse plain http")

// nonPublicPrefixes are ranges the netip predicates miss, including v6 forms that can embed a private v4.
var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fec0::/10"),
}

// A2ARoundTripper calls peers over https on public addresses only, unless private networks are allowed.
type A2ARoundTripper struct {
	base                 *http.Transport
	allowPrivateNetworks bool
}

func NewA2ARoundTripper(a2aSettings *settings.A2ASettings) *A2ARoundTripper {
	t := &A2ARoundTripper{allowPrivateNetworks: a2aSettings.AllowPrivateNetworks}
	dialer := &net.Dialer{Timeout: time.Duration(a2aSettings.RequestTimeoutSeconds) * time.Second}
	if !a2aSettings.AllowPrivateNetworks {
		dialer.Control = t.refuseNonPublic
	}
	t.base = http.DefaultTransport.(*http.Transport).Clone()
	t.base.DialContext = dialer.DialContext
	// A proxy would be dialed instead of the peer, bypassing the address check.
	t.base.Proxy = nil
	return t
}

func (t *A2ARoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Plain http stays for agents sharing a local network, the only place it cannot be read in transit.
	if !t.allowPrivateNetworks && req.URL.Scheme != "https" {
		return nil, fmt.Errorf("%w: %s", ErrA2AInsecureScheme, req.URL.Host)
	}
	return t.base.RoundTrip(req)
}

// refuseNonPublic checks the address actually dialed, after DNS, so a rebinding hostname cannot slip through.
func (t *A2ARoundTripper) refuseNonPublic(_ string, address string, _ syscall.RawConn) error {
	addressPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrA2AAddressRefused, address)
	}
	if ip := addressPort.Addr(); !t.isPublic(ip) {
		return fmt.Errorf("%w: %s", ErrA2AAddressRefused, ip.Unmap())
	}
	return nil
}

func (t *A2ARoundTripper) isPublic(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() || ip.IsMulticast() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() {
		return false
	}
	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}
