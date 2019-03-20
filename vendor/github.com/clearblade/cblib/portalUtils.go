package cblib

import (
	"encoding/json"

	"github.com/totherme/unstructured"
)

const outFile = "index"
const htmlKey = "HTML"
const javascriptKey = "JavaScript"
const cssKey = "CSS"
const dynamicDataType = "DYNAMIC_DATA_TYPE"
const portalConfigDirectory = "config"
const datasourceDirectory = "datasources"
const widgetsDirectory = "widgets"
const internalResourcesDirectory = "internalResources"
const portalConfigPath = "/config"
const portalWidgetsPath = portalConfigPath + "/widgets"
const portalDatasourcesPath = portalConfigPath + "/datasources"
const portalInternalResourcesPath = portalConfigPath + "/internalResources"
const parsersDirectory = "parsers"
const outgoingParserKey = "outgoing_parser"
const incomingParserKey = "incoming_parser"
const valueKey = "value"
const portalWidgetSettingsFile = "settings.json"
const portalWidgetMetaFile = "meta.json"
const portalInternalResourceMetaFile = "meta.json"

func actOnParserSettings(widgetSettings map[string]interface{}, cb func(string, string) error) error {
	for settingName, v := range widgetSettings {
		switch v.(type) {
		case map[string]interface{}:
			// if there's a dataType property we know this setting is a parser
			if dataType, ok := v.(map[string]interface{})["dataType"].(string); ok {
				if err := cb(settingName, dataType); err != nil {
					return err
				}
			}
		default:
			continue
		}
	}
	return nil
}

func convertPortalMapToUnstructured(p map[string]interface{}) (*unstructured.Data, error) {
	jason, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	portalConfig, err := unstructured.ParseJSON(string(jason))
	if err != nil {
		return nil, err
	}
	return &portalConfig, nil
}

func getPortalWidgetSettingsFile(widgetDir string) (map[string]interface{}, error) {
	return getDict(widgetDir + "/" + portalWidgetSettingsFile)
}

func getPortalWidgetMetaFile(widgetDir string) (map[string]interface{}, error) {
	return getDict(widgetDir + "/" + portalWidgetMetaFile)
}

func getPortalInternalResourceMetaFile(internalResourceDir string) (map[string]interface{}, error) {
	return getDict(internalResourceDir + "/" + portalInternalResourceMetaFile)
}

func getPortalInternalResourceCode(internalResourceDir, fileName string) (string, error) {
	return readFileAsString(internalResourceDir + "/" + fileName)
}
