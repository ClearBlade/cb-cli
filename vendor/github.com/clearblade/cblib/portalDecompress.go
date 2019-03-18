package cblib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	cb "github.com/clearblade/Go-SDK"
)

const outFile = "index"
const htmlKey = "HTML"
const javascriptKey = "JavaScript"
const cssKey = "CSS"
const dynamicDataType = "DYNAMIC_DATA_TYPE"
const portalConfigDirectory = "config"
const datasourceDirectory = "datasources"
const widgetsDirectory = "widgets"

func init() {

	usage :=
		`
	Compresses or decompresses Portal code
	`

	example :=
		`
	cb-cli decompress -portalName=portal1		#
	`

	decompressCommand := &SubCommand{
		name:         "decompress",
		usage:        usage,
		needsAuth:    false,
		mustBeInRepo: true,
		run:          decompress,
		example:      example,
	}
	decompressCommand.flags.StringVar(&PortalName, "portal", "", "Name of Portal to decompress after editing")
	AddCommand("decompress", decompressCommand)
}

func decompress(cmd *SubCommand, client *cb.DevClient, args ...string) error {
	if err := checkPortalCodeManagerArgsAndFlags(args); err != nil {
		return err
	}
	SetRootDir(".")

	return cleanUpAndDecompress(PortalName)

}

func cleanUpAndDecompress(name string) error {
	portal, err := getPortal(name)
	if err != nil {
		return err
	}

	if err = os.RemoveAll(filepath.Join(portalsDir, name, portalConfigDirectory)); err != nil {
		return err
	}

	if err = decompressDatasources(portal); err != nil {
		return err
	}
	if err = decompressWidgets(portal); err != nil {
		return err
	}
	return nil
}

func checkPortalCodeManagerArgsAndFlags(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("There are no arguments to the update command, only command line options")
	}
	return nil
}

func decompressDatasources(portal map[string]interface{}) error {
	var (
		portalName          string
		config, datasources map[string]interface{}
		ok                  bool
	)

	if portalName, ok = portal["name"].(string); !ok {
		return fmt.Errorf("Portal 'name' key missing in <Portal>.json file")
	}
	if config, ok = portal["config"].(map[string]interface{}); !ok {
		return fmt.Errorf("Portal 'config' key missing in <Portal>.json file")
	}
	if datasources, ok = config["datasources"].(map[string]interface{}); !ok {
		return fmt.Errorf("No Datasources defined in 'config' ")
	}

	for _, ds := range datasources {
		dataSourceData := ds.(map[string]interface{})
		dataSourceName := dataSourceData["name"].(string)
		if err := writeDatasource(portalName, dataSourceName, dataSourceData); err != nil {
			return err
		}
	}
	return nil
}

func writeDatasource(portalName, dataSourceName string, data map[string]interface{}) error {
	currentFileName := dataSourceName
	currDsDir := filepath.Join(portalsDir, portalName, portalConfigDirectory, datasourceDirectory)
	if err := os.MkdirAll(currDsDir, 0777); err != nil {
		return err
	}
	return writeEntity(currDsDir, currentFileName, data)
}

func decompressWidgets(portal map[string]interface{}) error {
	var (
		portalName      string
		config, widgets map[string]interface{}
		ok              bool
	)

	if portalName, ok = portal["name"].(string); !ok {
		return fmt.Errorf("Portal 'name' key missing in <Portal>.json file")
	}
	if config, ok = portal["config"].(map[string]interface{}); !ok {
		return fmt.Errorf("Portal 'config' key missing in <Portal>.json file")
	}
	if widgets, ok = config["widgets"].(map[string]interface{}); !ok {
		logInfo("No widgets defined in 'config'")
	}

	for _, widgetConfig := range widgets {
		widgetData := widgetConfig.(map[string]interface{})
		widgetName := getOrGenerateWidgetName(widgetData)
		if err := writeWidget(portalName, widgetName, widgetData); err != nil {
			return err
		}
	}
	return nil
}

func getOrGenerateWidgetName(widgetData map[string]interface{}) string {
	widgetID := widgetData["id"].(string)
	widgetType := widgetData["type"].(string)
	name := fmt.Sprintf("%s"+"_"+"%v", widgetType, widgetID)
	return name
}

func writeParserBasedOnDataType(dataType string, setting map[string]interface{}, filePath string) error {
	found := false
	if ip, ok := setting["incoming_parser"].(map[string]interface{}); ok {
		found = true
		if dataType != dynamicDataType {
			ip = setting
		}
		if err := writeParserFiles(filePath+"/incoming_parser", ip); err != nil {
			return err
		}
	}

	if op, ok := setting["outgoing_parser"].(map[string]interface{}); ok {
		found = true
		if dataType != dynamicDataType {
			op = setting
		}
		if err := writeParserFiles(filePath+"/outgoing_parser", op); err != nil {
			return err
		}
	}

	if !found {
		if _, okMap := setting["value"]; okMap {
			if err := writeParserFiles(filePath+"/incoming_parser", setting); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeWidget(portalName, widgetName string, data map[string]interface{}) error {
	currWidgetDir := filepath.Join(portalsDir, portalName, portalConfigDirectory, widgetsDirectory, widgetName)

	return actOnParserSettings(data, func(settingName, dataType string, parserSetting map[string]interface{}) error {
		if err := writeParserBasedOnDataType(dataType, parserSetting, currWidgetDir+"/"+settingName); err != nil {
			return err
		}
		return nil
	})
}

func writeParserFiles(currWidgetDir string, data map[string]interface{}) error {
	keysToIgnoreInData := map[string]interface{}{}
	absFilePath := filepath.Join(currWidgetDir, outFile)

	switch data["value"].(type) {
	case string:
		if err := writeFile(absFilePath+".js", data["value"].(string)); err != nil {
			return err
		}
	case map[string]interface{}:
		if err := writeWebFiles(absFilePath, data["value"].(map[string]interface{}), keysToIgnoreInData); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

func writeWebFiles(absFilePath string, data, keysToIgnoreInData map[string]interface{}) error {

	outjs := recursivelyFindValueForKey(javascriptKey, data, keysToIgnoreInData)
	outhtml := recursivelyFindValueForKey(htmlKey, data, keysToIgnoreInData)
	outcss := recursivelyFindValueForKey(cssKey, data, keysToIgnoreInData)
	if outhtml != nil {
		if err := writeFile(absFilePath+".html", outhtml.(interface{})); err != nil {
			return err
		}
	}

	if outjs != nil {
		if err := writeFile(absFilePath+".js", outjs.(interface{})); err != nil {
			return err
		}
	}

	if outcss != nil {
		if err := writeFile(absFilePath+".css", outcss.(interface{})); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(absFilePath string, data interface{}) error {
	if data == nil {
		return nil
	}
	outDir := filepath.Dir(absFilePath)
	if err := os.MkdirAll(outDir, 0777); err != nil {
		return err
	}
	switch data.(type) {
	case string:
		if err := ioutil.WriteFile(absFilePath, []byte(data.(string)), 0666); err != nil {
			return err
		}
	case map[string]interface{}:
		marshalled, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return fmt.Errorf("Could not marshall object: %s", err.Error())
		}
		if err := ioutil.WriteFile(absFilePath, []byte(marshalled), 0666); err != nil {
			return err
		}
	}
	return nil
}

func recursivelyFindValueForKey(queryKey string, data map[string]interface{}, keysToIgnoreInData map[string]interface{}) interface{} {
	for k, v := range data {
		if k == queryKey {
			return v
		}
		switch v.(type) {
		case map[string]interface{}:
			if keysToIgnoreInData[k] != nil {
				continue
			}
			val := recursivelyFindValueForKey(queryKey, v.(map[string]interface{}), keysToIgnoreInData)
			if val != nil {
				return val
			}
		default:
			continue
		}
	}
	return nil
}

func isFileEmpty(absFilePath string) bool {
	if fileInfo, err := os.Stat(absFilePath); err == nil {
		log.Println("Is File Empty", fileInfo.Size())
		if fileInfo.Size() == 0 {
			return true
		}
	}
	return false
}
