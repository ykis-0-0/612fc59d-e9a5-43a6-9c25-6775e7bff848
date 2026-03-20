package realip_zoning

import (
	"context"
	"net/http"
	"net/netip"
)

type RealIPZoningPlugin struct {
	next http.Handler

}

// #region Traefik Plugin Interface

type Config struct {
}

func CreateConfig() *Config {
	return &Config{}
}

func New(ctx context.Context, next http.Handler, config *Config) (http.Handler, error) {

	return &RealIPZoningPlugin{
		next: next,
	}, nil
}

// #endregion

// To comply with interface http.Handler
func (this *RealIPZoningPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// TODO

	this.next.ServeHTTP(rw, req)
}
