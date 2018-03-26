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

//SetSystemTokenTTL can change the value for the system's token TTL
func (d *DevClient) SetSystemTokenTTL(system_key string, token_ttl int) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	resp, err := put(d, d.preamble()+"/systemmanagement", map[string]interface{}{
		"id":        system_key,
		"token_ttl": token_ttl,
	}, creds, nil)
	if err != nil {
		return fmt.Errorf("Error changing system token TTL: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error changing system token TTL: %v", resp.Body)
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
	m["name"] = connectConfig.name()
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
	out["name"] = connectConfig.name()
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
		return nil, fmt.Errorf("Bad type returned by GetAllRoles: %T, %s", resp.Body, resp.Body.(string))
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
		"name": roleName,
		"changes": map[string]interface{}{
			"collections": []map[string]interface{}{},
			"topics":      []map[string]interface{}{},
			"services":    []map[string]interface{}{},
			"portals":     []map[string]interface{}{},
		},
	}
	changes := data["changes"].(map[string]interface{})

	if roleId, ok := role["ID"].(string); ok {
		data["id"] = roleId
	} else {
		return fmt.Errorf("The role id key (ID) must be present to update the role")
	}
	permissions, ok := role["Permissions"].(map[string]interface{})
	if !ok {
		permissions = map[string]interface{}{}
	}
	if collections, ok := permissions["collections"]; ok {
		changes["collections"] = collections
	}
	if topics, ok := permissions["topics"]; ok {
		changes["topics"] = topics
	}
	if services, ok := permissions["services"]; ok {
		changes["services"] = services
	}
	if portals, ok := permissions["portals"]; ok {
		changes["portals"] = portals
	}
	if msgHist, ok := permissions["msgHistory"]; ok {
		changes["msgHistory"] = msgHist
	}
	if deviceList, ok := permissions["devices"]; ok {
		changes["devices"] = deviceList
	}
	if userList, ok := permissions["users"]; ok {
		changes["users"] = userList
	}
	if allservices, ok := permissions["allservices"]; ok {
		changes["allservices"] = allservices
	}
	if allcollections, ok := permissions["allcollections"]; ok {
		changes["allcollections"] = allcollections
	}

	// Just to be safe, this is silly
	data["changes"] = changes
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	resp, err := put(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating role %s: %d", roleName, resp.StatusCode)
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

//AddDeviceToRoles assigns a role to a device
func (d *DevClient) AddDeviceToRoles(systemKey, deviceName string, roles []string) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]interface{}{"add": roles}
	resp, err := put(d, d.preamble()+"/devices/roles/"+systemKey+"/"+deviceName, data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error adding roles to a device: %v", resp.Body)
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

func (d *DevClient) AddPortalToRole(systemKey, portalName, roleId string, level int) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id": roleId,
		"changes": map[string]interface{}{
			"portals": []map[string]interface{}{
				map[string]interface{}{
					"itemInfo": map[string]interface{}{
						"name": portalName,
					},
					"permissions": level,
				},
			},
		},
	}
	resp, err := put(d, d.preamble()+"/user/"+systemKey+"/roles", data, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error updating a role to have a portal: %v", resp.Body)
	}
	return nil
}

//AddServiceToRole associates some kind of permission dealing with the specified service to the role
func (d *DevClient) AddServiceToRole(systemKey, service, roleId string, level int) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id": roleId,
		"changes": map[string]interface{}{
			"services": []map[string]interface{}{
				map[string]interface{}{
					"itemInfo": map[string]interface{}{
						"name": service,
					},
					"permissions": level,
				},
			},
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

func (d *DevClient) AddGenericPermissionToRole(systemKey, roleId, permission string, level int) error {
	creds, err := d.credentials()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"id":      roleId,
		"changes": map[string]interface{}{},
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
