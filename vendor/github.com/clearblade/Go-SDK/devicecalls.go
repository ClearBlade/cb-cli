package GoSDK

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

const (
	_DEVICE_HEADER_KEY       = "ClearBlade-DeviceToken"
	_DEVICES_DEV_PREAMBLE    = "/admin/devices/"
	_DEVICES_USER_PREAMBLE   = "/api/v/2/devices/"
	_DEVICE_SESSION          = "/admin/v/4/session"
	_DEVICE_V3_USER_PREAMBLE = "/api/v/3/devices/"
	_DEVICE_V4_PREAMBLE      = "/api/v/4/devices/"
)

func (d *DevClient) GetDevices(systemKey string, query *Query) ([]interface{}, error) {
	return getDevices(d, systemKey, _DEVICES_DEV_PREAMBLE, query)
}

func (u *UserClient) GetDevices(systemKey string, query *Query) ([]interface{}, error) {
	return getDevices(u, systemKey, _DEVICES_USER_PREAMBLE, query)
}

func (u *DeviceClient) GetDevices(systemKey string, query *Query) ([]interface{}, error) {
	return getDevices(u, systemKey, _DEVICES_USER_PREAMBLE, query)
}

func (d *DevClient) GetDevicesCount(systemKey string, query *Query) (CountResp, error) {
	return getDevicesCount(d, systemKey, _DEVICE_V3_USER_PREAMBLE, query)
}

func (u *UserClient) GetDevicesCount(systemKey string, query *Query) (CountResp, error) {
	return getDevicesCount(u, systemKey, _DEVICE_V3_USER_PREAMBLE, query)
}

func (u *DeviceClient) GetDevicesCount(systemKey string, query *Query) (CountResp, error) {
	return getDevicesCount(u, systemKey, _DEVICE_V3_USER_PREAMBLE, query)
}

func getDevices(client cbClient, systemKey string, preamble string, query *Query) ([]interface{}, error) {
	creds, err := client.credentials()
	if err != nil {
		return nil, err
	}

	qry, err := createQueryMap(query)
	if err != nil {
		return nil, err
	}

	resp, err := get(client, preamble+systemKey, qry, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func getDevicesCount(client cbClient, systemKey string, preamble string, query *Query) (CountResp, error) {
	creds, err := client.credentials()
	if err != nil {
		return CountResp{Count: 0}, err
	}

	qry, err := createQueryMap(query)
	if err != nil {
		return CountResp{Count: 0}, err
	}

	resp, err := get(client, preamble+systemKey+"/count", qry, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return CountResp{Count: 0}, err
	}
	rval, ok := resp.Body.(map[string]interface{})
	if !ok {
		return CountResp{Count: 0}, fmt.Errorf("Bad type returned by getDevicesCount: %T, %s", resp.Body, resp.Body.(string))
	}

	return CountResp{
		Count: rval["count"].(float64),
	}, nil
}

func (d *DevClient) UpdateDevices(systemKey string, query *Query, changes map[string]interface{}) ([]interface{}, error) {
	return updateDevices(d, systemKey, _DEVICES_DEV_PREAMBLE, query, changes)
}

func (u *UserClient) UpdateDevices(systemKey string, query *Query, changes map[string]interface{}) ([]interface{}, error) {
	return updateDevices(u, systemKey, _DEVICES_USER_PREAMBLE, query, changes)
}

func (u *DeviceClient) UpdateDevices(systemKey string, query *Query, changes map[string]interface{}) ([]interface{}, error) {
	return updateDevices(u, systemKey, _DEVICES_USER_PREAMBLE, query, changes)
}

func updateDevices(client cbClient, systemKey string, preamble string, query *Query, changes map[string]interface{}) ([]interface{}, error) {
	creds, err := client.credentials()
	if err != nil {
		return nil, err
	}

	qry := query.serialize()
	body := map[string]interface{}{
		"query": qry,
		"$set":  changes,
	}

	resp, err := put(client, preamble+systemKey, body, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}

	return resp.Body.(map[string]interface{})["DATA"].([]interface{}), nil
}

func (d *DevClient) DeleteDevices(systemKey string, query *Query) error {
	return deleteDevices(d, systemKey, _DEVICES_DEV_PREAMBLE, query)
}

func (u *UserClient) DeleteDevices(systemKey string, query *Query) error {
	return deleteDevices(u, systemKey, _DEVICES_USER_PREAMBLE, query)
}

func (u *DeviceClient) DeleteDevices(systemKey string, query *Query) error {
	return deleteDevices(u, systemKey, _DEVICES_USER_PREAMBLE, query)
}

func deleteDevices(client cbClient, systemKey string, preamble string, query *Query) error {
	creds, err := client.credentials()
	if err != nil {
		return err
	}

	qry, err := createQueryMap(query)
	if err != nil {
		return err
	}

	resp, err := delete(client, preamble+systemKey, qry, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return err
	}

	return nil
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

func (d *DeviceClient) GetDevice(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _DEVICES_USER_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
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

func (u *UserClient) CreateDevice(systemKey, name string,
	data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(u, _DEVICES_USER_PREAMBLE+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DeviceClient) CreateDevice(systemKey, name string,
	data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(d, _DEVICES_USER_PREAMBLE+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DeviceClient) AuthenticateDeviceWithKey(systemKey, name, activeKey string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	postBody := map[string]interface{}{
		"deviceName": name,
		"activeKey":  activeKey,
	}
	resp, err := post(d, _DEVICES_USER_PREAMBLE+systemKey+"/auth", postBody, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	theJewels, ok := resp.Body.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Got unexpected return value from AuthenticateDeviceWithKey: %+v", theJewels)
	}
	d.DeviceToken = theJewels["deviceToken"].(string)
	return theJewels, nil
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

func (d *DeviceClient) DeleteDevice(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, _DEVICES_USER_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	_, err = mapResponse(resp, err)
	return err
}

func (u *UserClient) DeleteDevice(systemKey, name string) error {
	creds, err := u.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(u, _DEVICES_USER_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
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

func (u *DeviceClient) UpdateDevice(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
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

//  This stuff is developer only -- key sets for devices

func (d *DevClient) GetKeyset(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _DEVICES_DEV_PREAMBLE+"keys/"+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GenerateKeyset(systemKey, name string, count int) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{"count": count}
	resp, err := post(d, _DEVICES_DEV_PREAMBLE+"keys/"+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) RotateKeyset(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	resp, err := put(d, _DEVICES_DEV_PREAMBLE+"keys/"+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) DeleteKeyset(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, _DEVICES_DEV_PREAMBLE+"keys/"+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return err
	}
	return nil
}

func (d *DevClient) GetDeviceColumns(systemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := get(d, _DEVICES_DEV_PREAMBLE+systemKey+"/columns", nil, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting device columns: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting device columns: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) CreateDeviceColumn(systemKey, columnName, columnType string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"column_name": columnName,
		"type":        columnType,
	}
	resp, err := post(d, _DEVICES_DEV_PREAMBLE+systemKey+"/columns", data, creds, nil)
	if err != nil {
		return fmt.Errorf("Error creating device column: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error creating device column: %v", resp.Body)
	}

	return nil
}

func (d *DevClient) DeleteDeviceColumn(systemKey, columnName string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]string{"column_name": columnName}

	resp, err := delete(d, _DEVICES_DEV_PREAMBLE+systemKey+"/columns", data, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting device column: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting device column: %v", resp.Body)
	}

	return nil
}

func (d *DevClient) GetDeviceSession(systemKey string, query *Query) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	var qry map[string]string
	if query != nil {
		query_map := query.serialize()
		query_bytes, err := json.Marshal(query_map)
		if err != nil {
			return nil, err
		}
		qry = map[string]string{
			"query": url.QueryEscape(string(query_bytes)),
		}
	} else {
		qry = nil
	}
	resp, err := get(d, _DEVICE_SESSION+"/"+systemKey+"/device", qry, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting device session data: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting device session data: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) DeleteDeviceSession(systemKey string, query *Query) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	var qry map[string]string
	if query != nil {
		query_map := query.serialize()
		query_bytes, err := json.Marshal(query_map)
		if err != nil {
			return err
		}
		qry = map[string]string{
			"query": url.QueryEscape(string(query_bytes)),
		}
	} else {
		qry = nil
	}
	resp, err := delete(d, _DEVICE_SESSION+"/"+systemKey+"/device", qry, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting device session data: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting device session data: %v", resp.Body)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (dvc *DeviceClient) credentials() ([][]string, error) {
	ret := make([][]string, 0)
	if dvc.DeviceToken != "" {
		ret = append(ret, []string{
			_DEVICE_HEADER_KEY,
			dvc.DeviceToken,
		})
	}
	if dvc.SystemKey != "" && dvc.SystemSecret != "" {
		ret = append(ret, []string{
			_HEADER_SECRET_KEY,
			dvc.SystemSecret,
		})
		ret = append(ret, []string{
			_HEADER_KEY_KEY,
			dvc.SystemKey,
		})

	}

	if len(ret) == 0 {
		return [][]string{}, errors.New("No SystemSecret/SystemKey combo, or DeviceToken found")
	} else {
		return ret, nil
	}
}

//  User (user/device) calls (type Client)
//func (d *DeviceClient) AuthenticateDeviceWithKey(systemKey, name, activeKey string) (map[string]interface{}, error) {

// "Login and logout"
func (dvc *DeviceClient) Authenticate() (*AuthResponse, error) {
	_, err := dvc.AuthenticateDeviceWithKey(dvc.SystemKey, dvc.DeviceName, dvc.ActiveKey)
	return nil, err
}

func (dvc *DeviceClient) Logout() error {
	return nil
}

// Device MQTT calls are mqtt.go

func (dvc *DeviceClient) preamble() string {
	return _DEVICES_USER_PREAMBLE
}

func (dvc *DeviceClient) setToken(tok string) {
	dvc.DeviceToken = tok
}

func (dvc *DeviceClient) getToken() string {
	return dvc.DeviceToken
}

func (dvc *DeviceClient) getRefreshToken() string {
	return dvc.RefreshToken
}
func (dvc *DeviceClient) setRefreshToken(t string) {
	dvc.RefreshToken = t
}
func (dvc *DeviceClient) setExpiresAt(t float64) {
	dvc.ExpiresAt = t
}
func (dvc *DeviceClient) getExpiresAt() float64 {
	return dvc.ExpiresAt
}

func (dvc *DeviceClient) getSystemInfo() (string, string) {
	return dvc.SystemKey, dvc.SystemSecret
}

func (dvc *DeviceClient) getMessageId() uint16 {
	return uint16(dvc.mrand.Int())
}

func (dvc *DeviceClient) getHttpAddr() string {
	return dvc.HttpAddr
}

func (dvc *DeviceClient) getMqttAddr() string {
	return dvc.MqttAddr
}

func ConnectedDevices(client cbClient, systemKey string) (map[string]interface{}, error) {
	creds, err := client.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(client, _DEVICE_V4_PREAMBLE+systemKey+"/connections", nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func DeviceConnections(client cbClient, systemKey, deviceName string) (map[string]interface{}, error) {
	creds, err := client.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(client, _DEVICE_V4_PREAMBLE+systemKey+"/connections/"+deviceName, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func ConnectedDeviceCount(client cbClient, systemKey string) (map[string]interface{}, error) {
	creds, err := client.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(client, _DEVICE_V4_PREAMBLE+systemKey+"/connectioncount", nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}
