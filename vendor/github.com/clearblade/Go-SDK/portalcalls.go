package GoSDK

const (
	_PORTALS_DEV_PREAMBLE  = "/admin/portals/"
	_PORTALS_USER_PREAMBLE = "/api/v/2/portals/"
)

func (d *DevClient) GetPortals(systemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _PORTALS_DEV_PREAMBLE+systemKey, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) GetPortal(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _PORTALS_USER_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) CreatePortal(systemKey, name string, dash map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := post(d, _PORTALS_USER_PREAMBLE+systemKey+"/"+name, dash, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) UpdatePortal(systemKey, name string, dash map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(d, _PORTALS_USER_PREAMBLE+systemKey+"/"+name, dash, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) DeletePortal(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, _PORTALS_USER_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return err
	}
	return nil
}
