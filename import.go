package main

import (
	// "fmt"
	cb "github.com/clearblade/Go-SDK"
	// "os"
	// "os/user"
)

func CreateSystem(cli *cb.DevClient, meta *System_meta) (string, error) {
	result, err := cli.NewSystem(meta.Name, meta.Description, true)
	if err != nil {
		return "", err
	}
	return result, nil
}

func CreateCollections(cli *cb.DevClient, sysKey string, meta []Collection_meta) error {
	for i := 0; i < len(meta); i++ {
		collID, err := cli.NewCollection(sysKey, meta[i].Name)
		if err != nil {
			return err
		}
		for j := 0; j < len(meta[i].Columns); j++ {
			if meta[i].Columns[j].ColumnName != "item_id" {
				err := cli.AddColumn(collID, meta[i].Columns[j].ColumnName, meta[i].Columns[j].ColumnType)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func CreateServices(cli *cb.DevClient, sysKey string, services []cb.Service) error {
	for i := 0; i < len(services); i++ {
		err := cli.NewService(sysKey, services[i].Name, services[i].Code, services[i].Params)
		if err != nil {
			return err
		}
	}
	return nil
}

func isCustomRole(role_id string) bool {
	switch role_id {
	case "Authenticated":
		return false
	case "Administrator":
		return false
	case "Anonymous":
		return false
	default:
		return true
	}
}

func AddGenericPermission(cli *cb.DevClient, sysKey, role, permission string, level int) error {
	err := cli.AddGenericPermissionToRole(sysKey, role, permission, level)
	if err != nil {
		return err
	}
	return nil
}

func CreateCustomRole(cli *cb.DevClient, sysKey, role_id string) (string, error) {
	resp, err := cli.CreateRole(sysKey, role_id)
	if err != nil {
		return "", err
	}
	return resp.(map[string]interface{})["role_id"].(string), nil
}

func CreateRoles(cli *cb.DevClient, sysKey string, roles []interface{}) error {
	for i := 0; i < len(roles); i++ {
		roleID := roles[i].(map[string]interface{})["ID"].(string)
		if isCustomRole(roles[i].(map[string]interface{})["Name"].(string)) {
			resp, err := CreateCustomRole(cli, sysKey, roles[i].(map[string]interface{})["Name"].(string))
			if err != nil {
				return err
			}
			roleID = resp
		}
		for k, v := range roles[i].(map[string]interface{})["Permissions"].(map[string]interface{}) {
			switch k {
			case "CodeServices":
				for j := 0; j < len(v.([]interface{})); j++ {
					err := cli.AddServiceToRole(sysKey, v.([]interface{})[j].(map[string]interface{})["Name"].(string), roleID, int(v.([]interface{})[j].(map[string]interface{})["Level"].(float64)))
					if err != nil {
						return err
					}
				}
			case "Collections":
				for j := 0; j < len(v.([]interface{})); j++ {
					if roles[i].(map[string]interface{})["Name"].(string) != "Administrator" {
						err := cli.AddCollectionToRole(sysKey, v.([]interface{})[j].(map[string]interface{})["ID"].(string), roleID, int(v.([]interface{})[j].(map[string]interface{})["Level"].(float64)))
						if err != nil {
							return err
						}
					}
				}
			case "MsgHistory":
				err := AddGenericPermission(cli, sysKey, roleID, "msgHistory", int(v.(map[string]interface{})["Level"].(float64)))
				if err != nil {
					return err
				}
			case "Push":
				err := AddGenericPermission(cli, sysKey, roleID, "push", int(v.(map[string]interface{})["Level"].(float64)))
				if err != nil {
					return err
				}
			case "UsersList":
				err := AddGenericPermission(cli, sysKey, roleID, "users", int(v.(map[string]interface{})["Level"].(float64)))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
