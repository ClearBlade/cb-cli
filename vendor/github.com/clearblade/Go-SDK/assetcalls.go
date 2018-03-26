package GoSDK

func (d *DevClient) GetSystemAssetDeployments(systemKey string, query *Query) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	qry, err := createQueryMap(query)
	if err != nil {
		return nil, err
	}

	resp, err := get(d, "/admin/"+systemKey+"/deploy_assets", qry, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}

	return resp.Body.([]interface{}), nil
}

func (d *DevClient) GetAssetClassDeployments(systemKey, assetClass string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := get(d, "/admin/"+systemKey+"/deploy_assets/"+assetClass, nil, creds, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) UpdateAssetClassDeployments(systemKey, assetClass string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := put(d, "/admin/"+systemKey+"/deploy_assets/"+assetClass, data, creds, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetAssetDeployments(systemKey, assetClass, assetId string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := get(d, "/admin/"+systemKey+"/deploy_assets/"+assetClass+"/"+assetId, nil, creds, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) UpdateAssetDeployments(systemKey, assetClass, assetId string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := put(d, "/admin/"+systemKey+"/deploy_assets/"+assetClass+"/"+assetId, data, creds, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetAssetsDeployedToEntity(systemKey, entityType, entityName string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	headers := map[string][]string{entityType: []string{entityName}}

	resp, err := put(d, "/admin/"+systemKey+"/deployed_assets", nil, creds, headers)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetAssetsNotDeployedOnPlatform(systemKey string, query *Query) ([]interface{}, error) {

	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	qry, err := createQueryMap(query)
	if err != nil {
		return nil, err
	}

	resp, err := get(d, "/admin/"+systemKey+"/deploy_on_platform", qry, creds, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (d *DevClient) GetAssetPlatformDeploymentStatus(systemKey, assetClass, assetId string) (map[string]interface{}, error) {

	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	resp, err := get(d, "/admin/"+systemKey+"/deploy_on_platform/"+assetClass+"/"+assetId, nil, creds, nil)
	if err != nil {
		return nil, err
	}

	return resp.Body.(map[string]interface{}), err
}

func (d *DevClient) UpdateAssetPlatformDeploymentStatus(systemKey, assetClass, assetId string, deploy bool) (map[string]interface{}, error) {

	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{"deploy": deploy}
	resp, err := mapResponse(put(d, "/admin/"+systemKey+"/deploy_on_platform/"+assetClass+"/"+assetId, reqBody, creds, nil))
	if err != nil {
		return nil, err
	}

	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetAllDeployments(systemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(get(d, "/admin/"+systemKey+"/deployments", nil, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

func (u *UserClient) GetAllDeployments(systemKey string) (map[string]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(get(u, "/api/v/3/"+systemKey+"/deployments", nil, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) GetDeploymentByName(systemKey, name string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(get(d, "/admin/"+systemKey+"/deployments/"+name, nil, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) GetDeploymentByName(systemKey, name string) (map[string]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(get(u, "/api/v/3/"+systemKey+"/deployments/"+name, nil, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) CreateDeploymentByName(systemKey, name string, info map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(post(d, "/admin/"+systemKey+"/deployments", info, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) CreateDeploymentByName(systemKey, name string, info map[string]interface{}) (map[string]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(post(u, "/api/v/3/"+systemKey+"/deployments", info, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) UpdateDeploymentByName(systemKey, name string, changes map[string]interface{}) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(put(d, "/admin/"+systemKey+"/deployments/"+name, changes, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (u *UserClient) UpdateDeploymentByName(systemKey, name string, changes map[string]interface{}) (map[string]interface{}, error) {
	creds, err := u.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := mapResponse(put(u, "/api/v/3/"+systemKey+"/deployments/"+name, changes, creds, nil))
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

func (d *DevClient) DeleteDeploymentByName(systemKey, name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	_, err = mapResponse(delete(d, "/admin/"+systemKey+"/deployments/"+name, nil, creds, nil))
	return err
}

func (u *UserClient) DeleteDeploymentByName(systemKey, name string) error {
	creds, err := u.credentials()
	if err != nil {
		return err
	}
	_, err = mapResponse(delete(u, "/api/v/3/"+systemKey+"/deployments/"+name, nil, creds, nil))
	return err
}
