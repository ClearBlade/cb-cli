package GoSDK

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

const (
	_USER_HEADER_KEY = "ClearBlade-UserToken"
	_USER_PREAMBLE   = "/api/v/1/user"
	_USER_V2         = "/api/v/2/user"
	_USER_ADMIN      = "/admin/user"
	_USER_SESSION    = "/admin/v/4/session"
)

func (u *UserClient) credentials() ([][]string, error) {
	ret := make([][]string, 0)
	if u.UserToken != "" {
		ret = append(ret, []string{
			_USER_HEADER_KEY,
			u.UserToken,
		})
	}
	if u.SystemSecret != "" && u.SystemKey != "" {
		ret = append(ret, []string{
			_HEADER_SECRET_KEY,
			u.SystemSecret,
		})
		ret = append(ret, []string{
			_HEADER_KEY_KEY,
			u.SystemKey,
		})

	}

	if len(ret) == 0 {
		return [][]string{}, errors.New("No SystemSecret/SystemKey combo, or UserToken found")
	} else {
		return ret, nil
	}
}

func (u *UserClient) preamble() string {
	return _USER_PREAMBLE
}

func (u *UserClient) getSystemInfo() (string, string) {
	return u.SystemKey, u.SystemSecret
}

func (u *UserClient) setToken(t string) {
	u.UserToken = t
}
func (u *UserClient) getToken() string {
	return u.UserToken
}

func (u *UserClient) getMessageId() uint16 {
	return uint16(u.mrand.Int())
}

func (u *UserClient) GetUserCount(systemKey string) (int, error) {
	creds, err := u.credentials()
	if err != nil {
		return -1, err
	}
	resp, err := get(u, u.preamble()+"/count", nil, creds, nil)
	if err != nil {
		return -1, fmt.Errorf("Error getting count: %v", err)
	}
	if resp.StatusCode != 200 {
		return -1, fmt.Errorf("Error getting count: %v", resp.Body)
	}
	bod := resp.Body.(map[string]interface{})
	theCount := int(bod["count"].(float64))
	return theCount, nil
}

//GetUserColumns returns the description of the columns in the user table
//Returns a structure shaped []map[string]interface{}{map[string]interface{}{"ColumnName":"blah","ColumnType":"int"}}
func (d *DevClient) GetUserColumns(systemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := get(d, _USER_ADMIN+"/"+systemKey+"/columns", nil, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting user columns: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting user columns: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

//CreateUserColumn creates a new column in the user table
func (d *DevClient) CreateUserColumn(systemKey, columnName, columnType string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"column_name": columnName,
		"type":        columnType,
	}

	resp, err := post(d, _USER_ADMIN+"/"+systemKey+"/columns", data, creds, nil)
	if err != nil {
		return fmt.Errorf("Error creating user column: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error creating user column: %v", resp.Body)
	}

	return nil
}

func (d *DevClient) DeleteUserColumn(systemKey, columnName string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]string{"column": columnName}

	resp, err := delete(d, _USER_ADMIN+"/"+systemKey+"/columns", data, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting user column: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting user column: %v", resp.Body)
	}

	return nil
}

func (u *UserClient) UpdateUser(userQuery *Query, changes map[string]interface{}) error {
	return updateUser(u, userQuery, changes)
}

func updateUser(c cbClient, userQuery *Query, changes map[string]interface{}) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	query := userQuery.serialize()
	body := map[string]interface{}{
		"query":   query,
		"changes": changes,
	}

	resp, err := put(c, _USER_V2+"/info", body, creds, nil)
	if err != nil {
		return fmt.Errorf("Error updating data: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating data: %v", resp.Body)
	}

	return nil
}

func (d *DevClient) GetUserSession(systemKey string, query *Query) ([]interface{}, error) {
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
	resp, err := get(d, _USER_SESSION+"/"+systemKey+"/user", qry, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting user session data: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting user session data: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) DeleteUserSession(systemKey string, query *Query) error {
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
	resp, err := delete(d, _USER_SESSION+"/"+systemKey+"/user", qry, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting user session data: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting user session data: %v", resp.Body)
	}
	return nil
}
