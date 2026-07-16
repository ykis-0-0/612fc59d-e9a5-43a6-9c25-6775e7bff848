package realip_zoning

import (
	"context"
	"net/http"
	"net/netip"
)

type RealIPZoningPlugin struct {
	next http.Handler

	// It may be nil, so the pointer
	proxyConf *proxyConf_t

	nullZoneHeaders headerSet
	invertedZones   map[netip.Prefix]*headerSet
}

// #region Traefik Plugin Interface

type Config struct {
	TrustedProxies *struct {
		UseHeader string `json:"useHeader"`

		IPSources
	} `json:"trustedProxies,omitempty"`
	NullZoneHeaders headerSet `json:"nullZoneHeaders,omitempty"`
	Zones           []Zone    `json:"zones,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

func New(ctx context.Context, next http.Handler, config *Config) (http.Handler, error) {
	var proxyConf *proxyConf_t
	proxyConf, err := mkProxyConf(ctx, config)
	if err != nil {
		return nil, err
	}

	var invertedZones map[netip.Prefix]*headerSet
	invertedZones, err = toInvertedZones(ctx, config.Zones)
	if err != nil {
		return nil, err
	}

	return &RealIPZoningPlugin{
		next: next,

		proxyConf: proxyConf,

		nullZoneHeaders: config.NullZoneHeaders,
		invertedZones:   invertedZones,
	}, nil
}

// #endregion

// To comply with interface http.Handler
func (this *RealIPZoningPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// TODO

	this.next.ServeHTTP(rw, req)
}
