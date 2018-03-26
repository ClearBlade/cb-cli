package GoSDK

import (
	"fmt"
)

const (
	_CODE_PREAMBLE      = "/api/v/1/code"
	_CODE_USER_PREAMBLE = "/api/v/3/code"
)

func GetServiceNames(c cbClient, systemKey string) ([]string, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, _CODE_USER_PREAMBLE+"/"+systemKey, nil, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting services: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting services: %v", resp.Body)
	}
	code := resp.Body.(map[string]interface{})["code"]
	sliceBody, isSlice := code.([]interface{})
	if !isSlice && code != nil {
		return nil, fmt.Errorf("Error getting services: server returned unexpected response")
	}
	services := make([]string, len(sliceBody))
	for i, service := range sliceBody {
		services[i] = service.(string)
	}
	return services, nil
}

func getService(c cbClient, systemKey, name string) (*Service, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, _CODE_USER_PREAMBLE+"/"+systemKey+"/service/"+name, nil, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting service: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting service: %v", resp.Body)
	}
	mapBody := resp.Body.(map[string]interface{})
	paramsSlice := mapBody["params"].([]interface{})
	params := make([]string, len(paramsSlice))
	for i, param := range paramsSlice {
		params[i] = param.(string)
	}
	svc := &Service{
		Name:    name,
		System:  systemKey,
		Code:    mapBody["code"].(string),
		Version: int(mapBody["current_version"].(float64)),
		Params:  params,
	}
	return svc, nil
}

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

func createService(c cbClient, systemKey, name, code string, extra map[string]interface{}) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	extra["code"] = code
	resp, err := post(c, _CODE_USER_PREAMBLE+"/"+systemKey+"/service/"+name, extra, creds, nil)
	if err != nil {
		return fmt.Errorf("Error creating new service: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error creating new service: %v", resp.Body)
	}
	return nil
}

func deleteService(c cbClient, systemKey, name string) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(c, _CODE_USER_PREAMBLE+"/"+systemKey+"/service/"+name, nil, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting service: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting service: %v", resp.Body)
	}
	return nil
}

func updateService(c cbClient, sysKey, name, code string, extra map[string]interface{}) (error, map[string]interface{}) {
	creds, err := c.credentials()
	if err != nil {
		return err, nil
	}
	resp, err := put(c, _CODE_USER_PREAMBLE+"/"+sysKey+"/service/"+name, extra, creds, nil)
	if err != nil {
		return fmt.Errorf("Error updating service: %v\n", err), nil
	}
	body, ok := resp.Body.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Service not created. First create service..."), nil
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating service: %v\n", resp.Body), nil
	}
	return nil, body
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

//GetServiceNames retrieves the service names for a particular system
func (u *UserClient) GetServiceNames(systemKey string) ([]string, error) {
	return GetServiceNames(u, systemKey)
}

//GetService returns information about a specified service
func (u *UserClient) GetService(systemKey, name string) (*Service, error) {
	return getService(u, systemKey, name)
}

func (u *DevClient) CreateService(systemKey, name, code string, params []string) error {
	extra := map[string]interface{}{"parameters": params}
	return createService(u, systemKey, name, code, extra)
}

func (u *UserClient) CreateService(systemKey, name, code string, params []string) error {
	extra := map[string]interface{}{"parameters": params}
	return createService(u, systemKey, name, code, extra)
}

func (u *UserClient) DeleteService(systemKey, name string) error {
	return deleteService(u, systemKey, name)
}

func (u *UserClient) UpdateService(systemKey, name, code string, params []string) (error, map[string]interface{}) {
	extra := map[string]interface{}{"code": code, "name": name, "parameters": params}
	return updateService(u, systemKey, name, code, extra)
}

func (u *UserClient) CreateTrigger(systemKey, name string,
	data map[string]interface{}) (map[string]interface{}, error) {
	return u.CreateEventHandler(systemKey, name, data)
}

func (u *UserClient) DeleteTrigger(systemKey, name string) error {
	return u.DeleteEventHandler(systemKey, name)
}

func (u *UserClient) UpdateTrigger(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return u.UpdateEventHandler(systemKey, name, data)
}

func (u *UserClient) GetTrigger(systemKey, name string) (map[string]interface{}, error) {
	return u.GetEventHandler(systemKey, name)
}
