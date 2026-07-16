package realip_zoning

import (
	"context"
	"errors"
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
		return nil, mkIfMultiError(err, "when gathering CIDRs for TrustedProxies")
	}

	return &proxyConf_t{
		useHeader: config.TrustedProxies.UseHeader,
		cidrs:     cidrs,
	}, nil
}

func toInvertedZones(ctx context.Context, zones []Zone) (map[netip.Prefix]*headerSet, error) {

	syncer := newCollector[Zipped2[netip.Prefix, *headerSet]](len(zones))
	syncer.wg.Add(len(zones))
	for i, zone := range zones {
		go mkInvertedZone(&syncer, ctx, zone, i)
	}

	policyList, errs := syncer.collect()
	if err := errors.Join(errs...); err != nil {
		return nil, mkIfMultiError(err, "when gathering CIDRs for zones")
	}

	rtv := make(map[netip.Prefix]*headerSet)
	for _, tuple := range policyList {
		rtv[tuple.el01] = tuple.el02
	}
	return rtv, nil
}

func mkInvertedZone(
	syncer *fetchCollector[Zipped2[netip.Prefix, *headerSet]], ctx context.Context,
	zone Zone, zoneIdx int,
) {
	defer syncer.wg.Done()

	cidrs, err := zone.IPs.getCIDRs()
	if err != nil {
		syncer.chErr <- mkIfMultiError(err, fmt.Sprintf("when gathering CIDRs for zone %d", zoneIdx))
		return
	}

	headers := &zone.AttachHeaders
	for _, cidr := range cidrs {
		tuple := Zipped2[netip.Prefix, *headerSet]{cidr, headers}
		syncer.chRtv <- tuple
	}
}
