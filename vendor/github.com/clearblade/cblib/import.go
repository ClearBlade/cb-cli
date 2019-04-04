package cblib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cb "github.com/clearblade/Go-SDK"
)

var (
	importRows  bool
	importUsers bool
)

func init() {

	usage :=
		`
	Import a system from your local filesystem to the ClearBlade Platform
	`

	example :=
		`
	cb-cli import 									# prompts for credentials
	cb-cli import -importrows=false -importusers=false			# prompts for credentials, excludes all collection-rows and users
	`
	myImportCommand := &SubCommand{
		name:         "import",
		usage:        usage,
		needsAuth:    false,
		mustBeInRepo: true,
		run:          doImport,
		example:      example,
	}
	DEFAULT_IMPORT_ROWS := true
	DEFAULT_IMPORT_USERS := true
	myImportCommand.flags.BoolVar(&importRows, "importrows", DEFAULT_IMPORT_ROWS, "imports all data into all collections")
	myImportCommand.flags.BoolVar(&importUsers, "importusers", DEFAULT_IMPORT_USERS, "imports all users into the system")
	myImportCommand.flags.StringVar(&URL, "url", "https://platform.clearblade.com", "Clearblade Platform URL where system is hosted, ex https://platform.clearblade.com")
	myImportCommand.flags.StringVar(&Email, "email", "", "Developer email for login to import destination")
	myImportCommand.flags.StringVar(&Password, "password", "", "Developer password at import destination")
	myImportCommand.flags.IntVar(&DataPageSize, "data-page-size", DataPageSizeDefault, "Number of rows in a collection to push/import at a time")
	AddCommand("import", myImportCommand)
	AddCommand("imp", myImportCommand)
	AddCommand("im", myImportCommand)
}

func createSystem(system map[string]interface{}, client *cb.DevClient) (map[string]interface{}, error) {
	name := system["name"].(string)
	desc := system["description"].(string)
	auth := system["auth"].(bool)
	sysKey, sysErr := client.NewSystem(name, desc, auth)
	if sysErr != nil {
		return nil, sysErr
	}
	realSystem, sysErr := client.GetSystem(sysKey)
	if sysErr != nil {
		return nil, sysErr
	}
	system["systemKey"] = realSystem.Key
	system["systemSecret"] = realSystem.Secret
	return system, nil
}

func createRoles(systemInfo map[string]interface{}, collectionsInfo []CollectionInfo, client *cb.DevClient) error {
	sysKey := systemInfo["systemKey"].(string)
	roles, err := getRoles()
	if err != nil {
		return err
	}
	for _, role := range roles {
		name := role["Name"].(string)
		fmt.Printf(" %s", name)
		//if name != "Authenticated" && name != "Administrator" && name != "Anonymous" {
		if err := createRole(sysKey, role, collectionsInfo, client); err != nil {
			return err
		}
		//}
	}
	fmt.Println("\nUpdating local roles with newly created role IDs... ")
	// ids were created on import for the new roles, grab those
	_, err = PullAndWriteRoles(sysKey, client, true)
	if err != nil {
		return err
	}

	return nil
}

func createUsers(systemInfo map[string]interface{}, users []map[string]interface{}, client *cb.DevClient) error {
	BLACKLISTED_USER_COLUMN := "user_id"
	//  Create user columns first -- if any
	sysKey := systemInfo["systemKey"].(string)
	sysSec := systemInfo["systemSecret"].(string)
	userCols := []interface{}{}
	userSchema, err := getUserSchema()
	if err == nil {
		userCols = userSchema["columns"].([]interface{})
	}
	for _, columnIF := range userCols {
		column := columnIF.(map[string]interface{})
		columnName := column["ColumnName"].(string)
		if columnName == BLACKLISTED_USER_COLUMN {
			continue
		}
		columnType := column["ColumnType"].(string)
		if err := client.CreateUserColumn(sysKey, columnName, columnType); err != nil {
			return fmt.Errorf("Could not create user column %s: %s", columnName, err.Error())
		}
	}

	if !importUsers {
		return nil
	}

	// Now, create users -- register, update roles, and update user-def colunms
	for _, user := range users {
		fmt.Printf(" %s", user["email"].(string))
		userId, err := createUser(sysKey, sysSec, user, client)
		if err != nil {
			// don't return an error because we don't want to stop other users from being created
			fmt.Printf("Error: Failed to create user %s - %s", user["email"].(string), err.Error())
		}
		if err := updateUserEmailToId(UserInfo{
			UserID: userId,
			Email:  user["email"].(string),
		}); err != nil {
			logErrorForUpdatingMapFile(getUserEmailToIdFullFilePath(), err)
		}

		if len(userCols) == 0 {
			continue
		}

		updates := map[string]interface{}{}
		for _, columnIF := range userCols {
			column := columnIF.(map[string]interface{})
			columnName := column["ColumnName"].(string)
			if columnName != "user_id" {
				if userVal, ok := user[columnName]; ok {
					if userVal != nil {
						updates[columnName] = userVal
					}
				}
			}
		}

		if len(updates) == 0 {
			continue
		}

		if err := client.UpdateUser(sysKey, userId, updates); err != nil {
			// don't return an error because we don't want to stop other users from being updated
			fmt.Printf("Could not update user: %s", err.Error())
		}
	}

	return nil
}

func unMungeRoles(roles []string) []interface{} {
	rval := []interface{}{}

	for _, role := range roles {
		rval = append(rval, role)
	}
	return rval
}

func createTriggers(systemInfo map[string]interface{}, client *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := systemInfo["systemKey"].(string)
	triggers, err := getTriggers()
	if err != nil {
		return nil, err
	}
	triggersRval := make([]map[string]interface{}, len(triggers))
	for idx, trigger := range triggers {
		fmt.Printf(" %s", trigger["name"].(string))
		trigVal, err := createTrigger(sysKey, trigger, client)
		if err != nil {
			return nil, err
		}
		triggersRval[idx] = trigVal
	}
	return triggersRval, nil
}

func createTimers(systemInfo map[string]interface{}, client *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := systemInfo["systemKey"].(string)
	timers, err := getTimers()
	if err != nil {
		return nil, err
	}
	timersRval := make([]map[string]interface{}, len(timers))
	for idx, timer := range timers {
		fmt.Printf(" %s", timer["name"].(string))
		timerVal, err := createTimer(sysKey, timer, client)
		if err != nil {
			return nil, err
		}
		timersRval[idx] = timerVal
	}
	return timersRval, nil
}

func createDeployments(systemInfo map[string]interface{}, client *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := systemInfo["systemKey"].(string)
	deployments, err := getDeployments()
	if err != nil {
		return nil, err
	}
	deploymentsRval := make([]map[string]interface{}, len(deployments))
	for idx, deployment := range deployments {
		fmt.Printf(" %s", deployment["name"].(string))
		deploymentVal, err := createDeployment(sysKey, deployment, client)
		if err != nil {
			return nil, err
		}
		deploymentsRval[idx] = deploymentVal
	}
	return deploymentsRval, nil
}

func createServices(systemInfo map[string]interface{}, client *cb.DevClient) error {
	sysKey := systemInfo["systemKey"].(string)
	services, err := getServices()
	if err != nil {
		fmt.Printf("getServices Failed: %s\n", err)
		return err
	}
	for _, service := range services {
		fmt.Printf(" %s", service["name"].(string))
		if err := createService(sysKey, service, client); err != nil {
			fmt.Printf("createService Failed: %s\n", err)
			return err
		}
	}
	return nil
}

func createLibraries(systemInfo map[string]interface{}, client *cb.DevClient) error {
	sysKey := systemInfo["systemKey"].(string)
	libraries, err := getLibraries()
	if err != nil {
		fmt.Printf("getLibraries Failed: %s\n", err)
		return err
	}
	for _, library := range libraries {
		fmt.Printf(" %s", library["name"].(string))
		if err := createLibrary(sysKey, library, client); err != nil {
			fmt.Printf("createLibrary Failed: %s\n", err)
			return err
		}
	}
	return nil
}

func createAdaptors(systemInfo map[string]interface{}, client *cb.DevClient) error {
	sysKey := systemInfo["systemKey"].(string)
	adaptors, err := getAdaptors(sysKey, client)
	if err != nil {
		return err
	}
	for i := 0; i < len(adaptors); i++ {
		err := createAdaptor(adaptors[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func createCollections(systemInfo map[string]interface{}, client *cb.DevClient) ([]CollectionInfo, error) {
	sysKey := systemInfo["systemKey"].(string)
	collections, err := getCollections()
	rtn := make([]CollectionInfo, 0)
	if err != nil {
		return rtn, err
	}

	for _, collection := range collections {
		fmt.Printf(" %s\n", collection["name"].(string))
		if info, err := CreateCollection(sysKey, collection, client); err != nil {
			return rtn, err
		} else {
			rtn = append(rtn, info)
		}
	}
	return rtn, nil
}

// Reads Filesystem and makes HTTP calls to platform to create edges and edge columns
// Note: Edge schemas are optional, so if it is not found, we log an error and continue
func createEdges(systemInfo map[string]interface{}, client *cb.DevClient) error {
	sysKey := systemInfo["systemKey"].(string)
	sysSecret := systemInfo["systemSecret"].(string)
	edgesSchema, err := getEdgesSchema()
	if err != nil {
		// To ensure backwards-compatibility, we do not require
		// this folder `edges` to be present
		// As a result, let's log this error, but proceed
		fmt.Printf("Warning, could not find optional edge schema -- ignoring\n")
		return nil
	}

	edgesCols, ok := edgesSchema["columns"].([]interface{})
	if ok {
		for _, columnIF := range edgesCols {
			column := columnIF.(map[string]interface{})
			columnName := column["ColumnName"].(string)
			columnType := column["ColumnType"].(string)
			if err := client.CreateEdgeColumn(sysKey, columnName, columnType); err != nil {
				return fmt.Errorf("Could not create edges column %s: %s", columnName, err.Error())
			}
		}
	}

	edges, err := getEdges()
	if err != nil {
		return err
	}
	for _, edge := range edges {
		fmt.Printf(" %s", edge["name"].(string))
		edgeName := edge["name"].(string)
		delete(edge, "name")
		edge["system_key"] = sysKey
		edge["system_secret"] = sysSecret
		if err := createEdge(sysKey, edgeName, edge, client); err != nil {
			return err
		}
	}
	return nil
}

func createDevices(systemInfo map[string]interface{}, client *cb.DevClient) ([]map[string]interface{}, error) {
	schemaPresent := true
	sysKey := systemInfo["systemKey"].(string)
	devicesSchema, err := getDevicesSchema()
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			schemaPresent = false
		} else {
			return nil, err
		}
	}
	if schemaPresent {
		deviceCols, ok := devicesSchema["columns"].([]interface{})
		if ok {
			for _, columnIF := range deviceCols {
				column := columnIF.(map[string]interface{})
				columnName := column["ColumnName"].(string)
				if columnName == "salt" {
					fmt.Printf("Warning: ignoring exported 'salt' column\n")
					continue
				}
				columnType := column["ColumnType"].(string)
				if err := client.CreateDeviceColumn(sysKey, columnName, columnType); err != nil {
					fmt.Printf("Failed Creating device column %s\n", columnName)
					return nil, fmt.Errorf("Could not create devices column %s: %s", columnName, err.Error())
				}
				fmt.Printf("Created device column %s\n", columnName)
			}
		} else {
			return nil, fmt.Errorf("columns key not present in schema.json for devices")
		}
	}
	devices, err := getDevices()
	if err != nil {
		return nil, err
	}
	devicesRval := make([]map[string]interface{}, len(devices))
	for idx, device := range devices {
		if !schemaPresent {
			if idx == 0 {
				for columnname, _ := range device {
					switch strings.ToLower(columnname) {
					case "device_key", "name", "system_key", "type", "state", "description", "enabled", "allow_key_auth", "active_key", "keys", "allow_certificate_auth", "certificate", "created_date", "last_active_date":
						continue
					default:
						err := client.CreateDeviceColumn(sysKey, columnname, "string")
						if err != nil {
							return nil, err
						}
					}
				}
			}
		}
		deviceName := device["name"].(string)
		fmt.Printf(" %s", deviceName)
		deviceInfo, err := createDevice(sysKey, device, client)
		if err != nil {
			return nil, err
		}
		deviceRoles, err := getDeviceRoles(deviceName)
		if err != nil {
			// system is probably in the legacy format, let's just set the roles to the default
			deviceRoles = convertStringSliceToInterfaceSlice([]string{"Authenticated"})
			logWarning(fmt.Sprintf("Could not find roles for device with name '%s'. This device will be created with only the default 'Authenticated' role.", deviceName))
		}
		defaultRoles := convertStringSliceToInterfaceSlice([]string{"Authenticated"})
		roleDiff := diffRoles(deviceRoles, defaultRoles)
		if err := client.UpdateDeviceRoles(sysKey, deviceName, convertInterfaceSliceToStringSlice(roleDiff.add), convertInterfaceSliceToStringSlice(roleDiff.remove)); err != nil {
			return nil, err
		}
		devicesRval[idx] = deviceInfo
	}
	return devicesRval, nil
}

func createPortals(systemInfo map[string]interface{}, client *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := systemInfo["systemKey"].(string)
	var portals []map[string]interface{}
	var err error
	if hasLegacyPortalDirectory() {
		portals, err = getLegacyPortals()
		if err != nil {
			return nil, err
		}
	} else {
		portals, err = getCompressedPortals()
		if err != nil {
			return nil, err
		}
	}
	portalsRval := make([]map[string]interface{}, len(portals))
	for idx, dash := range portals {
		fmt.Printf(" %s", dash["name"].(string))
		portalInfo, err := createPortal(sysKey, dash, client)
		if err != nil {
			return nil, err
		}
		portalsRval[idx] = portalInfo
	}
	return portalsRval, nil
}

func createAllEdgeDeployment(systemInfo map[string]interface{}, client *cb.DevClient) error {
	//  First, look for deploy.json file. This is the new way of doing edge
	//  deployment. If that fails try the old way.
	if fileExists(rootDir + "/deploy.json") {
		info, err := getEdgeDeployInfo()
		if err != nil {
			return err
		}
		return createEdgeDeployInfo(systemInfo, info, client)
	}
	return oldCreateEdgeDeployInfo(systemInfo, client) // old deprecated way
}

func createEdgeDeployInfo(systemInfo, deployInfo map[string]interface{}, client *cb.DevClient) error {
	deployList := deployInfo["deployInfo"].([]interface{})
	sysKey := systemInfo["systemKey"].(string)

	for _, deployOneIF := range deployList {
		deployOne, ok := deployOneIF.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Poorly structured edge deploy info")
		}
		platform := deployOne["platform"].(bool)
		resName := deployOne["resource_identifier"].(string)
		resType := deployOne["resource_type"].(string)

		//  Go sdk expects the edge query to be in the Query format, not a string
		edgeQueryStr := deployOne["edge"].(string)
		_, err := client.CreateDeployResourcesForSystem(sysKey, resName, resType, platform, edgeQueryStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func oldCreateEdgeDeployInfo(systemInfo map[string]interface{}, client *cb.DevClient) error {
	sysKey := systemInfo["systemKey"].(string)
	edgeInfo, ok := systemInfo["edgeSync"].(map[string]interface{})
	if !ok {
		fmt.Printf("Warning: Could not find any edge sync information\n")
		return nil
	}
	for edgeName, edgeSyncInfoIF := range edgeInfo {
		edgeSyncInfo, ok := edgeSyncInfoIF.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Poorly formed edge sync info")
		}
		converted, err := convertOldEdgeDeployInfo(edgeSyncInfo)
		if err != nil {
			return err
		}
		_, err = client.SyncResourceToEdge(sysKey, edgeName, converted, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func createPlugins(systemInfo map[string]interface{}, client *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := systemInfo["systemKey"].(string)
	plugins, err := getPlugins()
	if err != nil {
		return nil, err
	}
	pluginsRval := make([]map[string]interface{}, len(plugins))
	for idx, plug := range plugins {
		fmt.Printf(" %s", plug["name"].(string))
		pluginVal, err := createPlugin(sysKey, plug, client)
		if err != nil {
			return nil, err
		}
		pluginsRval[idx] = pluginVal
	}
	return pluginsRval, nil
}

func convertOldEdgeDeployInfo(info map[string]interface{}) (map[string][]string, error) {
	rval := map[string][]string{
		"service": []string{},
		"library": []string{},
		"trigger": []string{},
	}
	for resourceKey, _ := range info {
		stuff := strings.Split(resourceKey, "::")
		if len(stuff) != 2 {
			return nil, fmt.Errorf("Poorly formed edge sync info entry: '%s'", resourceKey)
		}
		rval[stuff[0]] = append(rval[stuff[0]], stuff[1])
	}
	return rval, nil
}

func enableLogs(service map[string]interface{}) bool {
	logVal, ok := service["logging_enabled"]
	if !ok {
		return false
	}
	switch logVal.(type) {
	case string:
		return logVal.(string) == "true"
	case bool:
		return logVal.(bool) == true
	}
	return false
}

func mkSvcParams(params []interface{}) []string {
	rval := []string{}
	for _, val := range params {
		rval = append(rval, val.(string))
	}
	return rval
}

func doImport(cmd *SubCommand, cli *cb.DevClient, args ...string) error {
	return importIt(cli)
}

func hijackAuthorize() (*cb.DevClient, error) {
	svMetaInfo := MetaInfo
	MetaInfo = nil
	SystemKey = "DummyTemporaryPlaceholder"
	cli, err := Authorize(nil)
	if err != nil {
		return nil, err
	}
	SystemKey = ""
	MetaInfo = svMetaInfo
	return cli, nil
}

// Used in pairing with importMySystem:
func devTokenHardAuthorize() (*cb.DevClient, error) {
	// MetaInfo should not be nil, else the current process will prompt user on command line
	if MetaInfo == nil {
		return nil, errors.New("MetaInfo cannot be nil")
	}
	SystemKey = "DummyTemporaryPlaceholder"
	cli, err := Authorize(nil)
	if err != nil {
		return nil, err
	}
	SystemKey = ""
	return cli, nil
}

// TODO Handle more specific error for if folder doesnt exist
// i.e. plugins folder not found vs plugins import failed due to syntax error
// https://clearblade.atlassian.net/browse/CBCOMM-227
func importAllAssets(systemInfo map[string]interface{}, users []map[string]interface{}, cli *cb.DevClient) error {

	// Common set of calls for a complete system import

	fmt.Printf(" Done.\nImporting collections...")
	collectionsInfo, err := createCollections(systemInfo, cli)
	if err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create collections: %s", err.Error())
	}
	fmt.Printf(" Done.\nImporting roles...")
	err = createRoles(systemInfo, collectionsInfo, cli)
	if err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create roles: %s", err.Error())
	}
	fmt.Printf(" Done.\nImporting users...")
	if err := createUsers(systemInfo, users, cli); err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create users: %s", err.Error())
	}
	fmt.Printf(" Done.\nImporting code services...")
	// Additonal modifications to the ImportIt functions
	if err := createServices(systemInfo, cli); err != nil {
		serr, _ := err.(*os.PathError)
		if err != serr {
			return err
		} else {
			fmt.Printf("Path Error importing services: Operation: %s Path %s, Error %s\n", serr.Op, serr.Path, serr.Err)
			fmt.Printf("Warning: Could not import code services... -- ignoring\n")
		}
	}
	fmt.Printf(" Done.\nImporting code libraries...")
	if err := createLibraries(systemInfo, cli); err != nil {
		serr, _ := err.(*os.PathError)
		if err != serr {
			return err
		} else {
			fmt.Printf("Path Error importing libraries: Operation: %s Path %s, Error %s\n", serr.Op, serr.Path, serr.Err)
			fmt.Printf("Warning: Could not import code libraries... -- ignoring\n")
		}
	}
	fmt.Printf(" Done.\nImporting triggers...")
	_, err = createTriggers(systemInfo, cli)
	if err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create triggers: %s", err.Error())
	}
	fmt.Printf(" Done.\nImporting timers...")
	_, err = createTimers(systemInfo, cli)
	if err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create timers: %s", err.Error())
	}

	fmt.Printf(" Done.\nImporting edges...")
	if err := createEdges(systemInfo, cli); err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create edges: %s", err.Error())
	}
	fmt.Printf(" Done.\nImporting devices...")
	_, err = createDevices(systemInfo, cli)
	if err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create devices: %s", err.Error())
	}
	fmt.Printf(" Done.\nImporting portals...")
	_, err = createPortals(systemInfo, cli)
	if err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create portals: %s", err.Error())
	}
	fmt.Printf(" Done.\nImporting plugins...")
	_, err = createPlugins(systemInfo, cli)
	if err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create plugins: %s", err.Error())
	}
	fmt.Printf(" Done. \nImporting adaptors...")
	if err := createAdaptors(systemInfo, cli); err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create adaptors: %s", err.Error())
	}
	fmt.Printf(" Done. \nImporting deployments...")
	if _, err := createDeployments(systemInfo, cli); err != nil {
		//  Don't return an err, just warn -- so we keep back compat with old systems
		fmt.Printf("Could not create deployments: %s", err.Error())
	}

	fmt.Printf(" Done\n")
	return nil
}

func importIt(cli *cb.DevClient) error {
	//fmt.Printf("Reading system configuration files...")
	SetRootDir(".")
	if err := setupDirectoryStructure(); err != nil {
		return err
	}
	users, err := getUsers()
	if err != nil {
		return err
	}

	systemInfo, err := getDict("system.json")
	if err != nil {
		return err
	}
	// The DevClient should be null at this point because we are delaying auth until
	// Now.
	cli, err = hijackAuthorize()
	if err != nil {
		return err
	}
	//fmt.Printf("Done.\nImporting system...")
	fmt.Printf("Importing system...")
	if data, err := createSystem(systemInfo, cli); err != nil {
		return fmt.Errorf("Could not create system %s: %s", systemInfo["name"], err.Error())
	} else {
		logInfo(fmt.Sprintf("Successfully created new system. System key is - %s", data["systemKey"].(string)))
	}

	return importAllAssets(systemInfo, users, cli)
}

// Import assuming the system is there in the root directory
// Alternative to ImportIt for Import from UI
// if intoExistingSystem is true then userInfo should have system key else error will be thrown

func importSystem(cli *cb.DevClient, rootdirectory string, userInfo map[string]interface{}) (map[string]interface{}, error) {

	// Point the rootDirectory to the extracted folder
	SetRootDir(rootdirectory)
	users, err := getUsers()
	if err != nil {
		return nil, err
	}
	// as we don't cd into folders we have to use full path !!
	path := filepath.Join(rootdirectory, "/system.json")

	systemInfo, err := getDict(path)
	if err != nil {
		return nil, err
	}

	// Hijack to make sure the MetaInfo is not nil
	cli, err = devTokenHardAuthorize() // Hijacking Authorize()
	if err != nil {
		return nil, err
	}
	// updating system info accordingly
	if userInfo["importIntoExistingSystem"].(bool) {
		systemInfo["systemKey"] = userInfo["system_key"]
		systemInfo["systemSecret"] = userInfo["system_secret"]
	} else {
		fmt.Printf("Importing system...")
		if userInfo["systemName"] != nil {
			systemInfo["name"] = userInfo["systemName"]
		}
		if _, err := createSystem(systemInfo, cli); err != nil {
			return nil, fmt.Errorf("Could not create system %s: %s", systemInfo["name"], err.Error())
		}
	}
	return systemInfo, importAllAssets(systemInfo, users, cli)
}

// Call this wrapper from the end point !!
func ImportSystem(cli *cb.DevClient, dir string, userInfo map[string]interface{}) (map[string]interface{}, error) {

	// Setting the MetaInfo which is used by Authorize() it has developerEmail, devToken, MsgURL, URL
	// not changing the overall metaInfo, in case its used some where else
	tempmetaInfo := MetaInfo
	MetaInfo = userInfo
	// similar to old importIt
	systemInfo, err := importSystem(cli, dir, userInfo)
	MetaInfo = tempmetaInfo
	return systemInfo, err
}
