package GoSDK

func (d *DevClient) GetEdgeGroups(systemKey string, query *Query) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	qry, err := createQueryMap(query)
	if err != nil {
		return nil, err
	}

	resp, err := get(d, "/admin/"+systemKey+"/edge_groups", qry, creds, nil)
	resp, err = mapResponse(resp, err)

	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) GetEdgeGroup(systemKey, name string, recursive bool) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	var queries map[string]string
	if recursive == true {
		queries = map[string]string{"recursive": "true"}
	}
	resp, err := get(d, "/admin/"+systemKey+"/edge_groups/"+name, queries, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) CreateEdgeGroup(systemKey, name string,
	data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	data["name"] = name
	resp, err := post(d, "/admin/"+systemKey+"/edge_groups", data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) DeleteEdgeGroup(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, "/admin/"+systemKey+"/edge_groups/"+name, nil, creds, nil)
	_, err = mapResponse(resp, err)
	return err
}

func (d *DevClient) UpdateEdgeGroup(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(d, "/admin/"+systemKey+"/edge_groups/"+name, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}
