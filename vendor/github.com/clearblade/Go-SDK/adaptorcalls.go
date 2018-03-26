package GoSDK

import (
	"encoding/base64"
	"fmt"
)

const (
	_ADAPTOR_HEADER_KEY    = "ClearBlade-DeviceToken"
	_ADAPTORS_DEV_PREAMBLE = "/admin/"
)

func (d *DevClient) GetAdaptors(systemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors", nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (u *UserClient) GetAdaptors(systemKey string) ([]interface{}, error) {
	return nil, fmt.Errorf("User client cannot get adaptors")
}

func (d *DeviceClient) GetAdaptors(systemKey string) ([]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot get adaptors")
}

func (d *DevClient) GetAdaptor(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) GetAdaptor(systemKey, name string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot get adaptor")
}

func (d *DeviceClient) GetAdaptor(systemKey, name string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot get adaptor")
}

func (d *DevClient) CreateAdaptor(systemKey, name string,
	data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	data["name"] = name
	resp, err := post(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors", data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) CreateAdaptor(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot create an adaptor")
}

func (d *DeviceClient) CreateAdaptor(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot create an adaptor")
}

func (d *DevClient) DeleteAdaptor(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+name, nil, creds, nil)
	_, err = mapResponse(resp, err)
	return err
}

func (u *UserClient) DeleteAdaptor(systemKey, name string) error {
	return fmt.Errorf("User client cannot delete an adaptor")
}

func (d *DeviceClient) DeleteAdaptor(systemKey, name string) error {
	return fmt.Errorf("Device client cannot delete an adaptor")
}

func (d *DevClient) UpdateAdaptor(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) UpdateAdaptor(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot update an adaptor")
}

func (d *DeviceClient) UpdateAdaptor(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot update an adaptor")
}

func (d *DevClient) GetAdaptorFiles(systemKey, adaptorName string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+adaptorName+"/files", nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (u *UserClient) GetAdaptorFiles(systemKey, adaptorName string) ([]interface{}, error) {
	return nil, fmt.Errorf("User client cannot get adaptor files")
}

func (d *DeviceClient) GetAdaptorFiles(systemKey, adaptorName string) ([]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot get adaptor files")
}

func (d *DevClient) GetAdaptorFile(systemKey, adaptorName, fileName string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+adaptorName+"/files/"+fileName, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) GetAdaptorFile(systemKey, adaptorName, fileName string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot get adaptor file")
}

func (d *DeviceClient) GetAdaptorFile(systemKey, adaptorName, fileName string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot get adaptor file")
}

func getContentsForFile(fileValue interface{}) (string, error) {
	// if a user sends us a byte array (probably an executable file) we need to base64 encode it
	// if a user sends us a string we assume it's base64 encoded
	var fileContents string
	switch fileValue.(type) {
	case []byte:
		fileContents = base64.StdEncoding.EncodeToString(fileValue.([]byte))
	case string:
		fileContents = fileValue.(string)
	default:
		return "", fmt.Errorf("Bad type for 'file' k/v pair in CreateAdaptorFile: %T: (%+v)", fileValue, fileValue)
	}
	return fileContents, nil
}

func (d *DevClient) CreateAdaptorFile(systemKey, adaptorName, fileName string,
	data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	contentsIF, exists := data["file"]
	if !exists {
		return nil, fmt.Errorf("'file' key/value pair missing in CreateAdaptorFile")
	}

	if fileContents, err := getContentsForFile(contentsIF); err != nil {
		return nil, err
	} else {
		data["file"] = fileContents
	}

	data["name"] = fileName
	data["adaptor_name"] = adaptorName

	resp, err := post(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+adaptorName+"/files", data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) CreateAdaptorFile(systemKey, adaptorName, fileName string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot create an adaptor file")
}

func (d *DeviceClient) CreateAdaptorFile(systemKey, adaptorName, fileName string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot create an adaptor file")
}

func (d *DevClient) DeleteAdaptorFile(systemKey, adaptorName, fileName string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+adaptorName+"/files/"+fileName, nil, creds, nil)
	_, err = mapResponse(resp, err)
	return err
}

func (u *UserClient) DeleteAdaptorFile(systemKey, adaptorName, fileName string) error {
	return fmt.Errorf("User client cannot delete an adaptor file")
}

func (d *DeviceClient) DeleteAdaptorFile(systemKey, adaptorName, fileName string) error {
	return fmt.Errorf("Device client cannot delete an adaptor file")
}

func (d *DevClient) UpdateAdaptorFile(systemKey, adaptorName, fileName string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	contentsIF, exists := data["file"]
	if !exists {
		return nil, fmt.Errorf("'file' key/value pair missing in UpdateAdaptorFile")
	}

	if fileContents, err := getContentsForFile(contentsIF); err != nil {
		return nil, err
	} else {
		data["file"] = fileContents
	}
	data["name"] = fileName
	resp, err := put(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+adaptorName+"/files/"+fileName, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) UpdateAdaptorFile(systemKey, adaptorName, fileName string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot update an adaptor file")
}

func (d *DeviceClient) UpdateAdaptorFile(systemKey, adaptorName, fileName string, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot update an adaptor file")
}

func (d *DevClient) DeployAdaptor(systemKey, adaptorName string, deploySpec map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+adaptorName+"/deploy", deploySpec, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) DeployAdaptor(systemKey, adaptorName string, deploySpec map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot deploy an adaptor")
}

func (d *DeviceClient) DeployAdaptor(systemKey, adaptorName string, deploySpec map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot deploy an adaptor")
}

func (d *DevClient) ControlAdaptor(systemKey, adaptorName string, controlSpec map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(d, _ADAPTORS_DEV_PREAMBLE+systemKey+"/adaptors/"+adaptorName+"/control", controlSpec, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) ControlAdaptor(systemKey, adaptorName string, controlSpec map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("User client cannot control an adaptor")
}

func (d *DeviceClient) ControlAdaptor(systemKey, adaptorName string, controlSpec map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Device client cannot control an adaptor")
}
