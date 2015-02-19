package main

import (
	"fmt"
	//

	cb "github.com/clearblade/Go-SDK"
	// "io/ioutil"
	// "os"
	// "os/user"
)

func CreateSystem(cli *cb.DevClient, meta *System_meta) error {
	result, err := cli.NewSystem(meta.Name, meta.Description, true)
	if err != nil {
		return err
	}
	fmt.Printf("result is: %s", result)
	return nil
}

func CreateCollection(meta *System_meta) error {
	return nil
}

func CreateServices(cli *cb.DevClient, meta *System_meta, services []*cb.Service) error {
	return nil
}

func CreateRoles(meta *System_meta) error {
	return nil
}
