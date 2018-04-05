package cblib

import (
	//"flag"
	"fmt"
	cb "github.com/clearblade/Go-SDK"
	"os"
	"strings"
)

func init() {

	usage := 
	`
	Initializes your filesystem with your ClearBlade Platform System.

	A new folder is created with same name as you system in your current directory.
	`

	example := 
	`	
	cb-cli init 																					# default init, will prompt for creds	
	cb-cli init -url=https://platform.clearblade.com -system-key=9b9eea9c0bda8896a3dab5aeec9601     # prompts for email, password
	`

	systemDotJSON = map[string]interface{}{}
	svcCode = map[string]interface{}{}
	rolesInfo = []map[string]interface{}{}
	myInitCommand := &SubCommand{
		name:            "init",
		usage:           usage,
		needsAuth:       true,
		mustBeInRepo:    false,
		mustNotBeInRepo: false,
		run:             doInit,
		example:		 example,
	}
	myInitCommand.flags.StringVar(&URL, "url", "https://platform.clearblade.com", "Clearblade Platform URL where system is hosted, ex https://platform.clearblade.com")
	myInitCommand.flags.StringVar(&MsgURL, "messaging-url", "platform.clearblade.com", "Clearblade messaging url for target system, ex platform.clearblade.com")
	myInitCommand.flags.StringVar(&SystemKey, "system-key", "", "System Key for target system, ex 9b9eea9c0bda8896a3dab5aeec9601")
	myInitCommand.flags.StringVar(&Email, "email", "", "Developer email for login")
	myInitCommand.flags.StringVar(&Password, "password", "", "Developer password")
	AddCommand("init", myInitCommand)
}

func doInit(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if len(args) != 0 {
		fmt.Printf("init command takes no arguments; only options: '%+v'\n", args)
		os.Exit(1)
	}
	return reallyInit(client, SystemKey)
}

func reallyInit(cli *cb.DevClient, sysKey string) error {
	sysMeta, err := pullSystemMeta(sysKey, cli)
	if err != nil {
		return err
	}

	if IsInRepo() {
			SetRootDir(".")
	} else {
		SetRootDir(strings.Replace(sysMeta.Name, " ", "_", -1))
		if err := setupDirectoryStructure(sysMeta); err != nil {
			return err
		}
	}
	storeMeta(sysMeta)

	if err = storeSystemDotJSON(systemDotJSON); err != nil {
		return err
	}
	metaStuff := map[string]interface{}{
		"platform_url":        cb.CB_ADDR,
		"messaging_url":       cb.CB_MSG_ADDR,
		"developer_email":     Email,
		"asset_refresh_dates": []interface{}{},
		"token":               cli.DevToken,
	}
	if err = storeCBMeta(metaStuff); err != nil {
		return err
	}

	fmt.Printf("System '%s' has been initialized.\n", sysMeta.Name)
	return nil
}
