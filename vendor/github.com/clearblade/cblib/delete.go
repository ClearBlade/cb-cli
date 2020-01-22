package cblib

import (
	"fmt"

	cb "github.com/clearblade/Go-SDK"
)

func init() {
	usage :=
		`
	Deletes a specified resource from the remote system on the Platform"
	`

	example :=
		`
	  cb-cli delete -service=ServiceToDelete      # Deletes ServiceToDelete from Platform's System
	  cb-cli delete -collection=CollectonToDelete # Deletes CollectonToDelete from Platform's System
	`
	deleteCommand := &SubCommand{
		name:      "delete",
		usage:     usage,
		needsAuth: true,
		run:       doDelete,
		example:   example,
	}
	deleteCommand.flags.StringVar(&ServiceName, "service", "", "Name of service to delete")
	deleteCommand.flags.StringVar(&LibraryName, "library", "", "Name of library to delete")
	deleteCommand.flags.StringVar(&CollectionId, "collectionId", "", "Collection's ID to delete, ex baa5ad9d0b96ac86a8e8cdb4e1b501")
	deleteCommand.flags.StringVar(&UserId, "userId", "", "Unique id of user to delete, aea5ad9d0bc488f5a0d9d592f58301")
	deleteCommand.flags.StringVar(&RoleName, "role", "", "Name of role to delete")
	deleteCommand.flags.StringVar(&TriggerName, "trigger", "", "Name of trigger to delete")
	deleteCommand.flags.StringVar(&TimerName, "timer", "", "Name of timer to delete")
	deleteCommand.flags.StringVar(&EdgeName, "edge", "", "Name of edge to delete")
	deleteCommand.flags.StringVar(&PortalName, "portal", "", "Name of portal to delete")
	deleteCommand.flags.StringVar(&DeviceName, "device", "", "Name of device to delete")
	AddCommand("delete", deleteCommand)
}

func checkDeleteArgsAndFlags(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("There are no arguments to the delete command, only command line options\n")
	}
	return nil
}

func doDelete(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if err := checkDeleteArgsAndFlags(args); err != nil {
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
		if err := deleteOneService(systemInfo, client); err != nil {
			return err
		}
	}

	if LibraryName != "" {
		didSomething = true
		if err := deleteOneLibrary(systemInfo, client); err != nil {
			return err
		}
	}

	if CollectionId != "" {
		didSomething = true
		if err := deleteOneCollection(systemInfo, client); err != nil {
			return err
		}
	}

	if UserId != "" {
		didSomething = true
		if err := deleteOneUser(systemInfo, client); err != nil {
			return err
		}
	}

	if RoleName != "" {
		didSomething = true
		if err := deleteOneRole(systemInfo, client); err != nil {
			return err
		}
	}

	if TriggerName != "" {
		didSomething = true
		if err := deleteOneTrigger(systemInfo, client); err != nil {
			return err
		}
	}

	if TimerName != "" {
		didSomething = true
		if err := deleteOneTimer(systemInfo, client); err != nil {
			return err
		}
	}

	if EdgeName != "" {
		didSomething = true
		if err := deleteOneEdge(systemInfo, client); err != nil {
			return err
		}
	}

	if PortalName != "" {
		didSomething = true
		if err := deleteOnePortal(systemInfo, client); err != nil {
			return err
		}
	}

	if DeviceName != "" {
		didSomething = true
		if err := deleteOneDevice(systemInfo, client); err != nil {
			return err
		}
	}

	if !didSomething {
		fmt.Printf("Nothing to delete -- you must specify something to delete (ie, -service=<svc_name>)\n")
	}

	return nil
}

func deleteOneService(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting service %s\n", ServiceName)
	return deleteService(systemInfo.Key, ServiceName, client)
}

func deleteOneLibrary(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting library %s\n", LibraryName)
	return deleteLibrary(systemInfo.Key, LibraryName, client)
}

func deleteOneCollection(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting collection %s\n", CollectionId)
	return deleteCollection(systemInfo.Key, CollectionId, client)
}

func deleteOneUser(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting user %s\n", UserId)
	return deleteUser(systemInfo.Key, UserId, client)
}

func deleteOneRole(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting user %s\n", RoleName)
	return deleteRole(systemInfo.Key, RoleName, client)
}

func deleteOneTrigger(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting trigger %s\n", TriggerName)
	return deleteTrigger(systemInfo.Key, TriggerName, client)
}

func deleteOneTimer(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting timer %s\n", TimerName)
	return deleteTimer(systemInfo.Key, TimerName, client)
}

func deleteOneEdge(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting edge %s\n", EdgeName)
	return deleteEdge(systemInfo.Key, EdgeName, client)
}

func deleteOnePortal(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting portal %s\n", PortalName)
	return deletePortal(systemInfo.Key, PortalName, client)
}

func deleteOneDevice(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Deleting device %s\n", DeviceName)
	return deleteDevice(systemInfo.Key, DeviceName, client)
}

func deleteService(systemKey string, name string, client *cb.DevClient) error {
	err := client.DeleteService(systemKey, name)
	if err != nil {
		return fmt.Errorf("Unable to delete service %s : %s", name, err)
	}
	return nil
}

func deleteLibrary(systemKey string, name string, client *cb.DevClient) error {
	err := client.DeleteLibrary(systemKey, name)
	if err != nil {
		return fmt.Errorf("Unable to delete library %s : %s", name, err)
	}
	return nil
}

func deleteCollection(systemKey string, colId string, client *cb.DevClient) error {
	err := client.DeleteCollection(colId)
	if err != nil {
		return fmt.Errorf("Unable to delete collection with Id %s : %s", colId, err)
	}
	return nil
}

func deleteUser(systemKey string, userId string, client *cb.DevClient) error {
	err := client.DeleteUser(systemKey, userId)
	if err != nil {
		return fmt.Errorf("Unable to delete user with Id %s : %s", userId, err)
	}
	return nil
}

func deleteRole(systemKey string, roleId string, client *cb.DevClient) error {
	err := client.DeleteRole(systemKey, roleId)
	if err != nil {
		return fmt.Errorf("Unable to delete role with Id %s : %s", roleId, err)
	}
	return nil
}

func deleteTrigger(systemKey string, name string, client *cb.DevClient) error {
	err := client.DeleteTrigger(systemKey, name)
	if err != nil {
		return fmt.Errorf("Unable to delete trigger %s : %s", name, err)
	}
	return nil
}

func deleteTimer(systemKey string, name string, client *cb.DevClient) error {
	err := client.DeleteTimer(systemKey, name)
	if err != nil {
		return fmt.Errorf("Unable to delete timer %s : %s", name, err)
	}
	return nil
}

func deleteEdge(systemKey string, name string, client *cb.DevClient) error {
	err := client.DeleteEdge(systemKey, name)
	if err != nil {
		return fmt.Errorf("Unable to delete edge %s : %s", name, err)
	}
	return nil
}

func deletePortal(systemKey string, name string, client *cb.DevClient) error {
	err := client.DeletePortal(systemKey, name)
	if err != nil {
		return fmt.Errorf("Unable to delete portal %s : %s", name, err)
	}
	return nil
}

func deleteDevice(systemKey string, name string, client *cb.DevClient) error {
	err := client.DeleteDevice(systemKey, name)
	if err != nil {
		return fmt.Errorf("Unable to delete device %s : %s", name, err)
	}
	return nil
}
