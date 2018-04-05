package cblib

import (
	"fmt"
	cb "github.com/clearblade/Go-SDK"
)

func init() {

	usage := 
	`
	Pushes an update from local filesystem to the platform, synonymous to #push
	`

	example := 
	`
	cb-cli update -service=Service1					# Pushes a local service up to platform
	`

	updateCommand := &SubCommand{
		name:         "update",
		usage:        usage,
		needsAuth:    true,
		mustBeInRepo: true,
		run:          doUpdate,
		example:	  example,
	}
	updateCommand.flags.StringVar(&ServiceName, "service", "", "Name of service to update")
	updateCommand.flags.StringVar(&LibraryName, "library", "", "Name of library to update")
	updateCommand.flags.StringVar(&CollectionName, "collection", "", "Unique id of collection to update")
	updateCommand.flags.StringVar(&CollectionId, "collectionID", "", "Unique id of collection to update")
	updateCommand.flags.StringVar(&User, "user", "", "Unique id of user to update")
	updateCommand.flags.StringVar(&UserId, "userID", "", "Unique id of user to update")
	updateCommand.flags.StringVar(&RoleName, "role", "", "Name of role to update")
	updateCommand.flags.StringVar(&TriggerName, "trigger", "", "Name of trigger to update")
	updateCommand.flags.StringVar(&TimerName, "timer", "", "Name of timer to update")
	AddCommand("update", updateCommand)
}

func checkUpdateArgsAndFlags(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("There are no arguments to the update command, only command line options\n")
	}
	return nil
}

func doUpdate(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if err := checkUpdateArgsAndFlags(args); err != nil {
		return err
	}
	systemInfo, err := getSysMeta()
	if err != nil {
		return err
	}
	SetRootDir(".")

	// This is a hack to check if token has expired and auth again
	// since we dont have an endpoint to determine this
	client, err = checkIfTokenHasExpired(client, systemInfo.Key)
	if err != nil {
		return fmt.Errorf("Re-auth failed: %s", err)
	}

	didSomething := false

	if ServiceName != "" {
		didSomething = true
		if err := pushOneService(systemInfo, client); err != nil {
			return err
		}
	}

	if LibraryName != "" {
		didSomething = true
		if err := pushOneLibrary(systemInfo, client); err != nil {
			return err
		}
	}

	if CollectionName != "" {
		didSomething = true
		if err := pushOneCollection(systemInfo, client); err != nil {
			return err
		}
	}

	if CollectionId != "" {
		didSomething = true
		if err := pushOneCollectionById(systemInfo, client); err != nil {
			return err
		}
	}

	if User != "" {
		didSomething = true
		if err := pushOneUser(systemInfo, client); err != nil {
			return err
		}
	}

	if UserId != "" {
		didSomething = true
		if err := pushOneUserById(systemInfo, client); err != nil {
			return err
		}
	}

	if RoleName != "" {
		didSomething = true
		if err := pushOneRole(systemInfo, client); err != nil {
			return err
		}
	}

	if TriggerName != "" {
		didSomething = true
		if err := pushOneTrigger(systemInfo, client); err != nil {
			return err
		}
	}

	if TimerName != "" {
		didSomething = true
		if err := pushOneTimer(systemInfo, client); err != nil {
			return err
		}
	}

	if !didSomething {
		fmt.Printf("Nothing to update -- you must specify something to update (ie, -service=<svc_name>)\n")
	}

	return nil
}
