package GoSDK

import (
	"errors"
	"fmt"
	"strings"
	//	"log"
)

const (
	_DEV_HEADER_KEY = "ClearBlade-DevToken"
	_DEV_PREAMBLE   = "/admin"
)

//System is a collection of facts about a system
type System struct {
	Key         string
	Secret      string
	Name        string
	Description string
	Users       bool
	TokenTTL    int32
}

const (
	PERM_READ   = 1
	PERM_CREATE = 2
	PERM_UPDATE = 4
	PERM_DELETE = 8
)

//NewSystem creates a new system. The users parameter has been depreciated. Returned is the systemid.
func (d *DevClient) NewSystem(name, description string, users bool) (string, error) {
	creds, err := d.credentials()
	if err != nil {
		return "", err
	}
	resp, err := post(d, d.preamble()+"/systemmanagement", map[string]interface{}{
		"name":          name,
		"description":   description,
		"auth_required": users,
	}, creds, nil)
	if err != nil {
		return "", fmt.Errorf("Error creating new system: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error Creating new system: %v", resp.Body)
	}

	switch resp.Body.(type) {
	case string:
		b := resp.Body.(string)
		s := strings.Split(b, ":")
		if len(s) < 1 {
			return "", fmt.Errorf("Error creating new system: Empty response")
		}
		return strings.TrimSpace(s[1]), nil
	case map[string]interface{}:
		b := resp.Body.(map[string]interface{})
		id, ok := b["appID"].(string)
		if !ok {
			return "", fmt.Errorf("Error creating new system: Missing appID")
		}
		return id, nil
	default:
		return "", fmt.Errorf("Error creating new system: Incorrect return type: %T\n", resp.Body)
	}
}

//GetSystem retrieves information about the system specified.
func (d *DevClient) GetSystem(key string) (*System, error) {
	creds, err := d.credentials()
	if err != nil {
		return &System{}, err
	} else if len(creds) != 1 {
		return nil, fmt.Errorf("Error getting system: No DevToken Supplied")
	}
	sysResp, sysErr := get(d, d.preamble()+"/systemmanagement", map[string]string{"id": key}, creds, nil)
	if sysErr != nil {
		return nil, fmt.Errorf("Error gathering system information: %v", sysErr)
	}
	if sysResp.StatusCode != 200 {
		return nil, fmt.Errorf("Error gathering system information: %v", sysResp.Body)
	}
	sysMap, isMap := sysResp.Body.(map[string]interface{})
	if !isMap {
		return nil, fmt.Errorf("Error gathering system information: incorrect return type\n")
	}
	newSys := &System{
		Key:         sysMap["appID"].(string),
		Secret:      sysMap["appSecret"].(string),
		Name:        sysMap["name"].(string),
		Description: sysMap["description"].(string),
		TokenTTL:    int32(sysMap["token_ttl"].(float64)),
	}
	return newSys, nil

}

//DeleteSystem removes the system
func (d *DevClient) DeleteSystem(s string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, d.preamble()+"/systemmanagement", map[string]string{"id": s}, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting system: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting system: %v", resp.Body)
	}
	return nil
}

//SetSystemName can change the name of the system
func (d *DevClient) SetSystemName(system_key, system_name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := put(d, d.preamble()+"/systemmanagement", map[string]interface{}{
		"id":   system_key,
		"name": system_name,
	}, creds, nil)
	if err != nil {
		return fmt.Errorf("Error changing system name: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error changing system name: %v", resp.Body)
	}
	return nil
}

//SetSystemDescription can change the content of the system's description
func (d *DevClient) SetSystemDescription(system_key, system_description string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := put(d, d.preamble()+"/systemmanagement", map[string]interface{}{
		"id":          system_key,
		"description": system_description,
	}, creds, nil)
	if err != nil {
		return fmt.Errorf("Error changing system description: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error changing system description: %v", resp.Body)
	}
	return nil
}

//SetSystemAuthOn is depreciated
func (d *DevClient) SetSystemAuthOn(system_key string) error {
	return fmt.Errorf("Auth is now mandatory")
}

//SetSystemAuthOff is depreciated
func (d *DevClient) SetSystemAuthOff(system_key string) error {
	return fmt.Errorf("Auth is now mandatory")
}

//DevUserInfo gets the user information for the developer
func (d *DevClient) DevUserInfo() (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, d.preamble()+"/userinfo", nil, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting userdata: %v", err)
	}
	return resp.Body.(map[string]interface{}), nil
}

//NewConnectCollection creates a new collection that is backed by a datastore of your own choosing.
func (d *DevClient) NewConnectCollection(systemkey string, connectConfig connectCollection) (string, error) {
	creds, err := d.credentials()
	m := connectConfig.toMap()
	m["appID"] = systemkey
	m["name"] = connectConfig.tableName()
	if err != nil {
		return "", err
	}
	resp, err := post(d, d.preamble()+"/collectionmanagement", m, creds, nil)
	if err != nil {
		return "", fmt.Errorf("Error creating collection: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error creating collection %v\n", resp.Body)
	}
	return resp.Body.(map[string]interface{})["collectionID"].(string), nil
}

//AlterConnectionDetails allows the developer to change or add connection information, such as updating a username
func (d *DevClient) AlterConnectionDetails(systemkey string, connectConfig connectCollection) error {
	creds, err := d.credentials()
	out := make(map[string]interface{})
	m := connectConfig.toMap()
	out["appID"] = systemkey
	out["name"] = connectConfig.tableName()
	out["connectionStringMap"] = m
	resp, err := put(d, d.preamble()+"/collectionmanagement", out, creds, nil)
	if err != nil {
		return fmt.Errorf("Error creating collection: %s", err.Error())
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("Error creating collection %v\n", resp.Body)
	} else {
		return nil
	}
}

//CreateCollection creates a new collection
func (d *DevClient) NewCollection(systemKey, name string) (string, error) {
	creds, err := d.credentials()
	if err != nil {
		return "", err
	}
	resp, err := post(d, d.preamble()+"/collectionmanagement", map[string]interface{}{
		"name":  name,
		"appID": systemKey,
	}, creds, nil)
	if err != nil {
		return "", fmt.Errorf("Error creating collection: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error creating collection %v", resp.Body)
	}
	return resp.Body.(map[string]interface{})["collectionID"].(string), nil
}

//DeleteCollection deletes the collection. Note that this does not apply to collections backed by a non-default datastore.
func (d *DevClient) DeleteCollection(colId string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, d.preamble()+"/collectionmanagement", map[string]string{
		"id": colId,
	}, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting collection %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting collection %v", resp.Body)
	}
	return nil
}

//AddColumn adds a column to a collection. Note that this does not apply to collections backed by a non-default datastore.
func (d *DevClient) AddColumn(collection_id, column_name, column_type string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := put(d, d.preamble()+"/collectionmanagement", map[string]interface{}{
		"id": collection_id,
		"addColumn": map[string]interface{}{
			"name": column_name,
			"type": column_type,
		},
	}, creds, nil)
	if err != nil {
		return fmt.Errorf("Error adding column: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error adding column: %v", resp.Body)
	}
	return nil
}

//DeleteColumn removes a column from a collection. Note that this does not apply to collections backed by a non-default datastore.
func (d *DevClient) DeleteColumn(collection_id, column_name string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := put(d, d.preamble()+"/collectionmanagement", map[string]interface{}{
		"id":           collection_id,
		"deleteColumn": column_name,
	}, creds, nil)
	if err != nil {
		return fmt.Errorf("Error deleting column: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting column: %v", resp.Body)
	}
	return nil
}

//GetCollectionInfo retrieves some describing information on the specified collection
//Keys "name","collectoinID","appID"
func (d *DevClient) GetCollectionInfo(collection_id string) (map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return map[string]interface{}{}, err
	}
	resp, err := get(d, d.preamble()+"/collectionmanagement", map[string]string{
		"id": collection_id,
	}, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting collection info: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting collection info: %v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

//GetAllCollections retrieves a list of every collection in the system
//The return value is a slice of strings
func (d *DevClient) GetAllCollections(SystemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, d.preamble()+"/allcollections", map[string]string{
		"appid": SystemKey,
	}, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting collection info: %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting collection info: %v", resp.Body)
	}

	return resp.Body.([]interface{}), nil
}

//GetAllRoles returns a slice of all roles, including their permissions
//the return value is a slice of [{"ID":"roleid","Name":"rolename","Description":"role description", "Permissions":{"Collections":[{"ID":"collectionid","Columns":[{"Name":"columnname","Level":0}],"Items":[{"Name":"itemid","Level":2}],"Name":"collectionname"}], "Topics":[{"Name":"topic/path","Level":1}],"CodeServices":[{"Name":"service name","SystemKey":"syskey","Level":4}],"UsersList":{"Name":"users","Level":8},"Push":{"Name":"push","Level":0},"MsgHistory":{"Name":"messagehistory","Level":1}}},...]
func (d *DevClient) GetAllRoles(SystemKey string) ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, d.preamble()+"/user/"+SystemKey+"/roles", map[string]string{
		"appid": SystemKey,
	}, creds, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get all roles: '%s'\n", err.Error())
	}

	rval, ok := resp.Body.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Bad type returned by GetAllRoles")
	}

	return rval, nil
}

//CreateRole creates a new role
//returns a JSON object shaped like {"role_id":"role id goes here"}
func (d *DevClient) CreateRole(systemKey, role_id string) (interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"name":        role_id,
		"collections": []map[string]interface{}{},
		"topics":      []map[string]interface{}{},
		"services":    []map[string]interface{}{},
	}
	resp, err := post(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error updating a role to have a collection: %v", resp.Body)
	}
	return resp.Body, nil
}

func (d *DevClient) UpdateRole(systemKey, roleName string, role map[string]interface{}) error {
	data := map[string]interface{}{
		"name":        role["Name"],
		"collections": []map[string]interface{}{},
		"topics":      []map[string]interface{}{},
		"services":    []map[string]interface{}{},
	}
	if collections, ok := role["Collections"]; ok {
		data["collections"] = collections
	}
	if topics, ok := role["Topics"]; ok {
		data["topics"] = topics
	}
	if services, ok := role["Services"]; ok {
		data["services"] = services
	}
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	// resp, err := post(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	resp, err := put(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating role %s", roleName)
	}
	return nil
}

//DeleteRole removes a role
func (d *DevClient) DeleteRole(systemKey, roleId string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, d.preamble()+"/user/"+systemKey+"/roles", map[string]string{"role": roleId}, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting role: %v", resp.Body)
	}
	return nil
}

func (d *DevClient) GetAllUsers(systemKey string) ([]map[string]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}

	/*
		allQuery := NewQuery()
		queryMap := allQuery.serialize()
		queryBytes, err := json.Marshal(queryMap)
	*/
	if err != nil {
		return nil, err
	}
	resp, err := get(d, d.preamble()+"/user/"+systemKey, nil, creds, nil)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting all users: %v", resp.Body)
	}
	dbResponse := resp.Body.(map[string]interface{})
	rawData := dbResponse["Data"].([]interface{})

	rval := make([]map[string]interface{}, len(rawData))
	for idx, oneRsp := range rawData {
		rval[idx] = oneRsp.(map[string]interface{})
	}

	return rval, nil
}

//DeleteUser removes a specific user
func (d *DevClient) DeleteUser(systemKey, userId string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(d, d.preamble()+"/user/"+systemKey, map[string]string{"user": userId}, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error deleting user: %v", resp.Body)
	}

	return nil
}

func (d *DevClient) UpdateUser(systemKey, userId string, info map[string]interface{}) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"user":    userId,
		"changes": info,
	}
	resp, err := put(d, d.preamble()+"/user/"+systemKey, data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating user: %v", resp.Body)
	}
	return nil
}

//AddUserToRoles assigns a role to a user
func (d *DevClient) AddUserToRoles(systemKey, userId string, roles []string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"user": userId,
		"changes": map[string]interface{}{
			"roles": map[string]interface{}{
				"add": roles,
			},
		},
	}
	resp, err := put(d, d.preamble()+"/user/"+systemKey, data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error adding roles to a user: %v", resp.Body)
	}

	return nil
}

func (d *DevClient) GetUserRoles(systemKey, userId string) ([]string, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, d.preamble()+"/user/"+systemKey+"/roles", map[string]string{"user": userId}, creds, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting roles for a user: %v", resp.Body)
	}
	rawBody := resp.Body.([]interface{})
	rval := make([]string, len(rawBody))
	for idx, oneBody := range rawBody {
		oneMap := oneBody.(map[string]interface{})
		rval[idx] = oneMap["Name"].(string)
	}
	return rval, nil
}

//AddCollectionToRole associates some kind of permission regarding the collection to the role.
func (d *DevClient) AddCollectionToRole(systemKey, collection_id, role_id string, level int) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id": role_id,
		"changes": map[string]interface{}{
			"collections": []map[string]interface{}{
				map[string]interface{}{
					"itemInfo": map[string]interface{}{
						"id": collection_id,
					},
					"permissions": level,
				},
			},
			"topics":   []map[string]interface{}{},
			"services": []map[string]interface{}{},
		},
	}
	resp, err := put(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating a role to have a collection: %v", resp.Body)
	}
	return nil
}

//AddServiceToRole associates some kind of permission dealing with the specified service to the role
func (d *DevClient) AddServiceToRole(systemKey, service, role_id string, level int) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id": role_id,
		"changes": map[string]interface{}{
			"services": []map[string]interface{}{
				map[string]interface{}{
					"itemInfo": map[string]interface{}{
						"name": service,
					},
					"permissions": level,
				},
			},
			"topics":      []map[string]interface{}{},
			"collections": []map[string]interface{}{},
		},
	}
	resp, err := put(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating a role to have a service: %v", resp.Body)
	}
	return nil
}

func (d *DevClient) AddGenericPermissionToRole(systemKey, role_id, permission string, level int) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"id": role_id,
		"changes": map[string]interface{}{
			"services":    []map[string]interface{}{},
			"topics":      []map[string]interface{}{},
			"collections": []map[string]interface{}{},
		},
	}

	data["changes"].(map[string]interface{})[permission] = map[string]interface{}{
		"permissions": level,
	}

	resp, err := put(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating a role to have a service: %v", resp.Body)
	}
	return nil
}

func (d *DevClient) credentials() ([][]string, error) {
	if d.DevToken != "" {
		return [][]string{
			[]string{
				_DEV_HEADER_KEY,
				d.DevToken,
			},
		}, nil
	} else {
		return [][]string{}, errors.New("No SystemSecret/SystemKey combo, or UserToken found")
	}
}

func (d *DevClient) preamble() string {
	return _DEV_PREAMBLE
}

func (d *DevClient) getSystemInfo() (string, string) {
	return "", ""
}

func (d *DevClient) setToken(t string) {
	d.DevToken = t
}
func (d *DevClient) getToken() string {
	return d.DevToken
}

func (d *DevClient) getMessageId() uint16 {
	return uint16(d.mrand.Int())
}
