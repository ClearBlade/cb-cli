package main

import (
	// "fmt"
	cb "github.com/clearblade/Go-SDK"
	// "io/ioutil"
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

func CreateGenericRole(cli *cb.DevClient, sysKey, role, permission string, level int) error {
	err := cli.AddGenericRole(sysKey, role, permission, level)
	if err != nil {
		return err
	}
	return nil
}

func CreateRoles(cli *cb.DevClient, sysKey string, roles []interface{}) error {
	for i := 0; i < len(roles); i++ {
		for k, v := range roles[i].(map[string]interface{})["Permissions"].(map[string]interface{}) {
			switch k {
			case "CodeServices":
				for j := 0; j < len(v.([]interface{})); j++ {
					err := cli.AddServiceToRole(sysKey, v.([]interface{})[j].(map[string]interface{})["Name"].(string), roles[i].(map[string]interface{})["ID"].(string), int(v.([]interface{})[j].(map[string]interface{})["Level"].(float64)))
					if err != nil {
						return err
					}
				}
			case "Collections":
				for j := 0; j < len(v.([]interface{})); j++ {
					if roles[i].(map[string]interface{})["ID"].(string) != "Administrator" {
						err := cli.AddCollectionToRole(sysKey, v.([]interface{})[j].(map[string]interface{})["ID"].(string), roles[i].(map[string]interface{})["ID"].(string), int(v.([]interface{})[j].(map[string]interface{})["Level"].(float64)))
						if err != nil {
							return err
						}
					}
				}
			case "MsgHistory":
				err := CreateGenericRole(cli, sysKey, roles[i].(map[string]interface{})["ID"].(string), "msgHistory", int(v.(map[string]interface{})["Level"].(float64)))
				if err != nil {
					return err
				}
			case "Push":
				err := CreateGenericRole(cli, sysKey, roles[i].(map[string]interface{})["ID"].(string), "push", int(v.(map[string]interface{})["Level"].(float64)))
				if err != nil {
					return err
				}
			case "UsersList":
				err := CreateGenericRole(cli, sysKey, roles[i].(map[string]interface{})["ID"].(string), "users", int(v.(map[string]interface{})["Level"].(float64)))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
