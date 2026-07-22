package realip_zoning

import (
	"context"
	"errors"
	"fmt"
)

type headerSet = map[string]string

type proxyConf_t = struct {
	useHeader string
	ipRanges  []ipRange
}

type Zone struct {
	IPs           IPSources `json:"ips"`
	AttachHeaders headerSet `json:"attachHeaders"`
}

func mkProxyConf(ctx context.Context, config *Config) (*proxyConf_t, error) {
	if config.TrustedProxies == nil {
		return nil, nil
	}

	ranges, err := config.TrustedProxies.intoIpRanges()
	if err != nil {
		return nil, mkIfMultiError(err, "when gathering IP ranges for TrustedProxies")
	}

	return &proxyConf_t{
		useHeader: config.TrustedProxies.UseHeader,
		ipRanges:  ranges,
	}, nil
}

func toInvertedZones(ctx context.Context, zones []Zone) (map[ipRange]*headerSet, error) {

	syncer := newCollector[Zipped2[ipRange, *headerSet]](len(zones))
	syncer.wg.Add(len(zones))
	for i, zone := range zones {
		go mkInvertedZone(&syncer, ctx, zone, i)
	}

	policyList, errs := syncer.collect()
	if err := errors.Join(errs...); err != nil {
		return nil, mkIfMultiError(err, "when building zone configurations")
	}

	rtv := make(map[ipRange]*headerSet)
	for _, tuple := range policyList {
		rtv[tuple.el01] = tuple.el02
	}
	return rtv, nil
}

func mkInvertedZone(
	syncer *fetchCollector[Zipped2[ipRange, *headerSet]], ctx context.Context,
	zone Zone, zoneIdx int,
) {
	defer syncer.wg.Done()

	ipSet, err := zone.IPs.intoIpRanges()
	if err != nil {
		syncer.chErr <- mkIfMultiError(err, fmt.Sprintf("when conslidating IP Ranges for zone %d", zoneIdx))
		return
	}

	headers := &zone.AttachHeaders
	for _, zoneIps := range ipSet {
		tuple := Zipped2[ipRange, *headerSet]{zoneIps, headers}
		syncer.chRtv <- tuple
	}
}
