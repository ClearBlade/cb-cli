package GoSDK

func (d *DevClient) EnterProvisioningMode() (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := post(d, "/admin/edgemode/provisioning", map[string]interface{}{}, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}

	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) EnterRuntimeMode(info map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := post(d, "/admin/edgemode/runtime", info, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}
