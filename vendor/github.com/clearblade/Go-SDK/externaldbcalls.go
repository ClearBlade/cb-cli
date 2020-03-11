package GoSDK

import (
	"fmt"
)

const (
	_EXTERNAL_DB_PREAMBLE = "/api/v/4/external-db/"
)

func (d *DevClient) AddExternalDBConnection(systemKey, name string, data map[string]interface{}) error {
	return addExternalDBConnection(d, systemKey, name, data)
}

func (u *UserClient) AddExternalDBConnection(systemKey, name string, data map[string]interface{}) error {
	return addExternalDBConnection(u, systemKey, name, data)
}

func (d *DevClient) GetAllExternalDBConnections(systemKey string) ([]interface{}, error) {
	return getAllExternalDBConnections(d, systemKey)
}

func (d *DevClient) GetExternalDBConnection(systemKey, name string) (map[string]interface{}, error) {
	return getExternalDBConnection(d, systemKey, name)
}

func (u *UserClient) GetExternalDBConnection(systemKey, name string) (map[string]interface{}, error) {
	return getExternalDBConnection(u, systemKey, name)
}

func (d *DevClient) UpdateExternalDBConnection(systemKey, name string, changes map[string]interface{}) error {
	return updateExternalDBConnection(d, systemKey, name, changes)
}

func (u *UserClient) UpdateExternalDBConnection(systemKey, name string, changes map[string]interface{}) error {
	return updateExternalDBConnection(u, systemKey, name, changes)
}

func (d *DevClient) DeleteExternalDBConnection(systemKey, name string) error {
	return deleteExternalDBConnection(d, systemKey, name)
}

func (u *UserClient) DeleteExternalDBConnection(systemKey, name string) error {
	return deleteExternalDBConnection(u, systemKey, name)
}

func (d *DevClient) PerformExternalDBOperation(systemKey, name string, operation map[string]interface{}) (map[string]interface{}, error) {
	return performExternalDBOperation(d, systemKey, name, operation)
}

func (u *UserClient) PerformExternalDBOperation(systemKey, name string, operation map[string]interface{}) (map[string]interface{}, error) {
	return performExternalDBOperation(u, systemKey, name, operation)
}

func addExternalDBConnection(c cbClient, systemKey, name string, data map[string]interface{}) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := post(c, _EXTERNAL_DB_PREAMBLE+systemKey+"/"+name, data, creds, nil)
	if err != nil {
		return fmt.Errorf("Error adding external db connection: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error adding external db connection: %v", resp.Body)
	}
	return nil
}

func getExternalDBConnection(c cbClient, systemKey, name string) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, _EXTERNAL_DB_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting external db connection: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting external db connection: %v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func getAllExternalDBConnections(c cbClient, systemKey string) ([]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, _EXTERNAL_DB_PREAMBLE+systemKey, nil, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting all external db connections: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting all external db connections: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

func updateExternalDBConnection(c cbClient, systemKey, name string, changes map[string]interface{}) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := put(c, _EXTERNAL_DB_PREAMBLE+systemKey+"/"+name, changes, creds, nil)
	if err != nil {
		return fmt.Errorf("Error updating external db connection: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating external db connection: %v", resp.Body)
	}
	return nil
}

func deleteExternalDBConnection(c cbClient, systemKey, name string) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(c, _EXTERNAL_DB_PREAMBLE+systemKey+"/"+name, nil, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting external db connection: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting external db connection: %v", resp.Body)
	}
	return nil
}

func performExternalDBOperation(c cbClient, systemKey, name string, operation map[string]interface{}) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(c, _EXTERNAL_DB_PREAMBLE+systemKey+"/"+name+"/data", operation, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error performing external db operation: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error performing external db operation: %v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}
