package cblib

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	cb "github.com/clearblade/Go-SDK"
)

type Stack struct {
	name      string
	stringRep string
	stack     []string
}

var (
	names            *Stack
	ignores          map[string][]string
	uniqueKeys       map[string]string
	suppressErrors   []int
	printedDiffCount int
)

func init() {

	usage :=
		`
	Perform a diff operation between your local assets and the remote assets in the ClearBlade Platform. For example, see what changes have been made in the platform since your last export, or pull.
	`

	example :=
		`
	cb-cli diff -all-services                           # Diffs all your local services against all the services in the platform
	cb-cli diff -collection=someone_modified_coll       # Shows diff between remote and local versions of the collection 'someone_modified_coll'
	`

	printedDiffCount = 0
	suppressErrors = []int{0}
	names = NewStack("names")
	ignores = map[string][]string{
		"system.json":           []string{"platformURL", "platform_url"},
		"system.json:data":      []string{"appID", "collectionID", "app_id", "collection_id"},
		"system.json:libraries": []string{"version", "system_key", "library_key"},
		"system.json:services":  []string{"current_version"},
		"users.json":            []string{"user_id", "creation_date"},
		"trigger":               []string{"system_key", "system_secret"},
		"timer":                 []string{"system_key", "system_secret"},
		"service":               []string{"source"},
		"library":               []string{"source"},
	}
	uniqueKeys = map[string]string{
		"role:Permissions:CodeServices": "Name",
		"system.json:data":              "name",
		"system.json:data:schema":       "ColumnName",
		"system.json:libraries":         "name",
		"system.json:services":          "name",
		"system.json:timers":            "name",
		"system.json:triggers":          "name",
		"system.json:users":             "ColumnName",
		"users.json":                    "email",
		"UserSchema:columns":            "ColumnName",
	}
	myDiffCommand := &SubCommand{
		name:         "diff",
		usage:        usage,
		needsAuth:    true,
		mustBeInRepo: true,
		run:          doDiff,
		example:      example,
	}
	myDiffCommand.flags.BoolVar(&UserSchema, "userschema", false, "diff user table schema")
	myDiffCommand.flags.BoolVar(&AllServices, "all-services", false, "diff all of the services stored locally")
	myDiffCommand.flags.BoolVar(&AllLibraries, "all-libraries", false, "diff all of the libraries stored locally")
	myDiffCommand.flags.StringVar(&ServiceName, "service", "", "Name of service to diff")
	myDiffCommand.flags.StringVar(&LibraryName, "library", "", "Name of library to diff")
	myDiffCommand.flags.StringVar(&CollectionName, "collection", "", "Name of collection to diff")
	myDiffCommand.flags.StringVar(&User, "user", "", "Name of user to diff")
	myDiffCommand.flags.StringVar(&RoleName, "role", "", "Name of role to diff")
	myDiffCommand.flags.StringVar(&TriggerName, "trigger", "", "Name of trigger to diff")
	myDiffCommand.flags.StringVar(&TimerName, "timer", "", "Name of timer to diff")
	myDiffCommand.flags.StringVar(&TempDir, "temp-dir", "", "Temporary dir to place diff files")
	AddCommand("diff", myDiffCommand)
}

func checkDiffArgsAndFlags(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("There are no arguments to the diff command, only command line options\n")
	}

	if AllServices && ServiceName != "" {
		return fmt.Errorf("Cannot specify both -all-services and -service=<service_name>\n")
	}

	if AllLibraries && LibraryName != "" {
		return fmt.Errorf("Cannot specify both -all-libraries and -library=<library_name>\n")
	}
	return nil
}

func doDiff(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if err := checkDiffArgsAndFlags(args); err != nil {
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

	if UserSchema {
		if err := diffUserSchema(systemInfo, client); err != nil {
			return err
		}
	}

	if AllServices {
		if err := diffAllServices(systemInfo, client); err != nil {
			return err
		}
	}

	if ServiceName != "" {
		if err := diffService(systemInfo, client, ServiceName); err != nil {
			return err
		}
	}

	if AllLibraries {
		if err := diffAllLibraries(systemInfo, client); err != nil {
			return err
		}
	}

	if LibraryName != "" {
		if err := diffLibrary(systemInfo, client, LibraryName); err != nil {
			return err
		}
	}

	if CollectionName != "" {
		if err := diffCollection(systemInfo, client, CollectionName); err != nil {
			return err
		}
	}
	if User != "" {
		if err := diffUser(systemInfo, client, User); err != nil {
			return err
		}
	}
	if RoleName != "" {
		if err := diffRole(systemInfo, client, RoleName); err != nil {
			return err
		}
	}
	if TriggerName != "" {
		if err := diffTrigger(systemInfo, client, TriggerName); err != nil {
			return err
		}
	}
	if TimerName != "" {
		if err := diffTimer(systemInfo, client, TimerName); err != nil {
			return err
		}
	}
	return nil
}

func dumbDownSchemaColumns(cols []map[string]interface{}) []interface{} {
	rval := make([]interface{}, len(cols))
	for idx, val := range cols {
		rval[idx] = val
	}
	return rval
}

func diffUserSchema(sys *System_meta, client *cb.DevClient) error {
	localSchema, err := getUserSchema()
	if err != nil {
		return err
	}
	remoteSchema, err := pullUserSchemaInfo(sys.Key, client, false)
	remoteSchema["columns"] = dumbDownSchemaColumns(remoteSchema["columns"].([]map[string]interface{}))
	if err != nil {
		return err
	}
	names.push("UserSchema")
	defer names.pop()
	printedDiffCount = 0
	diffMap(localSchema, remoteSchema)
	printSummary("user schema", "schema.json")
	return nil
}

type LocalFunc func(name string) (map[string]interface{}, error)
type RemoteFunc func(key, name string, client *cb.DevClient) (map[string]interface{}, error)

func diffCodeAndMeta(sys *System_meta, client *cb.DevClient, thangType, thangName string, lf LocalFunc, rf RemoteFunc) error {
	localThang, err := lf(thangName)
	if err != nil {
		return err
	}

	remoteThang, err := rf(sys.Key, thangName, client)
	if err != nil {
		return err
	}
	lCode := localThang["code"].(string)
	rCode := remoteThang["code"].(string)

	if lCode[len(lCode)-1] != '\n' {
		lCode = lCode + "\n"
	}
	if rCode[len(rCode)-1] != '\n' {
		rCode = rCode + "\n"
	}
	delete(localThang, "code")
	delete(remoteThang, "code")
	delete(remoteThang, "system_key")
	delete(remoteThang, "library_key")

	myPid := os.Getpid()

	tempDir := os.TempDir()
	if TempDir != "" {
		tempDir = TempDir
	}
	fmt.Println("Using Temp Dir: " + tempDir)

	localFile := fmt.Sprintf("%s%d-local.js", tempDir, myPid)
	remoteFile := fmt.Sprintf("%s%d-remote.js", tempDir, myPid)

	if err = ioutil.WriteFile(localFile, []byte(lCode), 0666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(remoteFile, []byte(rCode), 0666); err != nil {
		return err
	}

	var diffCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		diffCmd = exec.Command("FC", localFile, remoteFile)
	} else {
		diffCmd = exec.Command("/usr/bin/diff", localFile, remoteFile)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	diffCmd.Stdout = &stdout
	diffCmd.Stderr = &stderr
	names.push(thangType)
	defer names.pop()
	if err := diffCmd.Run(); err != nil && err.Error() != "exit status 1" {
		return fmt.Errorf("Internal error, exec failed: %s: %s", err.Error(), stderr.String())
	}
	if stdout.String() != "" {
		printErr("Local version of code for '%s' is different from remote:\n%s\n", thangName, stdout.String())
	}

	os.Remove(localFile)
	os.Remove(remoteFile)

	// Now, diff the meta data...
	printedDiffCount = 0
	diffMap(localThang, remoteThang)
	printSummary(thangType+" meta", thangName)
	return nil
}

func diffAllServices(sys *System_meta, client *cb.DevClient) error {
	services, err := getServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		svcName := service["name"].(string)
		fmt.Printf("Diffing service %s\n", svcName)
		if err := diffService(sys, client, svcName); err != nil {
			return err
		}
	}
	return nil
}

func diffService(sys *System_meta, client *cb.DevClient, serviceName string) error {
	return diffCodeAndMeta(sys, client, "service", serviceName, getService, pullService)
}

func diffAllLibraries(sys *System_meta, client *cb.DevClient) error {
	libraries, err := getLibraries()
	if err != nil {
		return err
	}
	for _, library := range libraries {
		libName := library["name"].(string)
		fmt.Printf("Diffing library %s\n", libName)
		if err := diffLibrary(sys, client, libName); err != nil {
			return err
		}
	}
	return nil
}

func diffLibrary(sys *System_meta, client *cb.DevClient, libraryName string) error {
	return diffCodeAndMeta(sys, client, "library", libraryName, getLibrary, pullLibrary)
}

func diffCollection(sys *System_meta, client *cb.DevClient, collectionName string) error {
	localCollection, err := getCollection(collectionName)
	if err != nil {
		return err
	}

	colId, ok := localCollection["collectionID"].(string)
	if !ok {
		colId = localCollection["collection_id"].(string)
	}
	ExportRows = false
	remoteCollection, err := pullCollectionAndInfo(sys, colId, client)
	if err != nil {
		return err
	}
	delete(localCollection, "items")
	delete(remoteCollection, "items")
	names.push("collection")
	defer names.pop()
	printedDiffCount = 0
	diffMap(localCollection, remoteCollection)
	printSummary("collection", collectionName)
	return nil
}

func diffUser(sys *System_meta, client *cb.DevClient, userName string) error {
	ExportUsers = true
	allUsers, err := pullUsers(sys, client, false)
	if err != nil {
		return err
	}
	var remoteUser map[string]interface{} = nil
	for _, user := range allUsers {
		if user["email"].(string) == userName {
			remoteUser = user
			remoteUser["roles"] = unMungeRoles(remoteUser["roles"].([]string))
			break
		}
	}
	if remoteUser == nil {
		return fmt.Errorf("Remote user '%s' not found", userName)
	}
	localUser, err := getUser(userName + ".json")
	if err != nil {
		return err
	}
	names.push("user")
	defer names.pop()
	printedDiffCount = 0
	diffMap(localUser, remoteUser)
	printSummary("user", userName)
	return nil
}

func diffRole(sys *System_meta, client *cb.DevClient, roleName string) error {
	localRole, err := getRole(roleName + ".json")
	if err != nil {
		return err
	}

	daRoles, err := pullRoles(sys.Key, client, false)
	if err != nil {
		return err
	}

	var remoteRole map[string]interface{}
	for _, daRole := range daRoles {
		if daRole["Name"].(string) == roleName {
			remoteRole = daRole
			break
		}
	}
	if remoteRole == nil {
		return fmt.Errorf("Could not find remote role '%s'\n", roleName)
	}
	names.push("role")
	defer names.pop()
	printedDiffCount = 0
	diffMap(localRole, remoteRole)
	printSummary("role", roleName)
	return nil
}

func diffTrigger(sys *System_meta, client *cb.DevClient, triggerName string) error {
	localTrigger, err := getTrigger(triggerName + ".json")
	if err != nil {
		return err
	}
	remoteTrigger, err := pullTrigger(sys.Key, triggerName, client)
	if err != nil {
		return err
	}
	names.push("trigger")
	defer names.pop()
	printedDiffCount = 0
	diffMap(localTrigger, remoteTrigger)
	printSummary("trigger", triggerName)
	return nil
}

func diffTimer(sys *System_meta, client *cb.DevClient, timerName string) error {
	localTimer, err := getTimer(timerName + ".json")
	if err != nil {
		return err
	}
	remoteTimer, err := pullTimer(sys.Key, timerName, client)
	if err != nil {
		return err
	}
	names.push("timer")
	defer names.pop()
	printedDiffCount = 0
	diffMap(localTimer, remoteTimer)
	printSummary("timer", timerName)
	return nil
}

func printErr(strFmt string, args ...interface{}) {
	if showErrors() {
		printedDiffCount++
		newArgs := append([]interface{}{names.stringRep}, args...)
		fmt.Printf("In %s: "+strFmt, newArgs...)
	}
}

func NewStack(name string) *Stack {
	return &Stack{
		name:      name,
		stringRep: "",
		stack:     make([]string, 0),
	}
}

func (s *Stack) push(item string) {
	s.stack = append(s.stack, item)
	s.stringRep = strings.Join(s.stack, ":")
}

func (s *Stack) top() (string, error) {
	rval := ""
	if len(s.stack) == 0 {
		return rval, fmt.Errorf("Attempt to get top of stack for empty stack %s", s.name)
	}
	return s.stack[len(s.stack)-1], nil
}

func (s *Stack) pop() (string, error) {
	rval := ""
	if len(s.stack) == 0 {
		return rval, fmt.Errorf("Attempt to pop empty stack %s", s.name)
	}
	rval, s.stack = s.stack[len(s.stack)-1], s.stack[:len(s.stack)-1]
	s.stringRep = strings.Join(s.stack, ":")
	return rval, nil
}

func diffSystemDotJSON(a, b map[string]interface{}) int {
	names.push("system.json")
	defer names.pop()
	diffMap(a, b)
	fmt.Printf("%d Total Errors\n", printedDiffCount)
	return printedDiffCount
}

func diffUsersDotJSON(a, b []interface{}) int {
	names.push("users.json")
	defer names.pop()
	return diffSlice(a, b)
}

func valuesAreEqual(a, b interface{}) bool {
	aVal := a
	bVal := b
	switch a.(type) {
	case float64:
		aVal = int(a.(float64))
	}
	switch b.(type) {
	case float64:
		bVal = int(b.(float64))
	}
	return aVal == bVal
}

func diffUnknownTypes(key string, a, b interface{}) int {
	if !sameTypes(a, b) {
		return 1
	}
	if a == nil {
		return 0
	} else if outerType(a) == "map" {
		if key != "" {
			names.push(key)
			defer names.pop()
		}
		return diffMap(a.(map[string]interface{}), b.(map[string]interface{}))
	} else if outerType(a) == "slice" {
		if key != "" {
			names.push(key)
			defer names.pop()
		}
		//return diffSlice(a.([]interface{}), b.([]interface{}))
		return diffSlice(a, b)
	} else if valuesAreEqual(a, b) {
		return 0
	}
	printErr("Found differing values for '%s': local '%v' != remote '%v'\n", key, a, b)
	return 1
}

func diffMap(a, b map[string]interface{}) int {
	totalErrors := 0
	checkedKeys := map[string]bool{}
	for aKey, aVal := range a {
		checkedKeys[aKey] = true
		if shouldIgnore(aKey) {
			continue
		}
		if bVal, ok := b[aKey]; ok {
			totalErrors += diffUnknownTypes(aKey, aVal, bVal)
		} else {
			totalErrors++
			//printErr("Item '%s' in local version missing in remote version\n", aKey)
		}
	}
	for bKey, _ := range b {
		_, ok := checkedKeys[bKey]
		if shouldIgnore(bKey) || ok {
			continue
		}
		if _, ok := a[bKey]; !ok {
			//printErr("Item '%s' in remote version missing in local version\n", bKey)
			totalErrors++
		}
	}
	return totalErrors
}

func diffSlice(aIF, bIF interface{}) int {
	a := aIF.([]interface{})
	b := bIF.([]interface{})
	if len(a) > 0 {
		if reflect.TypeOf(a[0]).String() == "map[string]interface {}" {
			pushErrorContext()
			defer popErrorContext()
		}
	}
	// Assumption
	totalErrors := 0
	if !sameTypes(a, b) {
		return 1
	}
	if len(a) != len(b) {
		newArgs := append([]interface{}{names.stringRep}, len(a), len(b))
		strFmt := "Slices are of different length: %d != %d\n"
		fmt.Printf("In %s: "+strFmt, newArgs...)
		totalErrors++
	}
	totalErrors += diffTwoSlices(a, b)
	return totalErrors
}

func getUniqueKeyInfo(valSlice []interface{}) (string, bool) {
	if len(valSlice) == 0 {
		return "", false
	}
	oneVal := valSlice[0]
	if outerType(oneVal) != "map" {
		return "", false
	}
	if uniqueKey, haveOne := uniqueKeys[names.stringRep]; haveOne {
		return uniqueKey, true
	}
	return "", false
}

func findMatchInOtherSlice(b []interface{}, uniqueKey string, uniqueVal interface{}) map[string]interface{} {
	for _, bValIF := range b {
		bVal := bValIF.(map[string]interface{})
		if valForKeyIF, ok := bVal[uniqueKey]; ok {
			if uniqueVal == valForKeyIF {
				return bVal
			}
		}
	}
	return nil
}

func valInSlice(val interface{}, slice []interface{}) bool {
	for _, sliceVal := range slice {
		if sliceVal == val {
			return true
		}
	}
	return false
}

func diffKeyedSlices(a, b []interface{}, uniqueKey string) int {
	myErrors := 0
	seenKeyVals := []interface{}{}
	for _, aValIF := range a {
		aVal := aValIF.(map[string]interface{})
		if valForKey, ok := aVal[uniqueKey]; ok {
			seenKeyVals = append(seenKeyVals, valForKey)
			bVal := findMatchInOtherSlice(b, uniqueKey, valForKey)
			if bVal == nil {
				myErrors++
				printErr("Item %s:%v not found in other system\n", uniqueKey, valForKey)
			} else {
				pushErrorContext()
				defer popErrorContext()
				myErrors += diffMap(aVal, bVal)
			}
		} else {
			printErr("Item supposedly with uniqueKey doesn't have one: %#v\n", aVal)
			return -1
		}
	}
	//  Now, we're just finding entries in b that aren't in a
	for _, bValIF := range b {
		bVal := bValIF.(map[string]interface{})
		if valForKey, ok := bVal[uniqueKey]; ok {
			if !valInSlice(valForKey, seenKeyVals) {
				printErr("Key %s with value %v not found in local system\n",
					uniqueKey, valForKey)
				myErrors++
			}
		} else {
			printErr("Item supposedly with uniqueKey doesn't have one: %#v\n", bVal)
			return -1
		}
	}
	return myErrors
}

func diffTwoSlices(a, b []interface{}) int {
	uniqueKey, useUniqueKey := getUniqueKeyInfo(a)
	if useUniqueKey {
		if myErrors := diffKeyedSlices(a, b, uniqueKey); myErrors != -1 {
			return myErrors
		}
	}
	return diffUnkeyedSlices(a, b)
}

func diffUnkeyedSlices(a, b []interface{}) int {
	totalErrors := 0
	printsBefore := printedDiffCount
	for _, aVal := range a {
		found := false
		for _, bVal := range b {
			blockErrors()
			errCount := diffUnknownTypes("", aVal, bVal)
			unblockErrors()
			if errCount == 0 {
				found = true
				break
			}
		}
		if !found {
			totalErrors++
			if printsBefore == printedDiffCount {
				printErr("Could not find item %#v in remote version\n", aVal)
			}
		}
	}

	printsBefore = printedDiffCount
	for _, bVal := range b {
		found := false
		for _, aVal := range a {
			blockErrors()
			errCount := diffUnknownTypes("", bVal, aVal)
			unblockErrors()
			if errCount == 0 {
				found = true
				break
			}
		}
		if !found {
			totalErrors++
			if printsBefore == printedDiffCount {
				printErr("Could not find item %#v in local version\n", bVal)
			}
		}
	}
	return totalErrors
}

func shouldIgnore(key string) bool {
	if keyList, ok := ignores[names.stringRep]; ok {
		for _, ignoreKey := range keyList {
			if ignoreKey == key {
				return true
			}
		}
	}
	return false
}

func sameTypes(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	outerA := outerType(a)
	outerB := outerType(b)
	if outerA == "slice" && outerB == "slice" {
		return true
	}
	if (outerA == "float64" && outerB == "int") || (outerA == "int" && outerB == "float64") {
		return true
	}

	typeA := reflect.TypeOf(a).String()
	typeB := reflect.TypeOf(b).String()
	rval := typeA == typeB
	if !rval {
		printErr("Encountered two different types: %s != %s\n", typeA, typeB)
	}
	return rval
}

func outerType(a interface{}) string {
	return reflect.ValueOf(a).Kind().String()
}

func showErrors() bool {
	return suppressErrors[len(suppressErrors)-1] == 0
}

func pushErrorContext() {
	suppressErrors = append(suppressErrors, 0)
}

func popErrorContext() {
	suppressErrors = suppressErrors[:len(suppressErrors)-1]
}

func blockErrors() {
	suppressErrors[len(suppressErrors)-1] = suppressErrors[len(suppressErrors)-1] + 1
}

func unblockErrors() {
	suppressErrors[len(suppressErrors)-1] = suppressErrors[len(suppressErrors)-1] - 1
}

func printSummary(objectType, objectName string) {
	if printedDiffCount == 0 {
		fmt.Printf("Local version of %s '%s' is the same as the remote version\n", objectType, objectName)
	} else {
		fmt.Printf("Found %d differences between local and remote versions of %s '%s'\n", printedDiffCount, objectType, objectName)
	}
}
