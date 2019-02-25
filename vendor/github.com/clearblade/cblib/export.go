package cblib

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	cb "github.com/clearblade/Go-SDK"
)

var (
	inARepo bool
)

func init() {
	usage := `
	Export a System from a Platform to your local filesystem. By default, all assets are exported, except for Collection rows and Users.

	1) Exporting for first time - Run from any directory, will create a folder with same name as your system
	2) Exporting into an existing folder - 'cd' into the system's directory, and run 'cb-cli export' to export into that existing folder
	`

	example := `
	  cb-cli export                             # export default assets, omits db rows and users, Note: may prompt for remaining flags
	  cb-cli export -exportrows -exportusers    # export default asset, additionally rows and users, Note: may prompt for remaining flags
	  cb-cli export -url=https://platform.clearblade.com -messaging-url=platform.clearblade.com -system-key=9b9eea9c0bda8896a3dab5aeec9601 -email=MyDevEmail@dev.com   # Prompts for just password
	`

	systemDotJSON = map[string]interface{}{}
	svcCode = map[string]interface{}{}
	myExportCommand := &SubCommand{
		name:         "export",
		usage:        usage,
		needsAuth:    false,
		mustBeInRepo: false,
		run:          doExport,
		example:      example,
	}
	myExportCommand.flags.StringVar(&URL, "url", "https://platform.clearblade.com", "Clearblade Platform URL where system is hosted")
	myExportCommand.flags.StringVar(&MsgURL, "messaging-url", "platform.clearblade.com", "Clearblade messaging url for target system")
	myExportCommand.flags.StringVar(&SystemKey, "system-key", "", "System Key for target system, ex 9b9eea9c0bda8896a3dab5aeec9601")
	myExportCommand.flags.StringVar(&Email, "email", "", "Developer Email for login")
	myExportCommand.flags.StringVar(&DevToken, "dev-token", "", "Advanced: Developer Token for login")
	myExportCommand.flags.BoolVar(&CleanUp, "cleanup", false, "Clean up directories before export, recommended after having deleted assets on Platform")
	myExportCommand.flags.BoolVar(&ExportRows, "exportrows", false, "Exports all rows from all collections, Note: Large collections may take a long time")
	myExportCommand.flags.BoolVar(&ExportUsers, "exportusers", false, "exports user, Note: Passwords are not exported")
	myExportCommand.flags.BoolVar(&ExportItemId, "exportitemid", ExportItemIdDefault, "exports a collection rows' item_id column, Default: true")
	myExportCommand.flags.BoolVar(&SortCollections, "sort-collections", SortCollectionsDefault, "Sort collections version control ease, Note: exportitemid must be enabled")
	myExportCommand.flags.IntVar(&DataPageSize, "data-page-size", DataPageSizeDefault, "Number of rows in a collection to fetch at a time, Note: Large collections should increase up to 1000 rows")
	AddCommand("export", myExportCommand)
}

func pullRoles(systemKey string, cli *cb.DevClient, writeThem bool) ([]map[string]interface{}, error) {
	r, err := cli.GetAllRoles(systemKey)
	if err != nil {
		return nil, err
	}
	rval := make([]map[string]interface{}, len(r))
	for idx, rIF := range r {
		thisRole := rIF.(map[string]interface{})
		rval[idx] = thisRole
		if writeThem {
			if err := writeRole(thisRole["Name"].(string), thisRole); err != nil {
				return nil, err
			}
		}
	}
	return rval, nil
}

func makeCollectionNameToIdMap(collections []map[string]interface{}) map[string]interface{} {
	rtn := make(map[string]interface{})
	for i := 0; i < len(collections); i++ {
		rtn[collections[i]["name"].(string)] = collections[i]["collection_id"]
	}
	return rtn
}

func pullCollections(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	colls, err := cli.GetAllCollections(sysMeta.Key)
	if err != nil {
		return nil, err
	}
	rval := make([]map[string]interface{}, 0)
	for _, col := range colls {
		// Checking if collection is CB collection or different
		// Exporting only CB collections
		_, ok := col.(map[string]interface{})["dbtype"]
		if ok {
			continue
		}
		if r, err := PullCollection(sysMeta, col.(map[string]interface{}), cli); err != nil {
			return nil, err
		} else {
			data := makeCollectionJsonConsistent(r)
			writeCollection(r["name"].(string), data)
			rval = append(rval, data)
		}
	}
	collectionNameToId := makeCollectionNameToIdMap(rval)
	err = writeCollectionNameToId(collectionNameToId)
	if err != nil {
		fmt.Printf("Error - Failed to write collection name to ID map; subsequent operations may fail. %+v\n", err.Error())
	}

	return rval, nil
}

func PullCollection(sysMeta *System_meta, co map[string]interface{}, cli *cb.DevClient) (map[string]interface{}, error) {
	id := co["collectionID"].(string)
	isConnect := isConnectCollection(co)
	var columnsResp []interface{}
	var err error
	if isConnect {
		columnsResp = []interface{}{}
	} else {
		columnsResp, err = cli.GetColumns(id, sysMeta.Key, sysMeta.Secret)
		if err != nil {
			return nil, err
		}
	}

	//remove the item_id column if it is not supposed to be exported
	if !ExportItemId {
		//Loop through the array of maps and find the one where ColumnName = item_id
		//Remove it from the slice
		for ndx, columnMap := range columnsResp {
			if columnMap.(map[string]interface{})["ColumnName"] == "item_id" {
				columnsResp = append(columnsResp[:ndx], columnsResp[ndx+1:]...)
				break
			}
		}
	}

	co["schema"] = columnsResp
	co["items"] = []interface{}{}
	if !isConnect && ExportRows {
		items, err := pullCollectionData(co, cli)
		if err != nil {
			return nil, err
		}
		co["items"] = items
	}
	return co, nil
}

func isConnectCollection(co map[string]interface{}) bool {
	if isConnect, ok := co["isConnect"]; ok {
		switch isConnect.(type) {
		case bool:
			return isConnect.(bool)
		case string:
			return isConnect.(string) == "true"
		default:
			return false
		}
	}
	return false
}

func pullCollectionAndInfo(sysMeta *System_meta, id string, cli *cb.DevClient) (map[string]interface{}, error) {

	colInfo, err := cli.GetCollectionInfo(id)
	if err != nil {
		return nil, err
	}
	return PullCollection(sysMeta, colInfo, cli)
}

func pullCollectionData(collection map[string]interface{}, client *cb.DevClient) ([]interface{}, error) {
	colId := collection["collectionID"].(string)
	totalItems, err := client.GetItemCount(colId)
	if err != nil {
		return nil, fmt.Errorf("GetItemCount Failed: %s", err.Error())
	}

	dataQuery := &cb.Query{}
	dataQuery.PageSize = DataPageSize

	//We have to add an orderby clause in order to ensure paging works. Without the orderby clause
	//The order returned for each page is not consistent and could therefore result in duplicate rows
	//
	//https://www.postgresql.org/docs/current/static/sql-select.html
	dataQuery.Order = []cb.Ordering{cb.Ordering{OrderKey: "item_id", SortOrder: true}} // SortOrder: true means we are sorting item_id ascending
	allData := []interface{}{}
	itemIDs := make(map[string]interface{})
	totalDownloaded := 0

	if totalItems/DataPageSize > 1000 {
		fmt.Println("Large dataset detected. Recommend increasing page size. use flag: -data-page-size=1000 or -data-page-size=10000")
	}

	for j := 0; j < totalItems; j += DataPageSize {
		dataQuery.PageNumber = (j / DataPageSize) + 1

		data, err := client.GetData(colId, dataQuery)
		if err != nil {
			return nil, err
		}
		curData := data["DATA"].([]interface{})

		//Loop through the array of maps and store the value of the item_id column in
		//a map so that we can prevent adding duplicate rows
		//
		//Duplicate rows can occur when dealing with very large tables if rows are added
		//to the table while we are attempting to read pages of data. There currently is
		//no solution to remedy this.
		for _, rowMap := range curData {
			itemID := (rowMap.(map[string]interface{})["item_id"]).(string)

			if _, ok := itemIDs[itemID]; !ok {
				itemIDs[itemID] = ""

				//remove the item_id data if it is not supposed to be exported
				if !ExportItemId {
					delete(rowMap.(map[string]interface{}), "item_id")
				}
				allData = append(allData, rowMap)
				totalDownloaded++
			}
		}
		fmt.Printf("Downloaded: \tPage(s): %v / %v \tItem(s): %v / %v\n", dataQuery.PageNumber, (totalItems/DataPageSize)+1, totalDownloaded, totalItems)
	}
	return allData, nil
}

func PullServices(systemKey string, cli *cb.DevClient) ([]map[string]interface{}, error) {
	svcs, err := cli.GetServiceNames(systemKey)
	if err != nil {
		return nil, err
	}
	services := make([]map[string]interface{}, len(svcs))
	for i, svc := range svcs {
		fmt.Printf(" %s", svc)
		if s, err := pullService(systemKey, svc, cli); err != nil {
			return nil, err
		} else {
			services[i] = s
			err = writeService(s["name"].(string), s)
			if err != nil {
				return nil, err
			}
		}
	}
	return services, nil
}

func PullLibraries(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	libs, err := cli.GetLibraries(sysMeta.Key)
	if err != nil {
		return nil, fmt.Errorf("Could not pull libraries out of system %s: %s", sysMeta.Key, err.Error())
	}
	libraries := []map[string]interface{}{}
	for _, lib := range libs {
		thisLib := lib.(map[string]interface{})
		if thisLib["visibility"] == "global" {
			continue
		}
		fmt.Printf(" %s", thisLib["name"].(string))
		libraries = append(libraries, thisLib)
		err = writeLibrary(thisLib["name"].(string), thisLib)
		if err != nil {
			return nil, err
		}
	}
	return libraries, nil
}

func pullTriggers(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	trigs, err := cli.GetEventHandlers(sysMeta.Key)
	if err != nil {
		return nil, fmt.Errorf("Could not pull triggers out of system %s: %s", sysMeta.Key, err.Error())
	}
	triggers := []map[string]interface{}{}
	for _, trig := range trigs {
		thisTrig := trig.(map[string]interface{})
		delete(thisTrig, "system_key")
		delete(thisTrig, "system_secret")
		triggers = append(triggers, thisTrig)
		err = writeTrigger(thisTrig["name"].(string), thisTrig)
		if err != nil {
			return nil, err
		}
	}
	return triggers, nil
}

func pullTimers(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	theTimers, err := cli.GetTimers(sysMeta.Key)
	if err != nil {
		return nil, fmt.Errorf("Could not pull timers out of system %s: %s", sysMeta.Key, err.Error())
	}
	timers := []map[string]interface{}{}
	for _, timer := range theTimers {
		thisTimer := timer.(map[string]interface{})
		// lotsa system and user dependent stuff to get rid of...
		delete(thisTimer, "system_key")
		delete(thisTimer, "system_secret")
		delete(thisTimer, "timer_key")
		delete(thisTimer, "user_id")
		delete(thisTimer, "user_token")
		timers = append(timers, thisTimer)
		err = writeTimer(thisTimer["name"].(string), thisTimer)
		if err != nil {
			return nil, err
		}
	}
	return timers, nil
}

func pullDeployments(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	theDeployments, err := cli.GetAllDeployments(sysMeta.Key)
	if err != nil {
		return nil, fmt.Errorf("Could not pull deployments out of system %s: %s", sysMeta.Key, err)
	}
	deployments := []map[string]interface{}{}
	for _, deploymentIF := range theDeployments {

		deploymentSummary := deploymentIF.(map[string]interface{})
		deplName := deploymentSummary["name"].(string)
		deploymentDetails, err := cli.GetDeploymentByName(sysMeta.Key, deplName)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, deploymentDetails)
		if err = writeDeployment(deploymentDetails["name"].(string), deploymentDetails); err != nil {
			return nil, err
		}
	}
	return deployments, nil
}

func pullSystemMeta(systemKey string, cli *cb.DevClient) (*System_meta, error) {
	sys, err := cli.GetSystem(systemKey)
	if err != nil {
		return nil, err
	}
	serv_metas := make(map[string]Service_meta)
	sysMeta := &System_meta{
		Name:        sys.Name,
		Key:         sys.Key,
		Secret:      sys.Secret,
		Description: sys.Description,
		Services:    serv_metas,
		PlatformUrl: URL,
	}
	return sysMeta, nil
}

func getUserTablePermissions(rolesInfo []map[string]interface{}) map[string]interface{} {
	rval := map[string]interface{}{}
	for _, roleInfo := range rolesInfo {
		roleName := roleInfo["Name"].(string)
		roleUsers := roleInfo["Permissions"].(map[string]interface{})["UsersList"].(map[string]interface{})
		level := int(roleUsers["Level"].(float64))
		if level != 0 {
			rval[roleName] = level
		}
	}
	return rval
}

func storeMeta(meta *System_meta) {
	systemDotJSON["platform_url"] = cb.CB_ADDR
	systemDotJSON["messaging_url"] = cb.CB_MSG_ADDR
	systemDotJSON["system_key"] = meta.Key
	systemDotJSON["system_secret"] = meta.Secret
	systemDotJSON["name"] = meta.Name
	systemDotJSON["description"] = meta.Description
	systemDotJSON["auth"] = true
}

func pullUsers(sysMeta *System_meta, cli *cb.DevClient, saveThem bool) ([]map[string]interface{}, error) {
	sysKey := sysMeta.Key
	if !ExportUsers {
		return []map[string]interface{}{}, nil
	}
	allUsers, err := cli.GetAllUsers(sysKey)
	if err != nil {
		return nil, fmt.Errorf("Could not get all users: %s", err.Error())
	}
	for _, aUser := range allUsers {
		userId := aUser["user_id"].(string)
		roles, err := cli.GetUserRoles(sysKey, userId)
		if err != nil {
			return nil, fmt.Errorf("Could not get roles for %s: %s", userId, err.Error())
		}
		aUser["roles"] = roles
		if saveThem {
			writeUser(aUser["email"].(string), aUser)
		}
	}
	return allUsers, nil
}

func PullEdges(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := sysMeta.Key
	allEdges, err := cli.GetEdges(sysKey)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, len(allEdges))
	for i := 0; i < len(allEdges); i++ {
		currentEdge := allEdges[i].(map[string]interface{})
		fmt.Printf(" %s", currentEdge["name"].(string))
		delete(currentEdge, "edge_key")
		delete(currentEdge, "isConnected")
		delete(currentEdge, "novi_system_key")
		delete(currentEdge, "broker_auth_port")
		delete(currentEdge, "broker_port")
		delete(currentEdge, "broker_tls_port")
		delete(currentEdge, "broker_ws_auth_port")
		delete(currentEdge, "broker_ws_port")
		delete(currentEdge, "broker_wss_port")
		delete(currentEdge, "communication_style")
		delete(currentEdge, "first_talked")
		delete(currentEdge, "last_talked")
		delete(currentEdge, "local_addr")
		delete(currentEdge, "local_port")
		delete(currentEdge, "public_addr")
		delete(currentEdge, "public_port")
		err = writeEdge(currentEdge["name"].(string), currentEdge)
		if err != nil {
			return nil, err
		}
		list = append(list, currentEdge)
	}

	return list, nil
}

func pullEdgesSchema(systemKey string, cli *cb.DevClient, writeThem bool) (map[string]interface{}, error) {
	resp, err := cli.GetEdgeColumns(systemKey)
	if err != nil {
		return nil, err
	}
	columns := []map[string]interface{}{}
	sort.Strings(DefaultEdgeColumns)
	for _, colIF := range resp {
		col := colIF.(map[string]interface{})
		switch strings.ToLower(col["ColumnName"].(string)) {
		case "edge_key", "novi_system_key", "system_key", "system_secret", "token", "name", "description", "location", "mac_address", "public_addr", "public_port", "local_addr", "local_port", "broker_port", "broker_tls_port", "broker_ws_port", "broker_wss_port", "broker_auth_port", "broker_ws_auth_port", "first_talked", "last_talked", "communication_style", "last_seen_version", "policy_name", "resolver_func", "sync_edge_tables":
			continue
		default:
			columns = append(columns, col)
		}
	}
	schema := map[string]interface{}{
		"columns": columns,
	}
	if writeThem {
		if err := writeEdge("schema", schema); err != nil {
			return nil, err
		}
	}
	return schema, nil
}

func pullDevicesSchema(systemKey string, cli *cb.DevClient, writeThem bool) (map[string]interface{}, error) {
	deviceCustomColumns, err := cli.GetDeviceColumns(systemKey)
	if err != nil {
		return nil, err
	}
	columns := []map[string]interface{}{}
	sort.Strings(DefaultDeviceColumns)
	for _, colIF := range deviceCustomColumns {
		col := colIF.(map[string]interface{})
		switch strings.ToLower(col["ColumnName"].(string)) {
		case "device_key", "name", "system_key", "type", "state", "description", "enabled", "allow_key_auth", "active_key", "keys", "allow_certificate_auth", "certificate", "created_date", "last_active_date", "salt":
			continue
		default:
			columns = append(columns, col)
		}
	}
	schema := map[string]interface{}{
		"columns": columns,
	}
	if writeThem {
		if err := writeDevice("schema", schema); err != nil {
			return nil, err
		}
	}
	return schema, nil
}

func PullDevices(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := sysMeta.Key
	allDevices, err := cli.GetDevices(sysKey, nil)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, len(allDevices))
	for i := 0; i < len(allDevices); i++ {
		currentDevice := allDevices[i].(map[string]interface{})
		fmt.Printf(" %s", currentDevice["name"].(string))
		delete(currentDevice, "device_key")
		delete(currentDevice, "system_key")
		delete(currentDevice, "last_active_date")
		delete(currentDevice, "__HostId__")
		delete(currentDevice, "created_date")
		delete(currentDevice, "last_active_date")
		err = writeDevice(currentDevice["name"].(string), currentDevice)
		if err != nil {
			return nil, err
		}
		list = append(list, currentDevice)
	}
	return list, nil
}

func pullEdgeDeployInfo(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := sysMeta.Key
	deployList, err := cli.GetDeployResourcesForSystem(sysKey)
	if err != nil {
		return nil, err
	}
	return deployList, nil
}

func PullPortals(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := sysMeta.Key
	allPortals, err := cli.GetPortals(sysKey)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, len(allPortals))
	for i := 0; i < len(allPortals); i++ {
		currentPortal := allPortals[i].(map[string]interface{})
		var err error
		if err := transformPortal(currentPortal); err != nil {
			return nil, err
		}
		fmt.Printf(" %s", currentPortal["name"].(string))
		err = writePortal(currentPortal["name"].(string), currentPortal)
		if err != nil {
			return nil, err
		}
		list = append(list, currentPortal)
	}
	return list, nil
}

func PullPlugins(sysMeta *System_meta, cli *cb.DevClient) ([]map[string]interface{}, error) {
	sysKey := sysMeta.Key
	allPlugins, err := cli.GetPlugins(sysKey)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, len(allPlugins))
	for i := 0; i < len(allPlugins); i++ {
		currentPlugin := allPlugins[i].(map[string]interface{})
		fmt.Printf(" %s", currentPlugin["name"].(string))
		if err = writePlugin(currentPlugin["name"].(string), currentPlugin); err != nil {
			return nil, err
		}
		list = append(list, currentPlugin)
	}

	return list, nil
}

func PullAdaptors(sysMeta *System_meta, cli *cb.DevClient) error {
	sysKey := sysMeta.Key
	allAdaptors, err := cli.GetAdaptors(sysKey)
	if err != nil {
		return err
	}
	for i := 0; i < len(allAdaptors); i++ {
		currentAdaptorName := allAdaptors[i].(map[string]interface{})["name"].(string)
		currentAdaptor, err := pullAdaptor(sysKey, currentAdaptorName, cli)
		if err != nil {
			return err
		}

		if err = writeAdaptor(currentAdaptor); err != nil {
			return err
		}
	}

	return nil
}

func doExport(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if len(args) != 0 {
		return fmt.Errorf("export command takes no arguments; only options\n")
	}
	inARepo = MetaInfo != nil
	if inARepo {
		if exportOptionsExist() {
			return fmt.Errorf("When in a repo, you cannot have command line options")
		}
		/*
			if err := os.Chdir(".."); err != nil {
				return fmt.Errorf("Could not change to parent directory: %s", err.Error())
			}
		*/
		setupFromRepo()
	}
	var err error
	//if exportOptionsExist() {
	if DevToken != "" {
		client = cb.NewDevClientWithToken(DevToken, Email)
	} else {
		client, err = Authorize(nil)
		if err != nil {
			return fmt.Errorf("Authorize FAILED: %s\n", err)
		}
	}

	// This is a hack to check if token has expired and auth again
	// since we dont have an endpoint to determine this
	client, err = checkIfTokenHasExpired(client, SystemKey)
	if err != nil {
		return fmt.Errorf("Re-auth failed: %s", err)
	}
	return ExportSystem(client, SystemKey)
}

func exportOptionsExist() bool {
	return URL != "" || SystemKey != "" || Email != "" || DevToken != ""
}

func ExportSystem(cli *cb.DevClient, sysKey string) error {
	fmt.Printf("Exporting System Info...")
	var sysMeta *System_meta
	var err error
	if inARepo {
		sysMeta, err = getSysMeta()
		os.Chdir("..")
	} else {
		sysMeta, err = pullSystemMeta(sysKey, cli)
	}
	if err != nil {
		return err
	}
	// This was overwriting the rootdir set by cb_console
	// Only set if it has not already been set
	if !RootDirIsSet {
		SetRootDir(strings.Replace(sysMeta.Name, " ", "_", -1))
	}

	if CleanUp {
		cleanUpDirectories(sysMeta)
	}

	if err := setupDirectoryStructure(sysMeta); err != nil {
		return err
	}
	storeMeta(sysMeta)
	fmt.Printf(" Done.\nExporting Roles...")

	_, err = pullRoles(sysKey, cli, true)
	if err != nil {
		return err
	}

	fmt.Printf(" Done.\nExporting Services...")
	services, err := PullServices(sysKey, cli)
	if err != nil {
		return err
	}
	systemDotJSON["services"] = services

	fmt.Printf(" Done.\nExporting Libraries...")
	libraries, err := PullLibraries(sysMeta, cli)
	if err != nil {
		return err
	}
	systemDotJSON["libraries"] = libraries

	fmt.Printf(" Done.\nExporting Triggers...")
	if triggers, err := pullTriggers(sysMeta, cli); err != nil {
		return err
	} else {
		systemDotJSON["triggers"] = triggers
	}

	fmt.Printf(" Done.\nExporting Timers...")
	if timers, err := pullTimers(sysMeta, cli); err != nil {
		return err
	} else {
		systemDotJSON["timers"] = timers
	}

	fmt.Printf(" Done.\nExporting Collections...")
	colls, err := pullCollections(sysMeta, cli)
	if err != nil {
		return err
	}
	systemDotJSON["data"] = colls

	fmt.Printf(" Done.\nExporting Users...")
	_, err = pullUsers(sysMeta, cli, true)
	if err != nil {
		return fmt.Errorf("GetAllUsers FAILED: %s", err.Error())
	}

	fmt.Printf(" Done.\nExporting User table schema...")
	_, err = pullUserSchemaInfo(sysKey, cli, true)
	if err != nil {
		return err
	}

	fmt.Printf(" Done.\nExporting Edges...")
	edges, err := PullEdges(sysMeta, cli)
	if err != nil {
		return err
	}
	if _, err := pullEdgesSchema(sysKey, cli, true); err != nil {
		fmt.Printf("\nNo custom columns to pull and create schema.json from... Continuing...\n")
	}
	systemDotJSON["edges"] = edges
	fmt.Printf(" Done.\nExporting Devices...")
	if _, err := pullDevicesSchema(sysKey, cli, true); err != nil {
		fmt.Printf("\nNo custom columns to pull and create schema.json from... Continuing...\n")
	}
	devices, err := PullDevices(sysMeta, cli)
	if err != nil {
		return err
	}
	systemDotJSON["devices"] = devices

	fmt.Printf(" Done.\nExporting Portals...")
	portals, err := PullPortals(sysMeta, cli)
	if err != nil {
		return err
	}
	systemDotJSON["portals"] = portals

	fmt.Printf(" Done.\nExporting Plugins...")
	plugins, err := PullPlugins(sysMeta, cli)
	if err != nil {
		return err
	}
	systemDotJSON["plugins"] = plugins

	fmt.Printf(" Done.\nExporting Adaptors...")
	err = PullAdaptors(sysMeta, cli)
	if err != nil {
		return err
	}

	fmt.Printf(" Done.\nExporting Deployments...")
	if deployments, err := pullDeployments(sysMeta, cli); err != nil {
		fmt.Printf("Warning: Could not pull deployments. Might not yet be implemented on your version of the platform: %s ", err)
		//return err
	} else {
		systemDotJSON["deployments"] = deployments
	}

	fmt.Printf(" Done.\n")

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

	fmt.Printf("System '%s' has been exported into directory %s\n", sysMeta.Name, strings.Replace(sysMeta.Name, " ", "_", -1))
	return nil
}

func setupFromRepo() {
	var ok bool
	sysMeta, err := getSysMeta()
	if err != nil {
		fmt.Printf("Error getting sys meta: %s\n", err.Error())
		curDir, _ := os.Getwd()
		fmt.Printf("Current directory is %s\n", curDir)
	}
	Email, ok = MetaInfo["developerEmail"].(string)
	if !ok {
		Email = MetaInfo["developer_email"].(string)
	}
	URL, ok = MetaInfo["platformURL"].(string)
	if !ok {
		URL = MetaInfo["platform_url"].(string)
	}
	DevToken = MetaInfo["token"].(string)
	SystemKey = sysMeta.Key
}

func parseIfNeeded(stuff interface{}) (map[string]interface{}, error) {
	switch stuff.(type) {
	case map[string]interface{}:
		return stuff.(map[string]interface{}), nil
	case string:
		parsed := map[string]interface{}{}
		if err := json.Unmarshal([]byte(stuff.(string)), &parsed); err != nil {
			return nil, err
		}
		return parsed, nil
	default:
		return nil, fmt.Errorf("Invalid type passed into parseIfNeeded. Must be string or map[string]interface{}")
	}
}
