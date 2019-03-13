package cblib

import (
	//"flag"
	"fmt"
	"os"
	"strings"

	cb "github.com/clearblade/Go-SDK"
)

func init() {

	usage :=
		`
	Point your local system to a different remote system within a ClearBlade Platform
	`

	example :=
		`
	cb-cli target
	cb-cli target -url=https://platform.clearblade.com -messaging-url=platform.clearblade.com -system-key=8abcd6aa0baadcd8bbe3fabca29301 -email=dev@dev.com -password=pw
	`
	systemDotJSON = map[string]interface{}{}
	svcCode = map[string]interface{}{}
	myTargetCommand := &SubCommand{
		name:         "target",
		usage:        usage,
		needsAuth:    false,
		mustBeInRepo: true,
		run:          doTarget,
		example:      example,
	}
	myTargetCommand.flags.StringVar(&URL, "url", "", "Clearblade platform url for target system")
	myTargetCommand.flags.StringVar(&MsgURL, "messaging-url", "", "Clearblade messaging url for target system")
	myTargetCommand.flags.StringVar(&SystemKey, "system-key", "", "System Key for target system, ex 9b9eea9c0bda8896a3dab5aeec9601")
	myTargetCommand.flags.StringVar(&Email, "email", "", "Developer email for login")
	myTargetCommand.flags.StringVar(&Password, "password", "", "Developer password")
	AddCommand("target", myTargetCommand)
}

func doTarget(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if len(args) != 0 {
		fmt.Printf("init command takes no arguments; only options: '%+v'\n", args)
		os.Exit(1)
	}
	defaults := setupTargetDefaults()
	oldSysMeta, err := getSysMeta()
	if err != nil {
		return err
	}
	// TODO causes CBCOMM-245
	if err = os.Chdir(".."); err != nil {
		return fmt.Errorf("Could not move up to parent directory: %s", err.Error())
	}
	MetaInfo = nil
	client, err = Authorize(defaults)
	if err != nil {
		return err
	}
	return reallyTarget(client, SystemKey, oldSysMeta)
}

func fixSystemName(sysName string) string {
	return strings.Replace(sysName, " ", "_", -1)
}

func reallyTarget(cli *cb.DevClient, sysKey string, oldSysMeta *System_meta) error {
	sysMeta, err := pullSystemMeta(sysKey, cli)
	if err != nil {
		return err
	}

	fixo, fixn := fixSystemName(oldSysMeta.Name), fixSystemName(sysMeta.Name)
	if fixo != fixn {
		fmt.Printf("Renaming %s to %s\n", fixo, fixn)
		os.Rename(fixo, fixn)
	}
	SetRootDir(fixn)
	if err := setupDirectoryStructure(sysMeta); err != nil {
		return err
	}
	storeMeta(sysMeta)

	if err = storeSystemDotJSON(systemDotJSON); err != nil {
		return err
	}

	metaStuff := map[string]interface{}{
		"platform_url":    URL,
		"messaging_url":   MsgURL,
		"developer_email": Email,
		"token":           cli.DevToken,
	}
	if err = storeCBMeta(metaStuff); err != nil {
		return err
	}

	logInfo("Updating map name to ID files...")
	updateMapNameToIDFiles(sysMeta, cli)

	fmt.Printf("System '%s' has been initialized.\n", sysMeta.Name)
	return nil
}

//
//  This stuff is a hack -- it's used when initing inside a repo to give prompts
//  with defaults. see fillInTheBlanks(..) in newAuth.go
//

type DefaultInfo struct {
	url       string
	email     string
	systemKey string
	msgUrl    string
}

func setupTargetDefaults() *DefaultInfo {
	meta, err := getSysMeta()
	if err != nil || MetaInfo == nil {
		return nil
	}

	platform_url, ok := MetaInfo["platformURL"].(string)
	if !ok {
		platform_url = MetaInfo["platform_url"].(string)
	}
	email, ok := MetaInfo["developerEmail"].(string)
	if !ok {
		email = MetaInfo["developer_email"].(string)
	}
	messaging_url, ok := MetaInfo["messagingURL"].(string)
	if !ok {
		messaging_url = MetaInfo["messaging_url"].(string)
	}

	return &DefaultInfo{
		url:       platform_url,
		email:     email,
		systemKey: meta.Key,
		msgUrl:    messaging_url,
	}
}
