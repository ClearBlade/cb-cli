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

func CreateRoles(meta *System_meta) error {
	return nil
}
