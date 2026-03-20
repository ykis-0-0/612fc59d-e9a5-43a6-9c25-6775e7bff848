package realip_zoning

import (
	"net/netip"
)

type headerSet = map[string]string

type proxyConf_t = struct {
	useHeader string
	cidrs     []netip.Prefix
}

type Zone struct {
	IPs           IPSources `json:"ips"`
	AttachHeaders headerSet `json:"attachHeaders"`
}
