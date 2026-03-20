// Wrapper for Unmarshalling URLs as stdlib url.URL doesn't do UnmarshalJSON

package realip_zoning

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type urlMarshal struct {
	*url.URL
}

func (u *urlMarshal) UnmarshalJSON(data []byte) error {
	var rawUrl string
	err1 := json.Unmarshal(data, &rawUrl)
	if err1 != nil {
		return fmt.Errorf("Unable to read URL string: %+w", err1)
	}

	var inner, err = url.Parse(rawUrl)
	if err != nil {
		return fmt.Errorf("Unable to parse URL: %+w", err)
	}
	u.URL = inner

	return nil
}

func (u urlMarshal) MarshalJSON() ([]byte, error) {
	if u.URL == nil {
		return []byte(`""`), nil
	}
	return json.Marshal(u.String())
}
