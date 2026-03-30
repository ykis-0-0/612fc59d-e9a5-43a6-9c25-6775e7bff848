package realip_zoning

import (
	"context"
	"fmt"

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

func mkProxyConf(ctx context.Context, config *Config) (*proxyConf_t, error) {
	if config.TrustedProxies == nil {
		return nil, nil
	}

	cidrs, err := config.TrustedProxies.getCIDRs()
	if err != nil {
		type multiError_t = interface {
			error
			Unwrap() []error
		}

		var spacer string
		if _, ok := err.(multiError_t); ok {
			spacer = " []error =>\n"
		} else {
			spacer = ""
		}

		return nil, fmt.Errorf("when gathering CIDRs for TrustedProxies:%s%w", spacer, err)
	}

	return &proxyConf_t{
		useHeader: config.TrustedProxies.UseHeader,
		cidrs:     cidrs,
	}, nil
}
