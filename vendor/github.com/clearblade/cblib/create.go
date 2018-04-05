package cblib

import (
	"fmt"
	cb "github.com/clearblade/Go-SDK"
)

func init() {
	usage := 
	`
	Creates a new asset locally
	`

	example := 
	`
	  cb-cli create -service=MyFancyNewService  # Creates a new code service: ./code/services/MyFancyNewServices/
	  cb-cli create -collection=FreshCollection # Creates a new code library: ./code/libraries/FreshCollection/
	`
	createCommand := &SubCommand{
		name:         "create",
		usage:        usage,
		needsAuth:    true,
		mustBeInRepo: true,
		run:          doCreate,
		example:	  example,
	}
	createCommand.flags.StringVar(&ServiceName, "service", "", "Name of service to create")
	createCommand.flags.StringVar(&LibraryName, "library", "", "Name of library to create")
	createCommand.flags.StringVar(&CollectionName, "collection", "", "Name of collection to create")
	createCommand.flags.StringVar(&User, "user", "", "Name of user to create")
	createCommand.flags.StringVar(&RoleName, "role", "", "Name of role to create")
	createCommand.flags.StringVar(&TriggerName, "trigger", "", "Name of trigger to create")
	createCommand.flags.StringVar(&TimerName, "timer", "", "Name of timer to create")
	AddCommand("create", createCommand)
}

func checkCreateArgsAndFlags(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("There are no arguments to the create command, only command line options\n")
	}
	return nil
}

func doCreate(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if err := checkCreateArgsAndFlags(args); err != nil {
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
		if err := createOneService(systemInfo, client); err != nil {
			return err
		}
	}

	if LibraryName != "" {
		didSomething = true
		if err := createOneLibrary(systemInfo, client); err != nil {
			return err
		}
	}

	if CollectionName != "" {
		didSomething = true
		if err := createOneCollection(systemInfo, client); err != nil {
			return err
		}
	}

	if User != "" {
		didSomething = true
		if err := createOneUser(systemInfo, client); err != nil {
			return err
		}
	}

	if RoleName != "" {
		didSomething = true
		if err := createOneRole(systemInfo, client); err != nil {
			return err
		}
	}

	if TriggerName != "" {
		didSomething = true
		if err := createOneTrigger(systemInfo, client); err != nil {
			return err
		}
	}

	if TimerName != "" {
		didSomething = true
		if err := createOneTimer(systemInfo, client); err != nil {
			return err
		}
	}

	if !didSomething {
		fmt.Printf("Nothing to update -- you must specify something to update (ie, -service=<svc_name>)\n")
	}

	return nil
}

func createOneService(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Creating service %s\n", ServiceName)
	service, err := getService(ServiceName)
	if err != nil {
		return err
	}
	return createService(systemInfo.Key, service, client)
}

func createOneCollection(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Creating collection %s\n", CollectionName)
	collection, err := getCollection(CollectionName)
	if err != nil {
		return err
	}
	return CreateCollection(systemInfo.Key, collection, client)
}

func createOneLibrary(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Creating library %s\n", LibraryName)
	library, err := getLibrary(LibraryName)
	if err != nil {
		return err
	}
	return createLibrary(systemInfo.Key, library, client)
}

func createOneUser(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Creating user %s\n", User)
	user, err := getUser(User)
	if err != nil {
		return err
	}
	_, err = createUser(systemInfo.Key, systemInfo.Secret, user, client)
	return err
}

func createOneRole(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Creating role %s\n", RoleName)
	role, err := getRole(RoleName)
	if err != nil {
		return err
	}
	return createRole(systemInfo.Key, role, client)
}

func createOneTrigger(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Creating trigger %s\n", TriggerName)
	trigger, err := getTrigger(TriggerName)
	if err != nil {
		return err
	}
	_, err = createTrigger(systemInfo.Key, trigger, client)
	return err
}

func createOneTimer(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Creating timer %s\n", TimerName)
	timer, err := getTimer(TimerName)
	if err != nil {
		return err
	}
	_, err = createTimer(systemInfo.Key, timer, client)
	return err
}
