package GoSDK

import (
	"fmt"
)

const (
	_CODE_PREAMBLE = "/api/v/1/code"
)

func callService(c cbClient, systemKey, name string, params map[string]interface{}, log bool) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	var resp *CbResp
	if log {

		resp, err = post(c, _CODE_PREAMBLE+"/"+systemKey+"/"+name, params, creds, map[string][]string{"Logging-enabled": []string{"true"}})
	} else {
		resp, err = post(c, _CODE_PREAMBLE+"/"+systemKey+"/"+name, params, creds, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("Error calling %s service: %v", name, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error calling %s service: %v", name, resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

//CallService performs a call against the specific service with the specified parameters. The logging argument will allow the developer to call the service with logging enabled for just that run.
//The return value is a map[string]interface{} where the results will be stored in the key "results". If logs were enabled, they'll be in "log".
func (d *DevClient) CallService(systemKey, name string, params map[string]interface{}, log bool) (map[string]interface{}, error) {
	return callService(d, systemKey, name, params, log)
}

//CallService performs a call against the specific service with the specified parameters.
//The return value is a map[string]interface{} where the results will be stored in the key "results". If logs were enabled, they'll be in "log".
func (u *UserClient) CallService(systemKey, name string, params map[string]interface{}) (map[string]interface{}, error) {
	return callService(u, systemKey, name, params, false)
}
