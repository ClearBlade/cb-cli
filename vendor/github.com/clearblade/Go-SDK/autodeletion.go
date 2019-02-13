package GoSDK

import (
	"fmt"
	"net/http"
	"path"

	"github.com/pkg/errors"
)

const (
	autodeletion_preamble = "/api/v/4/message"
)

type AutodeletionSettings struct {
	Enabled       bool  `json:"enabled"`
	MaxSizeKb     int   `json:"expiration_age_seconds"`
	MaxRows       int   `json:"max_size_kb"`
	MaxAgeSeconds int64 `json:"max_rows"`
}

func unpackMapToAutodeletionSettings(m interface{}) (AutodeletionSettings, error) {
	switch m := m.(type) {
	case map[string]interface{}:
		return AutodeletionSettings{
			Enabled:       m["enabled"].(bool),
			MaxSizeKb:     int(m["max_size_kb"].(float64)),
			MaxRows:       int(m["max_rows"].(float64)),
			MaxAgeSeconds: int64(m["expiration_age_seconds"].(float64)),
		}, nil
	default:
		return AutodeletionSettings{}, fmt.Errorf("Unexpected response from Autodeletion Settings Endpoint: %+v", m)
	}

}

func GetAutodeletionSettings(c cbClient, systemkey string) (AutodeletionSettings, error) {
	creds, err := c.credentials()
	if err != nil {
		return AutodeletionSettings{}, errors.Wrap(err, "Error with credentials")
	}

	resp, err := get(c, path.Join(autodeletion_preamble, systemkey, "autodelete"), nil, creds, nil)
	if err != nil {
		return AutodeletionSettings{}, fmt.Errorf("Error getting autodeletion settings: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return AutodeletionSettings{}, fmt.Errorf("Error getting autodeletion settings: %s", resp.Body)
	}

	fmt.Println(resp.Body)

	return unpackMapToAutodeletionSettings(resp.Body)
}

func SetAutodeletionSettings(c cbClient, systemkey string, newSettings AutodeletionSettings) (AutodeletionSettings, error) {
	creds, err := c.credentials()
	if err != nil {
		return AutodeletionSettings{}, errors.Wrap(err, "Error with credentials")
	}

	p := path.Join(autodeletion_preamble, systemkey, "autodelete")
	resp, err := post(c, p, newSettings, creds, nil)
	if err != nil {
		return AutodeletionSettings{}, errors.Wrap(err, "Error setting autodeletion settings ("+p+")")
	}
	if resp.StatusCode != http.StatusOK {
		return AutodeletionSettings{}, fmt.Errorf("Error setting autodeletion settings: %s: %d: %s", p, resp.StatusCode, resp.Body)
	}
	return unpackMapToAutodeletionSettings(resp.Body)
}

func SetAllAutodeletionSettings(c cbClient, newSettings AutodeletionSettings) (systems []string, err error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, errors.Wrap(err, "Error with credentials")
	}

	resp, err := post(c, path.Join(autodeletion_preamble, "bulksetautodelete"), newSettings, creds, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error Setting All Autodeletion Settings")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error Setting All Autodeletion Settings: %d: %s", resp.StatusCode, resp.Body)
	}
	fmt.Println(resp.Body)
	switch body := resp.Body.(type) {
	case map[string]interface{}:
		for _, sys := range body["systems"].([]interface{}) {
			systems = append(systems, sys.(string))
		}
	default:
		return nil, fmt.Errorf("Unexpected response from Autodeletion Settings Endpoint: %+v", body)
	}

	return systems, nil
}
