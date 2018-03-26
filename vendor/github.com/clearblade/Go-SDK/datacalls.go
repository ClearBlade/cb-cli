package GoSDK

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const (
	_DATA_PREAMBLE      = "/api/v/1/data/"
	_DATA_NAME_PREAMBLE = "/api/v/1/collection/"
)

//Inserts data into the platform. The interface is either a map[string]interface{} representing a row, or a []map[string]interface{} representing many rows.
func (u *UserClient) InsertData(collection_id string, data interface{}) error {
	_, err := insertdata(u, collection_id, data)
	return err
}

//Inserts data into the platform. The interface is either a map[string]interface{} representing a row, or a []map[string]interface{} representing many rows.
func (d *DevClient) InsertData(collection_id string, data interface{}) error {
	_, err := insertdata(d, collection_id, data)
	return err
}

//CreateData is an alias for InsertData, but returns a response value, it should be a slice of strings representing the item ids (if not using an external datastore)
func (d *DevClient) CreateData(collection_id string, data interface{}) ([]interface{}, error) {
	resp, err := insertdata(d, collection_id, data)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//CreateData is an alias for InsertData, but returns a response value, it should be a slice of strings representing the item ids (if not using an external datastore)
func (u *UserClient) CreateData(collection_id string, data interface{}) ([]interface{}, error) {
	resp, err := insertdata(u, collection_id, data)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func insertdata(c cbClient, collection_id string, data interface{}) ([]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(c, _DATA_PREAMBLE+collection_id, data, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error inserting: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error inserting: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

//GetData performs a query against a collection. The query object is discussed elsewhere. If the query object is nil, then it will return all of the data.
//The return value is a key-value of the types. Note that due to the transport mechanism being JSON, ints will be turned into float64s.
func (u *UserClient) GetData(collection_id string, query *Query) (map[string]interface{}, error) {
	return getdata(u, collection_id, query)
}

//GetDataByName performs a query against a collection, using the collection's name, rather than the ID. The query object is discussed elsewhere. If the query object is nil, then it will return all of the data.
//The return value is a key-value of the types. Note that due to the transport mechanism being JSON, ints will be turned into float64s.
func (u *UserClient) GetDataByName(collectionName string, query *Query) (map[string]interface{}, error) {
	return getDataByName(u, u.SystemKey, collectionName, query)
}

//GetDataByName performs a query against a collection, using the collection's name, rather than the ID. The query object is discussed elsewhere. If the query object is nil, then it will return all of the data.
//The return value is a key-value of the types. Note that due to the transport mechanism being JSON, ints will be turned into float64s.

func (d *DevClient) GetDataByName(collectionName string, query *Query) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Developer cannot call this yet")
}

//GetData performs a query against a collection. The query object is discussed elsewhere. If the query object is nil, then it will return all of the data.
//The return value is a key-value of the types. Note that due to the transport mechanism being JSON, ints will be turned into float64s.
func (d *DevClient) GetData(collection_id string, query *Query) (map[string]interface{}, error) {
	return getdata(d, collection_id, query)
}

func (d *UserClient) GetDataTotal(collection_id string, query *Query) (map[string]interface{}, error) {
	return getdatatotal(d, collection_id, query)

}
func (u *UserClient) GetItemCount(collection_id string) (int, error) {
	return getItemCount(u, collection_id)
}

func (d *DevClient) GetItemCount(collection_id string) (int, error) {
	return getItemCount(d, collection_id)
}

func getItemCount(c cbClient, collection_id string) (int, error) {
	creds, err := c.credentials()
	if err != nil {
		return -1, err
	}
	resp, err := get(c, _DATA_PREAMBLE+collection_id+"/count", nil, creds, nil)
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

func getDataByName(c cbClient, sysKey string, collectionName string, query *Query) (map[string]interface{}, error) {
	creds, err := c.credentials()
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
	resp, err := get(c, _DATA_NAME_PREAMBLE+sysKey+"/"+collectionName, qry, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting data: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting data: %v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func getdata(c cbClient, collection_id string, query *Query) (map[string]interface{}, error) {
	creds, err := c.credentials()
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
	resp, err := get(c, _DATA_PREAMBLE+collection_id, qry, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting data: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting data: %v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func getdatatotal(c cbClient, collection_id string, query *Query) (map[string]interface{}, error) {

	creds, err := c.credentials()
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
	resp, err := get(c, _DATA_PREAMBLE+collection_id+"/count", qry, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting data: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting data: %v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

//UpdateData mutates the values in extant rows, selecting them via a query. If the query is nil, it updates all rows
//changes should be a map of the names of the columns, and the value you want them updated to

func (u *UserClient) UpdateData(collection_id string, query *Query, changes map[string]interface{}) error {
	err := updatedata(u, collection_id, query, changes)
	return err
}

//UpdateData mutates the values in extant rows, selecting them via a query. If the query is nil, it updates all rows
//changes should be a map of the names of the columns, and the value you want them updated to
func (d *DevClient) UpdateData(collection_id string, query *Query, changes map[string]interface{}) error {
	err := updatedata(d, collection_id, query, changes)
	return err
}

func updatedata(c cbClient, collection_id string, query *Query, changes map[string]interface{}) error {
	qry := query.serialize()
	body := map[string]interface{}{
		"query": qry,
		"$set":  changes,
	}
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := put(c, _DATA_PREAMBLE+collection_id, body, creds, nil)
	if err != nil {
		return fmt.Errorf("Error updating data: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating data: %v", resp.Body)
	}
	return nil
}

//DeleteData removes data from a collection according to what matches the query. If the query is nil, then all data will be removed.
func (u *UserClient) DeleteData(collection_id string, query *Query) error {
	return deletedata(u, collection_id, query)
}

//DeleteData removes data from a collection according to what matches the query. If the query is nil, then all data will be removed.
func (d *DevClient) DeleteData(collection_id string, query *Query) error {
	return deletedata(d, collection_id, query)
}

func deletedata(c cbClient, collection_id string, query *Query) error {
	creds, err := c.credentials()
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
	resp, err := delete(c, _DATA_PREAMBLE+collection_id, qry, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting data: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting data: %v", resp.Body)
	}
	return nil
}

//GetColumns gets a slice of map[string]interface{} of the column names and values.
//As map[string]interface{}{"ColumnName":"name","ColumnType":"typename in string", "PK":bool}
func (d *DevClient) GetColumns(collectionId, systemKey, systemSecret string) ([]interface{}, error) {
	return getColumns(d, collectionId, systemKey, systemSecret)
}

//GetColumns gets a slice of map[string]interface{} of the column names and values.
//As map[string]interface{}{"ColumnName":"name","ColumnType":"typename in string", "PK":bool}
func (u *UserClient) GetColumns(collection_id string) ([]interface{}, error) {
	return getColumns(u, collection_id, "", "")
}

func getColumns(c cbClient, collection_id, systemKey, systemSecret string) ([]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}

	var headers map[string][]string = nil
	if systemKey != "" {
		headers = map[string][]string{
			"Clearblade-Systemkey":    []string{systemKey},
			"Clearblade-Systemsecret": []string{systemSecret},
		}
	}

	resp, err := get(c, _DATA_PREAMBLE+collection_id+"/columns", nil, creds, headers)
	if err != nil {
		return nil, fmt.Errorf("Error getting collection columns: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting collection columns: %v", resp.Body)
	}
	return resp.Body.([]interface{}), nil
}

//GetDataByKeyAndName is unimplemented
func (d *DevClient) GetDataByKeyAndName(string, string, *Query) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Unimplemented")
}

//GetDataByKeyAndName is unimplemented
func (d *UserClient) GetDataByKeyAndName(string, string, *Query) (map[string]interface{}, error) {
	return nil, fmt.Errorf("Unimplemented")

}
