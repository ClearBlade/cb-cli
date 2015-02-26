package main

import (
	"fmt"
	cb "github.com/clearblade/Go-SDK"
	// "os"
	// "os/user"
)

func CreateSystem(cli *cb.DevClient, meta *System_meta) (string, error) {
	fmt.Println("Creating system...")
	result, err := cli.NewSystem(meta.Name, meta.Description, true)
	if err != nil {
		return "", err
	}
	return result, nil
}

func CreateCollections(cli *cb.DevClient, sysKey string, meta []Collection_meta) ([]Collection_meta, error) {
	fmt.Println("Creating collections...")
	newCollections := make([]Collection_meta, len(meta))
	for i := 0; i < len(meta); i++ {
		collID, err := cli.NewCollection(sysKey, meta[i].Name)
		if err != nil {
			return nil, err
		}
		newCollections[i] = Collection_meta{
			Collection_id: collID,
			Columns:       meta[i].Columns,
		}
		for j := 0; j < len(meta[i].Columns); j++ {
			if meta[i].Columns[j].ColumnName != "item_id" {
				err := cli.AddColumn(collID, meta[i].Columns[j].ColumnName, meta[i].Columns[j].ColumnType)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return newCollections, nil
}

func CreateServices(cli *cb.DevClient, sysKey string, services []cb.Service) error {
	fmt.Println("Creating services...")
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
	fmt.Println("Creating roles...")
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

func getNilValue(columns []Column, columnName string) interface{} {
	nilValueMap := map[string]interface{}{
		"string":    "",
		"int":       0,
		"float":     0,
		"uuid":      "00000000-0000-0000-0000-000000000000",
		"bigint":    0,
		"blob":      "",
		"bool":      false,
		"double":    0,
		"timestamp": 0,
	}
	for i := 0; i < len(columns); i++ {
		if columns[i].ColumnName == columnName {
			return nilValueMap[columns[i].ColumnType]
		}
	}
	var thing interface{}
	return thing
}

func MigrateRows(cli *cb.DevClient, oldSystemMeta *System_meta, newSysKey string, oldCollections, newCollections []Collection_meta) error {
	fmt.Println("Migrating items...")
	investigatorQuery := new(cb.Query)
	investigatorQuery.PageNumber = 1
	investigatorQuery.PageSize = 0
	var oldSystemCli *cb.DevClient

	if URL != oldSystemMeta.PlatformUrl {
		cb.CB_ADDR = oldSystemMeta.PlatformUrl
		fmt.Println("Please enter your credentials for your old system")
		email, pass, err := auth_prompt()
		if err != nil {
			return err
		}
		oldSystemCli = cb.NewDevClient(email, pass)
		err = oldSystemCli.Authenticate()
		if err != nil {
			return err
		}
	} else {
		oldSystemCli = cli
	}

	for i := 0; i < len(oldCollections); i++ {
		cb.CB_ADDR = oldSystemMeta.PlatformUrl
		data, err := oldSystemCli.GetData(oldCollections[i].Collection_id, investigatorQuery)
		if err != nil {
			return err
		}
		totalItems := data["TOTAL"].(float64)

		for j := 0; j < int(totalItems); j += importPageSize {
			cb.CB_ADDR = oldSystemMeta.PlatformUrl
			currentQuery := new(cb.Query)
			currentQuery.PageNumber = (j / importPageSize) + 1
			currentQuery.PageSize = importPageSize
			data, err := oldSystemCli.GetData(oldCollections[i].Collection_id, currentQuery)
			if err != nil {
				return err
			}

			typedData := data["DATA"].([]interface{})

			for k := 0; k < len(typedData); k++ {
				delete(typedData[k].(map[string]interface{}), "item_id")
				for key, val := range typedData[k].(map[string]interface{}) {
					if val == nil {
						typedData[k].(map[string]interface{})[key] = getNilValue(newCollections[i].Columns, key)
					}

				}
			}
			cb.CB_ADDR = URL
			err = cli.InsertData(newCollections[i].Collection_id, data["DATA"].([]interface{}))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
