package cblib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clearblade/cblib/models"

	cb "github.com/clearblade/Go-SDK"
)

func init() {

	usage :=
		`
	Push a ClearBlade asset from local filesystem to ClearBlade Platform
	`

	example :=
		`
	cb-cli push -all							# Push all assets up to Platform
	cb-cli push -all -auto-approve				# Push all assets up to Platform and automatically confirm any prompts for creating new assets
	cb-cli push -all-services -all-portals		# Push all services and all portals up to Platform
	cb-cli push -service=Service1				# Push a code service up to Platform
	cb-cli push -collection=Collection1			# Push a code service up to Platform
	`

	pushCommand := &SubCommand{
		name:         "push",
		usage:        usage,
		needsAuth:    true,
		mustBeInRepo: true,
		run:          doPush,
		example:      example,
	}

	pushCommand.flags.BoolVar(&UserSchema, "userschema", false, "push user table schema")
	pushCommand.flags.BoolVar(&EdgeSchema, "edgeschema", false, "push edges table schema")
	pushCommand.flags.BoolVar(&DeviceSchema, "deviceschema", false, "push devices table schema")
	pushCommand.flags.BoolVar(&AllServices, "all-services", false, "push all of the local services")
	pushCommand.flags.BoolVar(&AllLibraries, "all-libraries", false, "push all of the local libraries")
	pushCommand.flags.BoolVar(&AllDevices, "all-devices", false, "push all of the local devices")
	pushCommand.flags.BoolVar(&AllEdges, "all-edges", false, "push all of the local edges")
	pushCommand.flags.BoolVar(&AllPortals, "all-portals", false, "push all of the local portals")
	pushCommand.flags.BoolVar(&AllPlugins, "all-plugins", false, "push all of the local plugins")
	pushCommand.flags.BoolVar(&AllAdaptors, "all-adapters", false, "push all of the local adapters")
	pushCommand.flags.BoolVar(&AllCollections, "all-collections", false, "push all of the local collections")
	pushCommand.flags.BoolVar(&AllRoles, "all-roles", false, "push all of the local roles")
	pushCommand.flags.BoolVar(&AllUsers, "all-users", false, "push all of the local users")
	pushCommand.flags.BoolVar(&AllAssets, "all", false, "push all of the local assets")
	pushCommand.flags.BoolVar(&AllTriggers, "all-triggers", false, "push all of the local triggers")
	pushCommand.flags.BoolVar(&AllTimers, "all-timers", false, "push all of the local timers")
	pushCommand.flags.BoolVar(&AllDeployments, "all-deployments", false, "push all of the local deployments")
	pushCommand.flags.BoolVar(&AutoApprove, "auto-approve", false, "automatically answer yes to all prompts. Useful for creating new entities when they aren't found in the platform")

	pushCommand.flags.StringVar(&CollectionSchema, "collectionschema", "", "Name of collection schema to push")
	pushCommand.flags.StringVar(&ServiceName, "service", "", "Name of service to push")
	pushCommand.flags.StringVar(&LibraryName, "library", "", "Name of library to push")
	pushCommand.flags.StringVar(&CollectionName, "collection", "", "Name of collection to push")
	pushCommand.flags.StringVar(&CollectionId, "collectionID", "", "Unique id of collection to update. -collection flag is preferred")
	pushCommand.flags.StringVar(&User, "user", "", "Name of user to push")
	pushCommand.flags.StringVar(&UserId, "userID", "", "Unique id of user to update. -user flag is preferred")
	pushCommand.flags.StringVar(&RoleName, "role", "", "Name of role to push")
	pushCommand.flags.StringVar(&TriggerName, "trigger", "", "Name of trigger to push")
	pushCommand.flags.StringVar(&TimerName, "timer", "", "Name of timer to push")
	pushCommand.flags.StringVar(&DeviceName, "device", "", "Name of device to push")
	pushCommand.flags.StringVar(&EdgeName, "edge", "", "Name of edge to push")
	pushCommand.flags.StringVar(&PortalName, "portal", "", "Name of portal to push")
	pushCommand.flags.StringVar(&PluginName, "plugin", "", "Name of plugin to push")
	pushCommand.flags.StringVar(&AdaptorName, "adapter", "", "Name of adapter to push")
	pushCommand.flags.StringVar(&DeploymentName, "deployment", "", "Name of deployment to push")

	pushCommand.flags.IntVar(&DataPageSize, "data-page-size", DataPageSizeDefault, "Number of rows in a collection to push/import at a time")

	AddCommand("push", pushCommand)
}

func checkPushArgsAndFlags(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("There are no arguments to the push command, only command line options\n")
	}
	if AllServices && ServiceName != "" {
		return fmt.Errorf("Cannot specify both -all-services and -service=<service_name>\n")
	}
	if AllLibraries && LibraryName != "" {
		return fmt.Errorf("Cannot specify both -all-libraries and -library=<library_name>\n")
	}
	return nil
}

func pushOneService(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing service %+s\n", name)
	service, err := getService(name)
	if err != nil {
		return err
	}
	return updateServiceWithRunAs(systemInfo.Key, name, service, client)
}

func pushUserSchema(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Printf("Pushing user schema\n")
	userschema, err := getUserSchema()
	if err != nil {
		return err
	}
	userColumns, err := client.GetUserColumns(systemInfo.Key)
	if err != nil {
		return fmt.Errorf("Error fetching user columns: %s", err.Error())
	}

	localSchema, ok := userschema["columns"].([]interface{})
	if !ok {
		return fmt.Errorf("Error in schema definition. Pls check the format of schema...\n")
	}

	diff := getDiffForColumns(localSchema, userColumns, DefaultUserColumns)
	for i := 0; i < len(diff.remove); i++ {
		if err := client.DeleteUserColumn(systemInfo.Key, diff.remove[i].(map[string]interface{})["ColumnName"].(string)); err != nil {
			return fmt.Errorf("User schema could not be updated. Deletion of column(s) failed: %s", err)
		}
	}
	for i := 0; i < len(diff.add); i++ {
		if err := client.CreateUserColumn(systemInfo.Key, diff.add[i].(map[string]interface{})["ColumnName"].(string), diff.add[i].(map[string]interface{})["ColumnType"].(string)); err != nil {
			return fmt.Errorf("Failed to create user column '%s': %s", diff.add[i].(map[string]interface{})["ColumnName"].(string), err.Error())
		}
	}
	return nil
}

func getDiffForColumns(localSchemaInterfaces, backendSchemaInterfaces []interface{}, defaultColumns []string) ListDiff {
	return compareLists(localSchemaInterfaces, backendSchemaInterfaces, columnExists, func(a interface{}) bool {
		return isDefaultColumn(defaultColumns, a.(map[string]interface{})["ColumnName"].(string))
	})
}

func columnExists(colA interface{}, colB interface{}) bool {
	if colA.(map[string]interface{})["ColumnName"].(string) == colB.(map[string]interface{})["ColumnName"].(string) {
		return true
	}
	return false
}

func pushEdgesSchema(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Println("Pushing edge schema")
	edgeschema, err := getEdgesSchema()
	if err != nil {
		return err
	}
	allEdgeColumns, err := client.GetEdgeColumns(systemInfo.Key)
	if err != nil {
		return err
	}

	typedLocalSchema, ok := edgeschema["columns"].([]interface{})
	if !ok {
		return fmt.Errorf("Error in schema definition. Please verify the format of the schema.json. Value is: %+v - %+v\n", edgeschema["columns"], ok)
	}

	diff := getDiffForColumns(typedLocalSchema, allEdgeColumns, DefaultEdgeColumns)
	for i := 0; i < len(diff.remove); i++ {
		if err := client.DeleteEdgeColumn(systemInfo.Key, diff.remove[i].(map[string]interface{})["ColumnName"].(string)); err != nil {
			return fmt.Errorf("Unable to delete column '%s': %s", diff.remove[i].(map[string]interface{})["ColumnName"].(string), err.Error())
		}
	}
	for i := 0; i < len(diff.add); i++ {
		if err := client.CreateEdgeColumn(systemInfo.Key, diff.add[i].(map[string]interface{})["ColumnName"].(string), diff.add[i].(map[string]interface{})["ColumnType"].(string)); err != nil {
			return fmt.Errorf("Unable to create column '%s': %s", diff.add[i].(map[string]interface{})["ColumnName"].(string), err.Error())
		}
	}

	return nil

}

func pushDevicesSchema(systemInfo *System_meta, client *cb.DevClient) error {
	fmt.Println("Pushing device schema")
	deviceSchema, err := getDevicesSchema()
	if err != nil {
		return err
	}
	allDeviceColumns, err := client.GetDeviceColumns(systemInfo.Key)
	if err != nil {
		return err
	}
	localSchema, ok := deviceSchema["columns"].([]interface{})
	if !ok {
		return fmt.Errorf("Error in schema definition. Please verify the format of the schema.json\n")
	}

	diff := getDiffForColumns(localSchema, allDeviceColumns, DefaultDeviceColumns)
	for i := 0; i < len(diff.remove); i++ {
		if err := client.DeleteDeviceColumn(systemInfo.Key, diff.remove[i].(map[string]interface{})["ColumnName"].(string)); err != nil {
			return fmt.Errorf("Unable to delete column '%s': %s", diff.remove[i].(map[string]interface{})["ColumnName"].(string), err.Error())
		}
	}
	for i := 0; i < len(diff.add); i++ {
		if err := client.CreateDeviceColumn(systemInfo.Key, diff.add[i].(map[string]interface{})["ColumnName"].(string), diff.add[i].(map[string]interface{})["ColumnType"].(string)); err != nil {
			return fmt.Errorf("Unable to create column '%s': %s", diff.add[i].(map[string]interface{})["ColumnName"].(string), err.Error())
		}
	}

	return nil

}

func pushAllCollections(systemInfo *System_meta, client *cb.DevClient) error {
	allColls, err := getCollections()
	if err != nil {
		return err
	}
	for i := 0; i < len(allColls); i++ {
		err := pushOneCollection(systemInfo, client, allColls[i]["name"].(string))
		if err != nil {
			return err
		}
	}
	return nil
}

func pushOneCollection(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing collection %s\n", name)
	collection, err := getCollection(name)
	if err != nil {
		fmt.Printf("error is %+v\n", err)
		return err
	}
	return updateCollection(systemInfo, collection, client)
}

func pushOneCollectionById(systemInfo *System_meta, client *cb.DevClient, wantedId string) error {
	fmt.Printf("Pushing collection with collectionID %s\n", wantedId)
	collections, err := getCollections()
	if err != nil {
		return err
	}
	for _, collection := range collections {
		id, ok := collection["collectionID"].(string)
		if !ok {
			continue
		}
		if id == wantedId {
			return updateCollection(systemInfo, collection, client)
		}
	}
	return fmt.Errorf("Collection with collectionID %+s not found.", wantedId)
}

func pushUsers(systemInfo *System_meta, client *cb.DevClient) error {
	users, err := getUsers()
	if err != nil {
		return err
	}
	for i := 0; i < len(users); i++ {
		// todo: make getUser accept user object so that it doesn't refetch from the FS
		if err := pushOneUser(systemInfo, client, users[i]["email"].(string)); err != nil {
			return err
		}
	}
	return nil
}

func pushOneUser(systemInfo *System_meta, client *cb.DevClient, email string) error {
	user, err := getFullUserObject(email)
	if err != nil {
		return err
	}
	return updateUser(systemInfo, user, client)
}

func pushOneUserById(systemInfo *System_meta, client *cb.DevClient, wantedId string) error {
	fmt.Printf("Pushing user with user_id %s\n", wantedId)
	users, err := getUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		id, ok := user["user_id"].(string)
		if !ok {
			continue
		}
		if id == wantedId {
			return updateUser(systemInfo, user, client)
		}
	}
	return fmt.Errorf("User with user_id %+s not found.", wantedId)
}

func pushRoles(systemInfo *System_meta, client *cb.DevClient) error {
	allRoles, err := getRoles()
	if err != nil {
		return err
	}
	for i := 0; i < len(allRoles); i++ {
		if err := pushOneRole(systemInfo, client, allRoles[i]["Name"].(string)); err != nil {
			return err
		}
	}
	return nil
}

func pushOneRole(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing role %s\n", name)
	role, err := getRole(name)
	if err != nil {
		return err
	}
	return updateRole(systemInfo.Key, role, getCollectionNameToIdAsSliceWithErrorCheck(), client)
}

func pushTriggers(systemInfo *System_meta, client *cb.DevClient) error {
	allTriggers, err := getTriggers()
	if err != nil {
		return err
	}
	for i := 0; i < len(allTriggers); i++ {
		if err := pushOneTrigger(systemInfo, client, allTriggers[i]["name"].(string)); err != nil {
			return err
		}
	}
	return nil
}

func pushOneTrigger(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing trigger %+s\n", name)
	trigger, err := getTrigger(name)
	if err != nil {
		return err
	}
	return updateTrigger(systemInfo.Key, trigger, client)
}

func pushTimers(systemInfo *System_meta, client *cb.DevClient) error {
	allTimers, err := getTimers()
	if err != nil {
		return err
	}
	for i := 0; i < len(allTimers); i++ {
		if err := pushOneTimer(systemInfo, client, allTimers[i]["name"].(string)); err != nil {
			return err
		}
	}
	return nil
}

func pushOneTimer(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing timer %+s\n", name)
	timer, err := getTimer(name)
	if err != nil {
		return err
	}
	return updateTimer(systemInfo.Key, timer, client)
}

func pushOneDevice(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing device %+s\n", name)
	device, err := getDevice(name)
	if err != nil {
		return err
	}
	return updateDevice(systemInfo.Key, device, client)
}

func pushAllDevices(systemInfo *System_meta, client *cb.DevClient) error {
	devices, err := getDevices()
	if err != nil {
		return err
	}
	for _, device := range devices {
		fmt.Printf("Pushing device %+s\n", device["name"].(string))
		if err := updateDevice(systemInfo.Key, device, client); err != nil {
			return fmt.Errorf("Error updating device '%s': %s\n", device["name"].(string), err.Error())
		}
	}
	return nil
}

func pushOneEdge(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing edge %+s\n", name)
	edge, err := getEdge(name)
	if err != nil {
		return err
	}
	return updateEdge(systemInfo.Key, edge, client)
}

func pushAllEdges(systemInfo *System_meta, client *cb.DevClient) error {
	edges, err := getEdges()
	if err != nil {
		return err
	}
	for _, edge := range edges {
		fmt.Printf("Pushing edge %+s\n", edge["name"].(string))
		if err := updateEdge(systemInfo.Key, edge, client); err != nil {
			return fmt.Errorf("Error updating edge '%s': %s\n", edge["name"].(string), err.Error())
		}
	}
	return nil
}

func pushOnePortal(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing portal %+s\n", name)
	compressedPortal, err := compressPortal(name)
	if err != nil {
		return err
	}
	return updatePortal(systemInfo.Key, compressedPortal, client)
}

func pushAllPortals(systemInfo *System_meta, client *cb.DevClient) error {
	portals, err := getCompressedPortals()
	if err != nil {
		return err
	}
	for _, portal := range portals {
		name := portal["name"].(string)
		fmt.Printf("Pushing portal %+s\n", name)
		if err := updatePortal(systemInfo.Key, portal, client); err != nil {
			return fmt.Errorf("Error updating portal '%s': %s\n", name, err.Error())
		}
	}
	return nil
}

func pushOnePlugin(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing portal %+s\n", name)
	plugin, err := getPlugin(name)
	if err != nil {
		return err
	}
	return updatePlugin(systemInfo.Key, plugin, client)
}

func pushAllPlugins(systemInfo *System_meta, client *cb.DevClient) error {
	plugins, err := getPlugins()
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		fmt.Printf("Pushing plugin %+s\n", plugin["name"].(string))
		if err := updatePlugin(systemInfo.Key, plugin, client); err != nil {
			return fmt.Errorf("Error updating plugin '%s': %s\n", plugin["name"].(string), err.Error())
		}
	}
	return nil
}

func pushOneAdaptor(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing adaptor %+s\n", name)
	sysKey := systemInfo.Key
	adaptor, err := getAdaptor(sysKey, name, client)
	if err != nil {
		return err
	}
	return handleUpdateAdaptor(systemInfo.Key, adaptor, client)
}

func pushAllAdaptors(systemInfo *System_meta, client *cb.DevClient) error {
	sysKey := systemInfo.Key
	adaptors, err := getAdaptors(sysKey, client)
	if err != nil {
		return err
	}
	for i := 0; i < len(adaptors); i++ {
		currentAdaptor := adaptors[i]
		fmt.Printf("Pushing adaptor %+s\n", currentAdaptor.Name)
		if err := handleUpdateAdaptor(sysKey, currentAdaptor, client); err != nil {
			return fmt.Errorf("Error updating adaptor '%s': %s\n", currentAdaptor.Name, err.Error())
		}
	}
	return nil
}

func pushAllServices(systemInfo *System_meta, client *cb.DevClient) error {
	services, err := getServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		name := service["name"].(string)
		fmt.Printf("Pushing service %+s\n", name)
		if err := updateServiceWithRunAs(systemInfo.Key, name, service, client); err != nil {
			return fmt.Errorf("Error updating service '%s': %s\n", service["name"].(string), err.Error())
		}
	}
	return nil
}

func pushOneLibrary(systemInfo *System_meta, client *cb.DevClient, name string) error {
	fmt.Printf("Pushing library %+s\n", name)

	library, err := getLibrary(name)
	if err != nil {
		return err
	}
	return updateLibrary(systemInfo.Key, library, client)
}

func pushAllLibraries(systemInfo *System_meta, client *cb.DevClient) error {
	libraries, err := getLibraries()
	if err != nil {
		return err
	}
	for _, library := range libraries {
		fmt.Printf("Pushing library %+s\n", library["name"].(string))
		if err := updateLibrary(systemInfo.Key, library, client); err != nil {
			return fmt.Errorf("Error updating library '%s': %s\n", library["name"].(string), err.Error())
		}
	}
	return nil
}

func doPush(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if err := checkPushArgsAndFlags(args); err != nil {
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
		return fmt.Errorf("Re-auth failed...")
	}

	didSomething := false

	if AllServices || AllAssets {
		didSomething = true
		if err := pushAllServices(systemInfo, client); err != nil {
			return err
		}
	}

	if ServiceName != "" {
		didSomething = true
		if err := pushOneService(systemInfo, client, ServiceName); err != nil {
			return err
		}
	}

	if AllLibraries || AllAssets {
		didSomething = true
		if err := pushAllLibraries(systemInfo, client); err != nil {
			return err
		}
	}

	if LibraryName != "" {
		didSomething = true
		if err := pushOneLibrary(systemInfo, client, LibraryName); err != nil {
			return err
		}
	}

	if AllCollections || AllAssets {
		didSomething = true
		if err := pushAllCollections(systemInfo, client); err != nil {
			return err
		}
	}

	if CollectionSchema != "" {
		didSomething = true
		if err := pushCollectionSchema(systemInfo, client, CollectionSchema); err != nil {
			return err
		}
	}

	if CollectionName != "" {
		didSomething = true
		if err := pushOneCollection(systemInfo, client, CollectionName); err != nil {
			return err
		}
	}

	if CollectionId != "" {
		didSomething = true
		if err := pushOneCollectionById(systemInfo, client, CollectionId); err != nil {
			return err
		}
	}

	if UserSchema || AllAssets {
		didSomething = true
		if err := pushUserSchema(systemInfo, client); err != nil {
			return err
		}
	}

	if AllUsers || AllAssets {
		didSomething = true
		if err := pushUsers(systemInfo, client); err != nil {
			return err
		}
	}

	if User != "" {
		didSomething = true
		if err := pushOneUser(systemInfo, client, User); err != nil {
			return err
		}
	}

	if UserId != "" {
		didSomething = true
		if err := pushOneUserById(systemInfo, client, UserId); err != nil {
			return err
		}
	}

	if AllRoles || AllAssets {
		didSomething = true
		if err := pushRoles(systemInfo, client); err != nil {
			return err
		}
	}

	if RoleName != "" {
		didSomething = true
		if err := pushOneRole(systemInfo, client, RoleName); err != nil {
			return err
		}
	}

	if AllTriggers || AllAssets {
		didSomething = true
		if err := pushTriggers(systemInfo, client); err != nil {
			return err
		}
	}

	if TriggerName != "" {
		didSomething = true
		if err := pushOneTrigger(systemInfo, client, TriggerName); err != nil {
			return err
		}
	}

	if AllTimers || AllAssets {
		didSomething = true
		if err := pushTimers(systemInfo, client); err != nil {
			return err
		}
	}

	if TimerName != "" {
		didSomething = true
		if err := pushOneTimer(systemInfo, client, TimerName); err != nil {
			return err
		}
	}

	if DeviceSchema || AllAssets {
		didSomething = true
		if err := pushDevicesSchema(systemInfo, client); err != nil {
			return err
		}
	}

	if AllDevices || AllAssets {
		didSomething = true
		if err := pushAllDevices(systemInfo, client); err != nil {
			return err
		}
	}

	if DeviceName != "" {
		didSomething = true
		if err := pushOneDevice(systemInfo, client, DeviceName); err != nil {
			return err
		}
	}

	if EdgeSchema || AllAssets {
		didSomething = true
		if err := pushEdgesSchema(systemInfo, client); err != nil {
			return err
		}
	}

	if AllEdges || AllAssets {
		didSomething = true
		if err := pushAllEdges(systemInfo, client); err != nil {
			return err
		}
	}

	if EdgeName != "" {
		didSomething = true
		if err := pushOneEdge(systemInfo, client, EdgeName); err != nil {
			return err
		}
	}

	if AllPortals || AllAssets {
		didSomething = true
		if err := pushAllPortals(systemInfo, client); err != nil {
			return err
		}
	}

	if PortalName != "" {
		didSomething = true
		if err := pushOnePortal(systemInfo, client, PortalName); err != nil {
			return err
		}
	}

	if AllPlugins || AllAssets {
		didSomething = true
		if err := pushAllPlugins(systemInfo, client); err != nil {
			return err
		}
	}

	if PluginName != "" {
		didSomething = true
		if err := pushOnePlugin(systemInfo, client, PluginName); err != nil {
			return err
		}
	}

	if AllAdaptors || AllAssets {
		didSomething = true
		if err := pushAllAdaptors(systemInfo, client); err != nil {
			return err
		}
	}

	if AdaptorName != "" {
		didSomething = true
		if err := pushOneAdaptor(systemInfo, client, AdaptorName); err != nil {
			return err
		}
	}

	if AllDeployments || AllAssets {
		didSomething = true
		if err := pushDeployments(systemInfo, client); err != nil {
			return err
		}
	}

	if DeploymentName != "" {
		didSomething = true
		if err := pushDeployment(systemInfo, client, DeploymentName); err != nil {
			return err
		}
	}

	if !didSomething {
		fmt.Printf("Nothing to push -- you must specify something to push (ie, -service=<svc_name>)\n")
	}

	return nil
}

func pushCollectionSchema(systemInfo *System_meta, cli *cb.DevClient, name string) error {
	fmt.Printf("Pushing collection schema for '%s'\n", name)
	allCollectionsInfo, err := getCollectionNameToIdAsSlice()
	if err != nil {
		return err
	}
	collID, err := getCollectionIdByName(name, allCollectionsInfo)
	if err != nil {
		return err
	}
	localCollInfo, err := getCollection(name)
	if err != nil {
		return err
	}

	backendSchema, err := cli.GetColumnsByCollectionName(systemInfo.Key, name)
	if err != nil {
		return err
	}
	localSchema, ok := localCollInfo["schema"].([]interface{})
	if !ok {
		return fmt.Errorf("Error in schema definition. Please verify the format of the schema.json\n")
	}

	diff := getDiffForColumns(localSchema, backendSchema, DefaultCollectionColumns)
	for i := 0; i < len(diff.remove); i++ {
		if err := cli.DeleteColumn(collID, diff.remove[i].(map[string]interface{})["ColumnName"].(string)); err != nil {
			return fmt.Errorf("Unable to delete column '%s': %s", diff.remove[i].(map[string]interface{})["ColumnName"].(string), err.Error())
		}
	}
	for i := 0; i < len(diff.add); i++ {
		if err := cli.AddColumn(collID, diff.add[i].(map[string]interface{})["ColumnName"].(string), diff.add[i].(map[string]interface{})["ColumnType"].(string)); err != nil {
			return fmt.Errorf("Unable to create column '%s': %s", diff.add[i].(map[string]interface{})["ColumnName"].(string), err.Error())
		}
	}

	return nil
}

func pushDeployments(systemInfo *System_meta, cli *cb.DevClient) error {
	deps, err := getDeployments()
	if err != nil {
		return err
	}
	for i := 0; i < len(deps); i++ {
		err := pushDeployment(systemInfo, cli, deps[i]["name"].(string))
		if err != nil {
			return err
		}
	}
	return nil
}

func pushDeployment(systemInfo *System_meta, cli *cb.DevClient, name string) error {
	dep, err := getDeployment(name)
	if err != nil {
		return err
	}
	return updateDeployment(systemInfo, cli, name, dep)
}

func updateDeployment(systemInfo *System_meta, cli *cb.DevClient, name string, dep map[string]interface{}) error {
	// fetch deployment
	backendDep, err := cli.GetDeploymentByName(systemInfo.Key, name)
	if err != nil {
		fmt.Printf("Could not find deployment '%s'. Error is - %s\n", name, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new deployment named %s?", name))
		if err != nil {
			return err
		} else {
			if c {
				if _, err := createDeployment(systemInfo.Key, dep, cli); err != nil {
					return fmt.Errorf("Could not create deployment %s: %s", name, err.Error())
				} else {
					fmt.Printf("Successfully created new deployment %s\n", name)
				}
			} else {
				fmt.Printf("Deployment will not be created.\n")
			}
		}
	} else {

		// diff backend deployment and local deployment
		theDiff := diffDeployments(dep, backendDep)
		if _, err := cli.UpdateDeploymentByName(systemInfo.Key, name, theDiff); err != nil {
			return err
		}
	}

	return nil
}

func diffDeployments(localDep map[string]interface{}, backendDep map[string]interface{}) map[string]interface{} {
	assetDiff := compareLists(localDep["assets"].([]interface{}), backendDep["assets"].([]interface{}), isAssetMatch, func(a interface{}) bool { return false })
	edgeDiff := compareLists(localDep["edges"].([]interface{}), backendDep["edges"].([]interface{}), isEdgeMatch, func(a interface{}) bool { return false })
	return map[string]interface{}{
		"assets": map[string]interface{}{
			"add":    assetDiff.add,
			"remove": assetDiff.remove,
		},
		"edges": map[string]interface{}{
			"add":    edgeDiff.add,
			"remove": edgeDiff.remove,
		},
	}
}

func isEdgeMatch(edgeA interface{}, edgeB interface{}) bool {
	return edgeA.(string) == edgeB.(string)
}

func isAssetMatch(assetA interface{}, assetB interface{}) bool {
	typedA := assetA.(map[string]interface{})
	typedB := assetB.(map[string]interface{})
	return typedA["asset_class"].(string) == typedB["asset_class"].(string) && typedA["asset_id"].(string) == typedB["asset_id"].(string) && typedA["sync"].(bool) == typedB["sync"].(bool)
}

func createRole(systemKey string, role map[string]interface{}, collectionsInfo []CollectionInfo, client *cb.DevClient) error {
	roleName := role["Name"].(string)
	var roleID string
	if roleName != "Authenticated" && roleName != "Anonymous" && roleName != "Administrator" {
		createIF, err := client.CreateRole(systemKey, role["Name"].(string))
		if err != nil {
			return err
		}
		createDict, ok := createIF.(map[string]interface{})
		if !ok {
			return fmt.Errorf("return value from CreateRole is not a map. It is %T", createIF)
		}
		roleID, ok = createDict["role_id"].(string)
		if !ok {
			return fmt.Errorf("Did not get role_id key back from successful CreateRole call")
		}
	} else {
		roleID = roleName // Administrator, Authorized, Anonymous
	}
	updateRoleBody, err := packageRoleForUpdate(roleID, role, collectionsInfo)
	if err != nil {
		return err
	}
	if err := client.UpdateRole(systemKey, role["Name"].(string), updateRoleBody); err != nil {
		return err
	}
	if err := updateRoleNameToId(RoleInfo{ID: roleID, Name: roleName}); err != nil {
		logErrorForUpdatingMapFile(getRoleNameToIdFullFilePath(), err)
	}
	return nil
}

func packageRoleForUpdate(roleID string, role map[string]interface{}, collectionsInfo []CollectionInfo) (map[string]interface{}, error) {
	permissions, ok := role["Permissions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Permissions for role do not exist or is not a map")
	}
	convertedPermissions := convertPermissionsStructure(permissions, collectionsInfo)
	return map[string]interface{}{"ID": roleID, "Permissions": convertedPermissions}, nil
}

func getCollectionIdByName(theNameWeWant string, collectionsInfo []CollectionInfo) (string, error) {
	for i := 0; i < len(collectionsInfo); i++ {
		if collectionsInfo[i].Name == theNameWeWant {
			return collectionsInfo[i].ID, nil
		}
	}
	return "", fmt.Errorf("Couldn't find ID for collection name '%s'\n", theNameWeWant)
}

// it's possible that there are duplicate permissions
// we need to remove any duplicates so that a role create/update succeeds
func removeDuplicatePermissions(perms []map[string]interface{}, idKey string) []map[string]interface{} {
	rtn := make([]map[string]interface{}, 0)
	foundIds := make(map[string]bool)

	for i := 0; i < len(perms); i++ {
		id := perms[i]["itemInfo"].(map[string]interface{})[idKey].(string)
		if _, found := foundIds[id]; !found {
			foundIds[id] = true
			rtn = append(rtn, perms[i])
		}
	}

	return rtn
}

//
//  The roles structure we get back when we retrieve roles is different from
//  the format accepted for updating a role. Thus, we have this beauty of a
//  conversion function. -swm
//
//  THis is a gigantic cluster. We need to fix and learn from this. -swm
//
func convertPermissionsStructure(in map[string]interface{}, collectionsInfo []CollectionInfo) map[string]interface{} {
	out := map[string]interface{}{}
	for key, valIF := range in {
		switch key {
		case "CodeServices":
			if valIF != nil {
				services, err := getASliceOfMaps(valIF)
				if err != nil {
					fmt.Printf("Bad format for services permissions, not a slice of maps: %T\n", valIF)
					os.Exit(1)
				}
				svcs := make([]map[string]interface{}, len(services))
				for idx, mapVal := range services {
					svcs[idx] = map[string]interface{}{
						"itemInfo":    map[string]interface{}{"name": mapVal["Name"]},
						"permissions": mapVal["Level"],
					}
				}
				out["services"] = removeDuplicatePermissions(svcs, "name")
			}
		case "Collections":
			if valIF != nil {
				collections, err := getASliceOfMaps(valIF)
				if err != nil {
					fmt.Printf("Bad format for collections permissions, not a slice of maps: %T\n", valIF)
					os.Exit(1)
				}
				cols := make([]map[string]interface{}, 0)
				for _, mapVal := range collections {
					collName := mapVal["Name"].(string)
					id, err := getCollectionIdByName(collName, collectionsInfo)
					if err != nil {
						fmt.Printf("Skipping permissions for collection '%s'; Error is - %s", collName, err.Error())
						continue
					}
					cols = append(cols, map[string]interface{}{
						"itemInfo":    map[string]interface{}{"id": id, "name": collName},
						"permissions": mapVal["Level"],
					})
				}
				out["collections"] = removeDuplicatePermissions(cols, "id")
			}
		case "DevicesList":
			if valIF != nil {
				val := getMap(valIF)
				out["devices"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "MsgHistory":
			if valIF != nil {
				val := getMap(valIF)
				out["msgHistory"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "SystemServices":
			if valIF != nil {
				val := getMap(valIF)
				out["system_services"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "Portals":
			if valIF != nil {
				portals, err := getASliceOfMaps(valIF)
				if err != nil {
					fmt.Printf("Bad format for portals permissions, not a slice of maps: %T\n", valIF)
					os.Exit(1)
				}
				ptls := make([]map[string]interface{}, len(portals))
				for idx, mapVal := range portals {
					ptls[idx] = map[string]interface{}{
						"itemInfo":    map[string]interface{}{"name": mapVal["Name"]},
						"permissions": mapVal["Level"],
					}
				}
				out["portals"] = removeDuplicatePermissions(ptls, "name")
			}
		case "Push":
			if valIF != nil {
				val := getMap(valIF)
				out["push"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "Topics":
			if valIF != nil {
				topics, err := getASliceOfMaps(valIF)
				if err != nil {
					fmt.Printf("Bad format for topics permissions, not a slice of maps: %T\n", valIF)
					os.Exit(1)
				}
				tpcs := make([]map[string]interface{}, len(topics))
				for idx, mapVal := range topics {
					tpcs[idx] = map[string]interface{}{
						"itemInfo":    map[string]interface{}{"name": mapVal["Name"]},
						"permissions": mapVal["Level"],
					}
				}
				out["topics"] = tpcs
			}
		case "UsersList":
			if valIF != nil {
				val := getMap(valIF)
				out["users"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "EdgesList":
			if valIF != nil {
				val := getMap(valIF)
				out["edges"] = map[string]interface{}{"permissions": val["Level"]}
			}

		case "Triggers":
			if valIF != nil {
				val := getMap(valIF)
				out["triggers"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "Timers":
			if valIF != nil {
				val := getMap(valIF)
				out["timers"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "Deployments":
			if valIF != nil {
				val := getMap(valIF)
				out["deployments"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "Roles":
			if valIF != nil {
				val := getMap(valIF)
				out["roles"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "AllCollections":
			if valIF != nil {
				val := getMap(valIF)
				out["allcollections"] = map[string]interface{}{"permissions": val["Level"]}
			}
		case "AllServices":
			if valIF != nil {
				val := getMap(valIF)
				out["allservices"] = map[string]interface{}{"permissions": val["Level"]}
			}
		default:

		}
	}
	return out
}

// The main thing I hate about go: type assertions
func getASliceOfMaps(val interface{}) ([]map[string]interface{}, error) {
	switch val.(type) {
	case []map[string]interface{}:
		return val.([]map[string]interface{}), nil
	case []interface{}:
		rval := make([]map[string]interface{}, len(val.([]interface{})))
		for idx, mapVal := range val.([]interface{}) {
			switch mapVal.(type) {
			case map[string]interface{}:
				rval[idx] = mapVal.(map[string]interface{})
			default:
				return nil, fmt.Errorf("slice values are not maps: %T\n", mapVal)
			}
		}
		return rval, nil
	default:
		return nil, fmt.Errorf("Bad type %T: expecting a slice", val)
	}
}

func getMap(val interface{}) map[string]interface{} {
	switch val.(type) {
	case map[string]interface{}:
		return val.(map[string]interface{})
	default:
		fmt.Printf("permissions type must be a map, not %T\n", val)
		os.Exit(1)
	}
	return map[string]interface{}{}
}

func updateUser(meta *System_meta, user map[string]interface{}, client *cb.DevClient) error {

	if email, ok := user["email"].(string); !ok {
		return fmt.Errorf("Missing user email %+v", user)
	} else {
		_, err := client.GetUserInfo(meta.Key, email)
		if err != nil {
			fmt.Printf("Could not update user '%s'. Error is - %s\n", email, err.Error())
			c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new user with email %s?", email))
			if err != nil {
				return err
			} else {
				if c {
					id, err := createUser(meta.Key, meta.Secret, user, client)
					if err != nil {
						return fmt.Errorf("Could not create user %s: %s", email, err.Error())
					} else {
						// tack the new user id onto the user object so it can be used in subsequent requests
						user["user_id"] = id
						fmt.Printf("Successfully created new user %s\n", email)
					}
				} else {
					fmt.Printf("User will not be created.\n")
					return nil
				}
			}
		}
	}

	userRoles, err := getUserRoles(user["email"].(string))
	if err != nil {
		return err
	}

	backendUserRoles, err := client.GetUserRoles(meta.Key, user["user_id"].(string))
	if err != nil {
		return err
	}
	roleDiff := diffRoles(userRoles, convertStringSliceToInterfaceSlice(backendUserRoles))
	user["roles"] = map[string]interface{}{
		"add":    convertInterfaceSliceToStringSlice(roleDiff.add),
		"delete": convertInterfaceSliceToStringSlice(roleDiff.remove),
	}
	return client.UpdateUser(meta.Key, user["user_id"].(string), user)
}

func createUser(systemKey string, systemSecret string, user map[string]interface{}, client *cb.DevClient) (string, error) {
	email := user["email"].(string)
	password := "password"
	if pwd, ok := user["password"]; ok {
		password = pwd.(string)
	}
	newUser, err := client.RegisterNewUser(email, password, systemKey, systemSecret)
	if err != nil {
		return "", fmt.Errorf("Could not create user %s: %s", email, err.Error())
	}
	userId := newUser["user_id"].(string)
	if err := updateUserEmailToId(UserInfo{
		UserID: userId,
		Email:  email,
	}); err != nil {
		logErrorForUpdatingMapFile(getUserEmailToIdFullFilePath(), err)
	}
	userRoles, err := getUserRoles(email)
	if err != nil {
		// couldn't get user roles, let's see if they're on the user map (legacy format)
		if r, ok := user["roles"].([]interface{}); ok {
			userRoles = r
		} else {
			logWarning(fmt.Sprintf("Could not find roles for user with email '%s'. This user will be created with only the default 'Authenticated' role.", email))
			userRoles = convertStringSliceToInterfaceSlice([]string{"Authenticated"})
		}
	}
	defaultRoles := convertStringSliceToInterfaceSlice([]string{"Authenticated"})
	roleDiff := diffRoles(userRoles, defaultRoles)
	if len(roleDiff.add) > 0 || len(roleDiff.remove) > 0 {
		add := convertInterfaceSliceToStringSlice(roleDiff.add)
		remove := convertInterfaceSliceToStringSlice(roleDiff.remove)
		if err := client.UpdateUserRoles(systemKey, userId, add, remove); err != nil {
			return "", err
		}
	}
	return userId, nil
}

func diffRoles(local, backend []interface{}) ListDiff {
	return compareLists(local, backend, roleExists, func(a interface{}) bool { return false })
}

func roleExists(a interface{}, b interface{}) bool {
	return a == b
}

func createTrigger(sysKey string, trigger map[string]interface{}, client *cb.DevClient) (map[string]interface{}, error) {
	triggerName := trigger["name"].(string)
	triggerDef := trigger["event_definition"].(map[string]interface{})
	trigger["def_module"] = triggerDef["def_module"]
	trigger["def_name"] = triggerDef["def_name"]
	trigger["system_key"] = sysKey
	delete(trigger, "name")
	delete(trigger, "event_definition")
	stuff, err := client.CreateEventHandler(sysKey, triggerName, trigger)
	if err != nil {
		return nil, fmt.Errorf("Could not create trigger %s: %s", triggerName, err.Error())
	}
	return stuff, nil
}

func updateTrigger(systemKey string, trigger map[string]interface{}, client *cb.DevClient) error {
	triggerName := trigger["name"].(string)

	triggerDef := trigger["event_definition"].(map[string]interface{})
	trigger["def_module"] = triggerDef["def_module"]
	trigger["def_name"] = triggerDef["def_name"]
	trigger["system_key"] = systemKey
	delete(trigger, "event_definition")

	if _, err := pullTrigger(systemKey, triggerName, client); err != nil {
		fmt.Printf("Could not find trigger '%s'. Error is - %s\n", triggerName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new trigger named %s?", triggerName))
		if err != nil {
			return err
		} else {
			if c {
				if _, err := client.CreateEventHandler(systemKey, triggerName, trigger); err != nil {
					return fmt.Errorf("Could not create trigger %s: %s", triggerName, err.Error())
				} else {
					fmt.Printf("Successfully created new trigger %s\n", triggerName)
				}
			} else {
				fmt.Printf("Trigger will not be created.\n")
			}
		}
	} else {

		delete(trigger, "name")

		if _, err := client.UpdateEventHandler(systemKey, triggerName, trigger); err != nil {
			return err
		}
	}
	return nil
}

func createTimer(systemKey string, timer map[string]interface{}, client *cb.DevClient) (map[string]interface{}, error) {
	timerName := timer["name"].(string)
	delete(timer, "name")
	startTime := timer["start_time"].(string)
	if startTime == "Now" {
		timer["start_time"] = time.Now().Format(time.RFC3339)
	}
	if _, err := client.CreateTimer(systemKey, timerName, timer); err != nil {
		return nil, fmt.Errorf("Could not create timer %s: %s", timerName, err.Error())
	}
	return timer, nil
}

func updateTimer(systemKey string, timer map[string]interface{}, client *cb.DevClient) error {
	timerName := timer["name"].(string)
	delete(timer, "name")
	startTime := timer["start_time"].(string)
	if startTime == "Now" {
		timer["start_time"] = time.Now().Format(time.RFC3339)
	}

	if _, err := pullTimer(systemKey, timerName, client); err != nil {
		fmt.Printf("Could not find timer '%s'. Error is - %s\n", timerName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new timer named %s?", timerName))
		if err != nil {
			return err
		} else {
			if c {
				if _, err := client.CreateTimer(systemKey, timerName, timer); err != nil {
					return fmt.Errorf("Could not create timer %s: %s", timerName, err.Error())
				} else {
					fmt.Printf("Successfully created new timer %s\n", timerName)
				}
			} else {
				fmt.Printf("Timer will not be created.\n")
			}
		}
	} else {
		if _, err := client.UpdateTimer(systemKey, timerName, timer); err != nil {
			return err
		}
	}
	return nil
}

func createDeployment(systemKey string, deployment map[string]interface{}, client *cb.DevClient) (map[string]interface{}, error) {
	deploymentName := deployment["name"].(string)
	//delete(deployment, "name")
	if _, err := client.CreateDeploymentByName(systemKey, deploymentName, deployment); err != nil {
		return nil, fmt.Errorf("Could not create deployment %s: %s", deploymentName, err.Error())
	}
	return deployment, nil
}

func updateDevice(systemKey string, device map[string]interface{}, client *cb.DevClient) error {
	deviceName := device["name"].(string)
	delete(device, "last_active_date")
	delete(device, "created_date")
	delete(device, "device_key")
	delete(device, "system_key")

	if _, err := pullDevice(systemKey, deviceName, client); err != nil {
		fmt.Printf("Could not find device '%s'. Error is - %s\n", deviceName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new device named %s?", deviceName))
		if err != nil {
			return err
		} else {
			if c {
				if _, err := createDevice(systemKey, device, client); err != nil {
					return err
				}
			} else {
				fmt.Printf("Device will not be created.\n")
				return nil
			}
		}
	} else {
		delete(device, "name")
		if _, err := client.UpdateDevice(systemKey, deviceName, device); err != nil {
			return err
		}
	}
	deviceRoles, err := getDeviceRoles(deviceName)
	if err != nil {
		return err
	}
	backendDeviceRoles, err := pullDeviceRoles(systemKey, deviceName, client)
	if err != nil {
		return err
	}
	roleDiff := diffRoles(deviceRoles, convertStringSliceToInterfaceSlice(backendDeviceRoles))
	return client.UpdateDeviceRoles(systemKey, deviceName, convertInterfaceSliceToStringSlice(roleDiff.add), convertInterfaceSliceToStringSlice(roleDiff.remove))
}

func updateEdge(systemKey string, edge map[string]interface{}, client *cb.DevClient) error {
	edgeName := edge["name"].(string)
	delete(edge, "name")
	delete(edge, "edge_key")
	delete(edge, "isConnected")
	delete(edge, "novi_system_key")
	delete(edge, "broker_auth_port")
	delete(edge, "broker_port")
	delete(edge, "broker_tls_port")
	delete(edge, "broker_ws_auth_port")
	delete(edge, "broker_ws_port")
	delete(edge, "broker_wss_port")
	delete(edge, "communication_style")
	delete(edge, "first_talked")
	delete(edge, "last_talked")
	delete(edge, "local_addr")
	delete(edge, "local_port")
	delete(edge, "public_addr")
	delete(edge, "public_port")
	delete(edge, "location")
	delete(edge, "mac_address")
	if edge["description"] == nil {
		edge["description"] = ""
	}

	originalColumns := make(map[string]interface{})
	customColumns := make(map[string]interface{})
	for columnName, value := range edge {
		switch strings.ToLower(columnName) {
		case "system_key", "system_secret", "token", "description", "location", "mac_address", "policy_name", "resolver_func", "sync_edge_tables", "last_seen_version":
			originalColumns[columnName] = value
			break
		default:
			customColumns[columnName] = value
			break
		}
	}

	_, err := client.GetEdge(systemKey, edgeName)
	if err != nil {
		// Edge does not exist
		fmt.Printf("Could not update edge '%s'. Error is - %s\n", edgeName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new edge named %s?", edgeName))
		if err != nil {
			return err
		} else {
			if c {
				if _, err := client.CreateEdge(systemKey, edgeName, originalColumns); err != nil {
					return fmt.Errorf("Could not create edge %s: %s", edgeName, err.Error())
				} else {
					fmt.Printf("Successfully created new edge %s\n", edgeName)
				}
				_, err = client.UpdateEdge(systemKey, edgeName, customColumns)
				if err != nil {
					return err
				} else {
					return nil
				}
			} else {
				fmt.Printf("Edge will not be created.\n")
			}
		}
	} else {
		client.UpdateEdge(systemKey, edgeName, edge)
	}
	return nil
}

func updatePortal(systemKey string, portal map[string]interface{}, client *cb.DevClient) error {
	portalName := portal["name"].(string)
	delete(portal, "system_key")
	if portal["description"] == nil {
		portal["description"] = ""
	}
	if portal["config"] == nil {
		portal["config"] = "{\"version\":1,\"allow_edit\":true,\"plugins\":[],\"panes\":[],\"datasources\":[],\"columns\":null}"
	} else {
		rawConfig, _ := json.Marshal(portal["config"])
		portal["config"] = string(rawConfig)
	}

	_, err := client.GetPortal(systemKey, portalName)
	if err != nil {
		// Portal DNE
		fmt.Printf("Could not update portal '%s'. Error is - %s\n", portalName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new portal named %s?", portalName))
		if err != nil {
			return err
		}

		if c {
			if _, err := client.CreatePortal(systemKey, portalName, portal); err != nil {
				return fmt.Errorf("Could not create portal %s: %s", portalName, err.Error())
			}
			fmt.Printf("Successfully created new portal %s\n", portalName)
		} else {
			fmt.Printf("Portal will not be created.\n")
		}
	} else {
		client.UpdatePortal(systemKey, portalName, portal)
	}

	return nil
}

func updatePlugin(systemKey string, plugin map[string]interface{}, client *cb.DevClient) error {
	pluginName := plugin["name"].(string)

	_, err := client.GetPlugin(systemKey, pluginName)
	if err != nil {
		// plugin DNE
		fmt.Printf("Could not find plugin %s\n", pluginName)
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new plugin named %s?", pluginName))
		if err != nil {
			return err
		} else {
			if c {
				if _, err := client.CreatePlugin(systemKey, plugin); err != nil {
					return fmt.Errorf("Could not create plugin %s: %s", pluginName, err.Error())
				} else {
					fmt.Printf("Successfully created new plugin %s\n", pluginName)
				}
			} else {
				fmt.Printf("Plugin will not be created.\n")
			}
		}
	} else {
		client.UpdatePlugin(systemKey, pluginName, plugin)
	}

	return nil
}

func updateAdaptor(adaptor *models.Adaptor) error {
	return adaptor.UpdateAllInfo()
}

func handleUpdateAdaptor(systemKey string, adaptor *models.Adaptor, client *cb.DevClient) error {
	adaptorName := adaptor.Name

	_, err := client.GetAdaptor(systemKey, adaptorName)
	if err != nil {
		// adaptor DNE
		fmt.Printf("Could not update adapter '%s'. Error is - %s\n", adaptorName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new adapter named %s?", adaptorName))
		if err != nil {
			return err
		} else {
			if c {
				if err := createAdaptor(adaptor); err != nil {
					return fmt.Errorf("Could not create adapter %s: %s", adaptorName, err.Error())
				} else {
					fmt.Printf("Successfully created new adapter %s\n", adaptorName)
				}
			} else {
				fmt.Printf("Adapter will not be created.\n")
			}
		}
	} else {
		return updateAdaptor(adaptor)
	}

	return nil
}

func findService(systemKey, serviceName string) (map[string]interface{}, error) {
	services, err := getServices()
	if err != nil {
		return nil, err
	}
	for _, service := range services {
		if service["name"] == serviceName {
			return service, nil
		}
	}
	return nil, fmt.Errorf(NotExistErrorString)
}

func updateServiceWithRunAs(systemKey, name string, service map[string]interface{}, client *cb.DevClient) error {
	if savedRunAs, ok := service[runUserKey].(string); ok {
		if id, err := getUserIdByEmail(savedRunAs); err == nil {
			service[runUserKey] = id
		} else if savedRunAs != "" {
			service[runUserKey] = ""
			logWarning(fmt.Sprintf("Failed to retrieve run_user ID for email '%s'. Please check to make sure that the user exists and that there is a matching entry in .cb-cli/map-name-to-id/users.json. Empty value will be used for run_user", savedRunAs))
		}
	}

	return updateService(systemKey, name, service, client)
}

func updateService(systemKey, name string, service map[string]interface{}, client *cb.DevClient) error {
	if _, err := pullService(systemKey, name, client); err != nil {
		fmt.Printf("Could not find service '%s'. Error is - %s\n", name, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new service named %s?", name))
		if err != nil {
			return err
		} else {
			if c {
				if err := createService(systemKey, service, client); err != nil {
					return fmt.Errorf("Could not create service %s: %s", name, err.Error())
				} else {
					fmt.Printf("Successfully created new service %s\n", name)
				}
			} else {
				fmt.Printf("Service will not be created.\n")
			}
		}
	} else {

		svcCode := service["code"].(string)

		extra := getServiceBody(service)
		_, err := client.UpdateServiceWithBody(systemKey, name, svcCode, extra)
		if err != nil {
			return err
		}

	}
	return nil
}

func getServiceBody(service map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{
		"logging_enabled":   false,
		"execution_timeout": 60,
		"parameters":        make([]interface{}, 0),
		"auto_balance":      false,
		"auto_restart":      false,
		"concurrency":       0,
		"dependencies":      "",
		runUserKey:          "",
	}
	if loggingEnabled, ok := service["logging_enabled"]; ok {
		ret["logging_enabled"] = loggingEnabled
	}
	if executionTimeout, ok := service["execution_timeout"].(float64); ok {
		ret["execution_timeout"] = executionTimeout
	}
	if parameters, ok := service["params"].([]interface{}); ok { // GET for a service returns 'params' but POST/PUT expect 'parameters'
		ret["parameters"] = mkSvcParams(parameters)
	}
	if dependencies, ok := service["dependencies"].(string); ok {
		ret["dependencies"] = dependencies
	}
	if autoBalance, ok := service["auto_balance"].(bool); ok {
		ret["auto_balance"] = autoBalance
	}
	if autoRestart, ok := service["auto_restart"].(bool); ok {
		ret["auto_restart"] = autoRestart
	}
	if concurrency, ok := service["concurrency"].(float64); ok {
		ret["concurrency"] = concurrency
	}
	if runUser, ok := service[runUserKey].(string); ok {
		ret[runUserKey] = runUser
	}
	return ret
}

func createService(systemKey string, service map[string]interface{}, client *cb.DevClient) error {
	svcName := service["name"].(string)
	if ServiceName != "" {
		svcName = ServiceName
	}
	svcCode := service["code"].(string)
	extra := getServiceBody(service)
	if err := client.NewServiceWithBody(systemKey, svcName, svcCode, extra); err != nil {
		return err
	}
	if enableLogs(service) {
		if err := client.EnableLogsForService(systemKey, svcName); err != nil {
			return err
		}
	}
	return nil
}

func updateLibrary(systemKey string, library map[string]interface{}, client *cb.DevClient) error {
	libName := library["name"].(string)

	if _, err := pullLibrary(systemKey, libName, client); err != nil {
		fmt.Printf("Could not find library '%s'. Error is - %s\n", libName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new library named %s?", libName))
		if err != nil {
			return err
		} else {
			if c {
				library["name"] = libName
				if err := createLibrary(systemKey, library, client); err != nil {
					return fmt.Errorf("Could not create library %s: %s", libName, err.Error())
				} else {
					fmt.Printf("Successfully created new library %s\n", libName)
				}
			} else {
				fmt.Printf("Library will not be created.\n")
			}
		}
	} else {

		delete(library, "name")
		delete(library, "version")
		if _, err := client.UpdateLibrary(systemKey, libName, library); err != nil {
			return err
		}
	}
	return nil
}

func createLibrary(systemKey string, library map[string]interface{}, client *cb.DevClient) error {
	libName := library["name"].(string)
	if LibraryName != "" {
		libName = LibraryName
	}
	delete(library, "name")
	delete(library, "version")
	if _, err := client.CreateLibrary(systemKey, libName, library); err != nil {
		return fmt.Errorf("Could not create library %s: %s", libName, err.Error())
	}
	return nil
}

func updateCollection(meta *System_meta, collection map[string]interface{}, client *cb.DevClient) error {
	collection_name, ok := collection["name"].(string)
	if !ok {
		return fmt.Errorf("No name in collection json file: %+v\n", collection)
	}

	_, err := client.GetDataTotalByName(meta.Key, collection_name, cb.NewQuery())
	if err != nil {
		fmt.Printf("Could not find collection '%s'. Error is - %s\n", collection_name, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new collection named %s?", collection_name))
		if err != nil {
			return err
		} else {
			if c {
				if _, err := CreateCollection(meta.Key, collection, client); err != nil {
					return fmt.Errorf("Could not create collection %s: %s", collection_name, err.Error())
				} else {
					fmt.Printf("Successfully created new collection. Updating local copy... %s\n", collection_name)
					return PullAndWriteCollection(meta, collection_name, client, true, true)
				}
			} else {
				fmt.Printf("Collection will not be created.\n")
				return nil
			}
		}
	}

	// here's our workflow for updating a collection:
	// 1) diff and update the collection schema
	// 2) attempt to update all of our items
	// 3) if update fails, we assume the item doesn't exist, so we create the item
	if err := pushCollectionSchema(meta, client, collection_name); err != nil {
		return err
	}

	items := collection["items"].([]interface{})
	for _, row := range items {
		query := cb.NewQuery()
		query.EqualTo("item_id", row.(map[string]interface{})["item_id"])
		if resp, err := client.UpdateDataByName(meta.Key, collection_name, query, row.(map[string]interface{})); err != nil {
			fmt.Printf("Error updating item '%s'. Skipping. Error is - %s\n", row.(map[string]interface{})["item_id"], err.Error())
		} else if resp.Count == 0 {
			if _, err := client.CreateDataByName(meta.Key, collection_name, row.(map[string]interface{})); err != nil {
				return fmt.Errorf("Failed to create item. Error is - %s", err.Error())
			}
		}
	}
	return nil
}

type CollectionInfo struct {
	ID   string
	Name string
}

type RoleInfo struct {
	ID   string
	Name string
}

func CreateCollection(systemKey string, collection map[string]interface{}, client *cb.DevClient) (CollectionInfo, error) {
	collectionName := collection["name"].(string)
	isConnect := isConnectCollection(collection)
	var colId string
	var err error
	if isConnect {
		col, err := cb.GenerateConnectCollection(collection)
		if err != nil {
			return CollectionInfo{}, err
		}
		colId, err = client.NewConnectCollection(systemKey, col)
		if err != nil {
			return CollectionInfo{}, err
		}
	} else {
		colId, err = client.NewCollection(systemKey, collectionName)
		if err != nil {
			return CollectionInfo{}, err
		}
	}

	myInfo := CollectionInfo{
		ID:   colId,
		Name: collectionName,
	}
	if isConnect {
		return myInfo, nil
	}

	if err := updateCollectionNameToId(myInfo); err != nil {
		logErrorForUpdatingMapFile(getCollectionNameToIdFullFilePath(), err)
	}

	columns := collection["schema"].([]interface{})
	for _, columnIF := range columns {
		column := columnIF.(map[string]interface{})
		colName := column["ColumnName"].(string)
		colType := column["ColumnType"].(string)
		if colName == "item_id" {
			continue
		}
		if err := client.AddColumn(colId, colName, colType); err != nil {
			return CollectionInfo{}, err
		}
	}
	allItems := collection["items"].([]interface{})
	totalItems := len(allItems)
	if totalItems == 0 {
		return myInfo, nil
	}
	if totalItems/DataPageSize > 1000 {
		fmt.Println("Large dataset detected. Recommend increasing page size. Use flag: -data-page-size=1000")
	}

	for i := 0; i < totalItems; i += DataPageSize {

		beginningOfRange := i

		// this will be equal to max index + 1
		// to account for golang #slice conventions
		endOfRange := i + DataPageSize

		// if this is last page, and items on this page are fewer than page size
		if totalItems < endOfRange {
			endOfRange = totalItems
		}

		itemsInThisPage := allItems[beginningOfRange:endOfRange]

		for i, item := range itemsInThisPage {
			itemsInThisPage[i] = item.(map[string]interface{})
		}
		if _, err := client.CreateData(colId, itemsInThisPage); err != nil {
			return CollectionInfo{}, err
		}
	}
	return myInfo, nil
}

func createEdge(systemKey, name string, edge map[string]interface{}, client *cb.DevClient) error {
	originalColumns := make(map[string]interface{})
	customColumns := make(map[string]interface{})
	for columnName, value := range edge {
		switch strings.ToLower(columnName) {
		case "system_key", "system_secret", "token", "description", "location", "mac_address", "policy_name", "resolver_func", "sync_edge_tables", "last_seen_version":
			originalColumns[columnName] = value
		default:
			if value != nil {
				customColumns[columnName] = value
			}
		}
	}
	_, err := client.CreateEdge(systemKey, name, originalColumns)
	if err != nil {
		return err
	}
	if len(customColumns) == 0 {
		return nil
	}

	//  We only do this if there ARE custom columns to create
	_, err = client.UpdateEdge(systemKey, name, customColumns)
	if err != nil {
		return err
	}
	return nil
}

func createDevice(systemKey string, device map[string]interface{}, client *cb.DevClient) (map[string]interface{}, error) {
	var randomActiveKey string
	activeKey, ok := device["active_key"].(string)
	if !ok {
		// Active key not present in json file. Creating a random one
		fmt.Printf(" Active key not present. Creating a random one for device creation. Please update the active key from the ClearBlade Console after creation\n")
		randomActiveKey = randSeq(8)
		device["active_key"] = randomActiveKey
	} else {
		if activeKey == "" || len(activeKey) < 6 {
			fmt.Printf("Active is either an empty string or less than 6 characters. Creating a random one for device creation. Please update the active key from the ClearBlade Console after creation\n")
			randomActiveKey = randSeq(8)
			device["active_key"] = randomActiveKey
		}
	}

	originalColumns := make(map[string]interface{})
	customColumns := make(map[string]interface{})
	for columnName, value := range device {
		switch strings.ToLower(columnName) {
		case "name", "type", "state", "description", "enabled", "allow_key_auth", "keys", "active_key", "allow_certificate_auth", "certificate":
			originalColumns[columnName] = value
			break
		default:
			customColumns[columnName] = value
			break
		}
	}
	deviceStuff, err := client.CreateDevice(systemKey, device["name"].(string), originalColumns)
	if err != nil {
		fmt.Printf("CREATE DEVICE ERROR: %s\n", err)
		return nil, err
	}
	_, err = client.UpdateDevice(systemKey, device["name"].(string), customColumns)
	if err != nil {
		fmt.Printf("UPDATE DEVICE ERROR: %s\n", err)
		return nil, err
	}
	return deviceStuff, nil
}

func createPortal(systemKey string, port map[string]interface{}, client *cb.DevClient) (map[string]interface{}, error) {
	delete(port, "system_key")
	if port["description"] == nil {
		port["description"] = ""
	}
	if port["last_updated"] == nil {
		port["last_updated"] = ""
	}
	// Export stores config as dict, but import wants it as a string
	config, ok := port["config"]
	if ok {
		configStr := ""
		switch config.(type) {
		case string:
			configStr = config.(string)
		default:
			configBytes, err := json.Marshal(config)
			if err != nil {
				return nil, err
			}
			configStr = string(configBytes)
		}
		port["config"] = configStr
	}
	portalStuff, err := client.CreatePortal(systemKey, port["name"].(string), port)
	if err != nil {
		return nil, err
	}
	return portalStuff, nil
}

func createPlugin(systemKey string, plug map[string]interface{}, client *cb.DevClient) (map[string]interface{}, error) {
	return client.CreatePlugin(systemKey, plug)
}

func createAdaptor(adap *models.Adaptor) error {
	return adap.UploadAllInfo()
}

func updateRole(systemKey string, role map[string]interface{}, collectionsInfo []CollectionInfo, client *cb.DevClient) error {
	roleName := role["Name"].(string)

	if _, err := pullRole(systemKey, roleName, client); err != nil {
		fmt.Printf("Could not find role '%s'. Error is - %s\n", roleName, err.Error())
		c, err := confirmPrompt(fmt.Sprintf("Would you like to create a new role named %s?", roleName))
		if err != nil {
			return err
		} else {
			if c {
				if err := createRole(systemKey, role, collectionsInfo, client); err != nil {
					return fmt.Errorf("Could not create role %s: %s", roleName, err.Error())
				} else {
					fmt.Printf("Successfully created new role %s\n", roleName)
				}
			} else {
				fmt.Printf("Role will not be created.\n")
			}
		}
	} else {
		roleID, err := getRoleIdByName(roleName)
		if err != nil {
			return fmt.Errorf("Error updating role: %s", err.Error())
		}
		updateRoleBody, err := packageRoleForUpdate(roleID, role, collectionsInfo)
		if err != nil {
			return err
		}
		if err := client.UpdateRole(systemKey, roleName, updateRoleBody); err != nil {
			if byts, err := json.Marshal(updateRoleBody); err == nil {
				fmt.Printf("Failed to update role '%s'. Request body is - \n%s\n", roleName, string(byts))
			}
			return err
		}
	}
	return nil
}
