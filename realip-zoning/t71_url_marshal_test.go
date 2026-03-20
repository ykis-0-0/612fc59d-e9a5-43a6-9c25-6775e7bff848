package realip_zoning

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestUnmarshalURL(harness *testing.T) {

	var output urlMarshal
	const input = `"https://who_still_use:this_to_login@subdomain.example.com:8901/some/where/there.jsp?key=value#anchor"`

	err := json.Unmarshal([]byte(input), &output)
	if err != nil {
		harness.Fatalf("Failed to unmarshal URL: %v", err)
	}

	harness.Logf("Parsed URL: %+v", output)
}

func TestMarshalURL(harness *testing.T) {
	var input = urlMarshal{
		URL: &url.URL{
			Scheme:   "scheme",
			User:     url.UserPassword("user", "password"),
			Host:     "host.tld:1234",
			Path:     "path/file.ext",
			RawQuery: "query=answer",
			Fragment: "fragment",
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		harness.Fatalf("Failed to marshal URL: %v", err)
	}

	harness.Logf("%s", data)
}
