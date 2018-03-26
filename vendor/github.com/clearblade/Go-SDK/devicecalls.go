package GoSDK

import ()

const (
	_DEVICES_DEV_PREAMBLE  = "/admin/devices/"
	_DEVICES_USER_PREAMBLE = "/api/v/2/devices"
)

func (d *DevClient) GetDevices(systemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _DEVICES_DEV_PREAMBLE+systemKey, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (u *UserClient) GetDevices(systemKey string) ([]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(u, _DEVICES_USER_PREAMBLE+systemKey, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) GetDevice(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _DEVICES_DEV_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) GetDevice(systemKey, name string) (map[string]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(u, _DEVICES_USER_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) CreateDevice(systemKey, name string,
	data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(d, _DEVICES_DEV_PREAMBLE+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) DeleteDevice(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, _DEVICES_DEV_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	_, err = mapResponse(resp, err)
	return err
}

func (d *DevClient) UpdateDevice(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(d, _DEVICES_DEV_PREAMBLE+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) UpdateDevice(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(u, _DEVICES_USER_PREAMBLE+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}
