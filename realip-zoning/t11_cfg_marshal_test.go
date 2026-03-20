package realip_zoning

import (
	"encoding/json"
	"testing"
)

const sampleIPConf = `{
	"fromList": [ "127.0.0.0/8" ],
	"fromURLs": [ "http://example.com/ips.txt" ],
	"fromFiles": [ "/etc/example/ips.txt" ],
	"fromDir": "/etc/example/ips.d/"
}`

const sampleConfig = `{
	"trustedProxies": {
		"useHeader": "X-Use-Header-Test",
		"fromURLs": [
			"http://example.com/ips"
		],
		"fromDir": "/etc/traefik/plugin.d/realip-zoning/ips.d/",
		"fromList": [
			"192.168.0.0/16"
		]
	},
	"nullZoneHeaders": {
		"X-Null-Zone-Header": "test-value"
	},
	"zones": [
		{
			"ips": {
				"fromList": [
					"10.0.0.0/8"
				],
				"fromURLs": [
					"http://example.com/zone1-ips"
				],
				"fromDir": "/etc/traefik/plugin.d/realip-zoning/zone1-ips.d/"
			},
			"attachHeaders": {
				"X-Zone-Header": "zone1"
			}
		}
	]
}`

func TestUnmarshalIPSources(harness *testing.T) {
	var ips IPSources
	err := json.Unmarshal([]byte(sampleIPConf), &ips)
	if err != nil {
		harness.Fatalf("Failed to unmarshal IPSources: %v", err)
	}

	harness.Logf("Parsed IPSources: %+v", ips)
}

func TestUnmarshalWholeConf(harness *testing.T) {
	var cfg Config
	err := json.Unmarshal([]byte(sampleConfig), &cfg)
	if err != nil {
		harness.Fatalf("Failed to unmarshal Config: %v", err)
	}

	harness.Logf("Parsed Config: %+v", cfg)
}
