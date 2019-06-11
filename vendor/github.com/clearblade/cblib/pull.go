package cblib

import (
	"fmt"

	cb "github.com/clearblade/Go-SDK"
	"github.com/clearblade/cblib/models"
)

var (
	PULL_ALL_USERS = "%PULL_ALL_USERS%"
)

func init() {
	usage :=
		`
	Pull a ClearBlade asset from the Platform to your local filesystem. Use -sort-collections for easier version controlling of datasets.

	Note: Collection rows are pulled by default.
	`

	example :=
		`
	cb-cli pull -all												# Pull all assets from Platform to local filesystem
	cb-cli pull -all-services -all-portals							# Pull all services and all portals from Platform to local filesystem
	cb-cli pull -service=Service1 									# Pulls Service1 from Platform to local filesystem
	cb-cli pull -collection=Collection1								# Pulls Collection1 from Platform to local filesystem, with all rows, unsorted
	cb-cli pull -collection=Collection1 -sort-collections=true		# Pulls Collection1 from Platform to local filesystem, with all rows, sorted
	`
	pullCommand := &SubCommand{
		name:         "pull",
		usage:        usage,
		needsAuth:    true,
		mustBeInRepo: true,
		run:          doPull,
		example:      example,
	}

	pullCommand.flags.BoolVar(&AllServices, "all-services", false, "pull all services from system")
	pullCommand.flags.BoolVar(&AllLibraries, "all-libraries", false, "pull all libraries from system")
	pullCommand.flags.BoolVar(&AllEdges, "all-edges", false, "pull all edges from system")
	pullCommand.flags.BoolVar(&AllDevices, "all-devices", false, "pull all devices from system")
	pullCommand.flags.BoolVar(&AllPortals, "all-portals", false, "pull all portals from system")
	pullCommand.flags.BoolVar(&AllPlugins, "all-plugins", false, "pull all plugins from system")
	pullCommand.flags.BoolVar(&AllAdaptors, "all-adapters", false, "pull all adapters from system")
	pullCommand.flags.BoolVar(&AllDeployments, "all-deployments", false, "pull all deployments from system")
	pullCommand.flags.BoolVar(&AllCollections, "all-collections", false, "pull all collections from system")
	pullCommand.flags.BoolVar(&AllRoles, "all-roles", false, "pull all roles from system")
	pullCommand.flags.BoolVar(&AllUsers, "all-users", false, "pull all users from system")
	pullCommand.flags.BoolVar(&UserSchema, "userschema", false, "pull user table schema")
	pullCommand.flags.BoolVar(&AllAssets, "all", false, "pull all assets from system")
	pullCommand.flags.BoolVar(&AllTriggers, "all-triggers", false, "pull all triggers from system")
	pullCommand.flags.BoolVar(&AllTimers, "all-timers", false, "pull all timers from system")

	pullCommand.flags.StringVar(&CollectionSchema, "collectionschema", "", "Name of collection schema to pull")
	pullCommand.flags.StringVar(&ServiceName, "service", "", "Name of service to pull")
	pullCommand.flags.StringVar(&LibraryName, "library", "", "Name of library to pull")
	pullCommand.flags.StringVar(&CollectionName, "collection", "", "Name of collection to pull")
	pullCommand.flags.BoolVar(&SortCollections, "sort-collections", SortCollectionsDefault, "Sort collections by item id, for version control ease")
	pullCommand.flags.IntVar(&DataPageSize, "page-size", DataPageSizeDefault, "Number of rows in a collection to request at a time")
	pullCommand.flags.StringVar(&User, "user", "", "Name of user to pull")
	pullCommand.flags.StringVar(&RoleName, "role", "", "Name of role to pull")
	pullCommand.flags.StringVar(&TriggerName, "trigger", "", "Name of trigger to pull")
	pullCommand.flags.StringVar(&TimerName, "timer", "", "Name of timer to pull")
	pullCommand.flags.StringVar(&EdgeName, "edge", "", "Name of edge to pull")
	pullCommand.flags.StringVar(&DeviceName, "device", "", "Name of device to pull")
	pullCommand.flags.StringVar(&PortalName, "portal", "", "Name of portal to pull")
	pullCommand.flags.StringVar(&PluginName, "plugin", "", "Name of plugin to pull")
	pullCommand.flags.StringVar(&AdaptorName, "adapter", "", "Name of adapter to pull")
	pullCommand.flags.StringVar(&DeploymentName, "deployment", "", "Name of deployment to pull")

	pullCommand.flags.IntVar(&MaxRetries, "max-retries", 3, "Number of retries to attempt if a request fails")

	AddCommand("pull", pullCommand)
}

func doPull(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	SetRootDir(".")
	systemInfo, err := getSysMeta()
	if err != nil {
		return err
	}

	if err := setupDirectoryStructure(); err != nil {
		return err
	}

	// This is a hack to check if token has expired and auth again
	// since we dont have an endpoint to determine this
	client, err = checkIfTokenHasExpired(client, systemInfo.Key)
	if err != nil {
		return fmt.Errorf("Re-auth failed: %s\n", err)
	}

	assetsToPull := createAffectedAssets()
	assetsToPull.ExportItemId = true
	assetsToPull.ExportRows = true
	assetsToPull.ExportUsers = true
	didSomething, err := pullAssets(systemInfo, client, assetsToPull)

	if !didSomething {
		fmt.Printf("Nothing to pull -- you must specify something to pull (ie, -service=<svc_name>)\n")
	}
	return nil
}

func pullUserSchemaInfo(systemKey string, cli *cb.DevClient, writeThem bool) (map[string]interface{}, error) {
	resp, err := cli.GetUserColumns(systemKey)
	if err != nil {
		return nil, err
	}
	columns := []map[string]interface{}{}
	for _, colIF := range resp {
		col := colIF.(map[string]interface{})
		if col["ColumnName"] == "email" || col["ColumnName"] == "creation_date" {
			continue
		}
		columns = append(columns, col)
	}
	schema := map[string]interface{}{
		"columns": columns,
	}
	if writeThem {
		if err := writeUserSchema(schema); err != nil {
			return nil, err
		}
	}
	return schema, nil
}

func pullRole(systemKey string, roleName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetRole(systemKey, roleName)
}

func PullAndWriteRoles(systemKey string, cli *cb.DevClient, writeThem bool) ([]map[string]interface{}, error) {
	r, err := cli.GetAllRoles(systemKey)
	if err != nil {
		return nil, err
	}
	rval := make([]map[string]interface{}, 0)
	for _, rIF := range r {
		thisRole := rIF.(map[string]interface{})
		fmt.Printf(" %s", thisRole["Name"].(string))
		rval = append(rval, thisRole)
		if writeThem {
			if err := writeRole(thisRole["Name"].(string), thisRole); err != nil {
				return nil, err
			}
		}
	}
	return rval, nil
}

func PullAndWriteService(systemKey string, serviceName string, client *cb.DevClient) error {
	if svc, err := pullService(systemKey, serviceName, client); err != nil {
		return err
	} else {
		return writeService(serviceName, getRunUserEmail(svc), svc)
	}
}

func pullService(systemKey string, serviceName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetServiceRaw(systemKey, serviceName)
}

func PullAndWriteLibrary(systemKey string, libraryName string, client *cb.DevClient) error {
	if svc, err := pullLibrary(systemKey, libraryName, client); err != nil {
		return err
	} else {
		return writeLibrary(libraryName, svc)
	}
}

func pullAllUsers(systemKey string, client *cb.DevClient) ([]interface{}, error) {
	return paginateRequests(systemKey, DataPageSize, client.GetUserCountWithQuery, client.GetUsersWithQuery)
}

func PullAndWriteUsers(systemKey string, userName string, client *cb.DevClient, saveThem bool) ([]map[string]interface{}, error) {
	if users, err := pullAllUsers(systemKey, client); err != nil {
		return nil, err
	} else {
		ok := false
		rtn := make([]map[string]interface{}, 0)
		for _, u := range users {
			user := u.(map[string]interface{})
			if user["email"] == userName || userName == PULL_ALL_USERS {
				email := user["email"].(string)
				fmt.Printf(" %s", email)
				ok = true
				userId := user["user_id"].(string)
				roles, err := client.GetUserRoles(systemKey, userId)
				if err != nil {
					return nil, fmt.Errorf("Could not get roles for %s: %s", userId, err.Error())
				}
				rtn = append(rtn, user)
				if saveThem {
					err := writeUser(email, user)
					if err != nil {
						return nil, err
					}
					err = writeUserRoles(email, roles)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		if !ok {
			if userName == PULL_ALL_USERS {
				return nil, fmt.Errorf("No users found")
			} else {
				return nil, fmt.Errorf("User %+s not found\n", userName)
			}

		} else {
			return rtn, nil
		}

	}
	return nil, nil
}

func PullAndWriteCollection(systemInfo *System_meta, collectionName string, client *cb.DevClient, shouldExportRows, shouldExportItemId bool) error {
	if allColls, err := client.GetAllCollections(systemInfo.Key); err != nil {
		return err
	} else {
		var collID string
		// iterate over allColls and find one with matching name
		for _, c := range allColls {
			coll := c.(map[string]interface{})
			if collectionName == coll["name"] {
				collID = coll["collectionID"].(string)
			}
		}
		if len(collID) < 1 {
			return fmt.Errorf("Collection %s not found.", collectionName)
		}
		if coll, err := client.GetCollectionInfo(collID); err != nil {
			return err
		} else {
			if data, err := PullCollection(systemInfo, client, coll, shouldExportRows, shouldExportItemId); err != nil {
				return err
			} else {
				d := makeCollectionJsonConsistent(data)
				err = writeCollection(d["name"].(string), d)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func pullLibrary(systemKey string, libraryName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetLibrary(systemKey, libraryName)
}

func stripTriggerFields(trig map[string]interface{}) {
	delete(trig, "system_key")
	delete(trig, "system_secret")
	return
}

func writeTriggerWithUserInfo(name string, trig map[string]interface{}) error {
	users, err := getUserEmailToId()
	if err != nil {
		logWarning(fmt.Sprintf("Unable to fetch user email map when writing trigger. This can be ignored if your system doesn't have users or doesn't have any user triggers; Any user triggers in the system will be stored with userId rather than email which will affect their portability between systems. Any user triggers will need to be recreated after importing into a new system. Message: %s", err.Error()))
	} else {
		replaceUserIdWithEmailInTriggerKeyValuePairs(trig, users)
	}
	return writeTrigger(name, trig)
}

func PullAndWriteTrigger(systemKey, trigName string, client *cb.DevClient) error {
	if trigg, err := pullTrigger(systemKey, trigName, client); err != nil {
		return err
	} else {
		stripTriggerFields(trigg)
		err = writeTriggerWithUserInfo(trigName, trigg)
		if err != nil {
			return err
		}
	}
	return nil
}

func PullAndWriteTriggers(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	trigs, err := cli.GetEventHandlers(sysMeta.Key)
	if err != nil {
		return nil, fmt.Errorf("Could not pull triggers out of system %s: %s", sysMeta.Key, err.Error())
	}
	triggers := []map[string]interface{}{}
	for _, trig := range trigs {
		thisTrig := trig.(map[string]interface{})
		fmt.Printf(" %s", thisTrig["name"].(string))
		stripTriggerFields(thisTrig)
		triggers = append(triggers, thisTrig)
		err = writeTriggerWithUserInfo(thisTrig["name"].(string), thisTrig)
		if err != nil {
			return nil, err
		}
	}
	return triggers, nil
}

func PullAndWriteTimer(systemKey, timerName string, client *cb.DevClient) error {
	if timer, err := pullTimer(systemKey, timerName, client); err != nil {
		return err
	} else {
		err = writeTimer(timerName, timer)
		if err != nil {
			return err
		}
	}
	return nil
}

func PullAndWriteTimers(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	theTimers, err := cli.GetTimers(sysMeta.Key)
	if err != nil {
		return nil, fmt.Errorf("Could not pull timers out of system %s: %s", sysMeta.Key, err.Error())
	}
	timers := []map[string]interface{}{}
	for _, timer := range theTimers {
		thisTimer := timer.(map[string]interface{})
		fmt.Printf(" %s", thisTimer["name"].(string))
		timers = append(timers, thisTimer)
		err = writeTimer(thisTimer["name"].(string), thisTimer)
		if err != nil {
			return nil, err
		}
	}
	return timers, nil
}

func PullAndWritePortal(systemKey, name string, client *cb.DevClient) error {
	if portal, err := pullPortal(systemKey, name, client); err != nil {
		return err
	} else {
		return writePortal(name, portal)
	}
}

func PullAndWritePlugin(systemKey, name string, client *cb.DevClient) error {
	if plugin, err := pullPlugin(systemKey, name, client); err != nil {
		return err
	} else {
		if err = writePlugin(name, plugin); err != nil {
			return err
		}
	}
	return nil
}

func PullAndWriteAdaptor(systemKey, name string, client *cb.DevClient) error {
	if adaptor, err := pullAdaptor(systemKey, name, client); err != nil {
		return err
	} else {
		if err = writeAdaptor(adaptor); err != nil {
			return err
		}
	}
	return nil
}

func pullTrigger(systemKey string, triggerName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetEventHandler(systemKey, triggerName)
}

func pullTimer(systemKey string, timerName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetTimer(systemKey, timerName)
}

func pullDevice(systemKey string, deviceName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetDevice(systemKey, deviceName)
}

func pullEdge(systemKey string, edgeName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetEdge(systemKey, edgeName)
}

func transformPortal(portal map[string]interface{}) error {
	portal = removeBlacklistedPortalKeys(portal)
	if parsed, err := parseIfNeeded(portal["config"]); err != nil {
		return err
	} else {
		portal["config"] = parsed
	}
	return nil
}

func pullPortal(systemKey string, portalName string, client *cb.DevClient) (map[string]interface{}, error) {
	portal, err := client.GetPortal(systemKey, portalName)
	if err != nil {
		return nil, err
	}
	if err := transformPortal(portal); err != nil {
		return nil, err
	}
	return portal, nil
}

func pullPlugin(systemKey string, pluginName string, client *cb.DevClient) (map[string]interface{}, error) {
	return client.GetPlugin(systemKey, pluginName)
}

func pullAdaptor(systemKey, adaptorName string, client *cb.DevClient) (*models.Adaptor, error) {
	fmt.Printf("\n %s", adaptorName)
	currentAdaptor := models.InitializeAdaptor(adaptorName, systemKey, client)

	if err := currentAdaptor.FetchAllInfo(); err != nil {
		return nil, err
	}

	return currentAdaptor, nil
}

func updateMapNameToIDFiles(systemInfo *System_meta, client *cb.DevClient) {
	logInfo("Updating roles...")
	if data, err := PullAndWriteRoles(systemInfo.Key, client, false); err != nil {
		logError(fmt.Sprintf("Failed to update %s. %s", getRoleNameToIdFullFilePath(), err.Error()))
	} else {
		for i := 0; i < len(data); i++ {
			updateRoleNameToId(RoleInfo{
				ID:   data[i]["ID"].(string),
				Name: data[i]["Name"].(string),
			})
		}
	}
	logInfo("\nUpdating collections...")
	if data, err := PullAndWriteCollections(systemInfo, client, false, false, false); err != nil {
		logError(fmt.Sprintf("Failed to update %s. %s", getCollectionNameToIdFullFilePath(), err.Error()))
	} else {
		for i := 0; i < len(data); i++ {
			updateCollectionNameToId(CollectionInfo{
				ID:   data[i]["collection_id"].(string),
				Name: data[i]["name"].(string),
			})
		}
	}
	logInfo("Updating users...")
	if data, err := PullAndWriteUsers(systemInfo.Key, PULL_ALL_USERS, client, false); err != nil {
		logError(fmt.Sprintf("Failed to update %s. %s", getUserEmailToIdFullFilePath(), err.Error()))
	} else {
		for i := 0; i < len(data); i++ {
			updateUserEmailToId(UserInfo{
				Email:  data[i]["email"].(string),
				UserID: data[i]["user_id"].(string),
			})
		}
	}
}
