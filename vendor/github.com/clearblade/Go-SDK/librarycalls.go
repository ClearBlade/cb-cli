package GoSDK

import (
	//"fmt"
	"strconv"
)

const (
	_LIB_PREAMBLE  = "/codeadmin/v/2/library"
	_HIST_PREAMBLE = "/codeadmin/v/2/history/library"
)

//GetLibrary returns a list of libraries for a system
//Returns an object of the following []map[string]interface{}{map[string]interface{}{"system_key":"associated system key","name":"the name of the library","description":"library description","version":1,"code":"function blabla(){return "blahbla"}","dependencies":"clearblade"}, ...}
func (d *DevClient) GetLibraries(systemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _LIB_PREAMBLE+"/"+systemKey, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

//GetLibrary returns information about a specific library
//Returns a single object following the pattern specified in GetLibraries
func (d *DevClient) GetLibrary(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _LIB_PREAMBLE+"/"+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

//CreateLibrary allows the developer to create a library to be called by other service functions
//returns a single object following the pattern specified in GetLibraries for the newly-created library
func (d *DevClient) CreateLibrary(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(d, _LIB_PREAMBLE+"/"+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

//UpdateLibrary allows the developer to change the content of the library
//returns a single object following the pattern specified in GetLibraries with the updated details
func (d *DevClient) UpdateLibrary(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(d, _LIB_PREAMBLE+"/"+systemKey+"/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

//DeleteLibrary allows the developer to remove library content
func (d *DevClient) DeleteLibrary(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, _LIB_PREAMBLE+"/"+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return err
	}
	return nil
}

//GetVersionHistory returns a slice of library descriptions corresponding to each library
//Returns an object with the same shape as that of GetLibrariesForSystem, but with each version of the specified library
func (d *DevClient) GetVersionHistory(systemKey, name string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _HIST_PREAMBLE+"/"+systemKey+"/"+name, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

//GetVersion gets the current version of a library
//Returns an object with the same shape as that of GetLibrariesForSystem, but a specific version thereof
func (d *DevClient) GetVersion(systemKey, name string, version int) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _HIST_PREAMBLE+"/"+systemKey+"/"+name+"/"+strconv.Itoa(version), nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}
