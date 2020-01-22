package GoSDK

func (d *DevClient) PerformMonitoring() (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := get(d, "/admin/metrics", nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}

	return resp.Body.(map[string]interface{}), nil
}
