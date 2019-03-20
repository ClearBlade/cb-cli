package cblib

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/totherme/unstructured"
)

func processDataSourceDir(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	return nil
}

func compressDatasources(portal *unstructured.Data, decompressedPortalDir string) error {
	portalConfig, err := portal.GetByPointer(portalConfigPath)
	if err != nil {
		return err
	}
	if err := portalConfig.SetField("datasources", map[string]interface{}{}); err != nil {
		return err
	}
	myPayloadData, err := portal.GetByPointer(portalDatasourcesPath)

	if err != nil {
		return fmt.Errorf("Couldn't address datasources into my own json")
	}

	datasourcesDir := filepath.Join(decompressedPortalDir, "datasources")
	filepath.Walk(datasourcesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		currDS, err := readFileAsString(path)
		if err != nil {
			return err
		}
		currDSObj, err := unstructured.ParseJSON(currDS)
		if err != nil {
			return err
		}
		dsID, err := currDSObj.GetByPointer("/id")
		if err != nil {
			return err
		}
		stringifiedDsID, err := dsID.StringValue()
		if err != nil {
			return err
		}
		updatedDsData, err := currDSObj.ObValue()
		if err != nil {
			return err
		}
		err = myPayloadData.SetField(stringifiedDsID, updatedDsData)

		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

func recursivelyFindKeyPath(queryKey string, data map[string]interface{}, keysToIgnoreInData map[string]interface{}, keyPath string) string {
	for k, v := range data {
		if k == queryKey {
			return keyPath
		}
		switch v.(type) {
		case map[string]interface{}:
			if keysToIgnoreInData[k] != nil {
				continue
			}
			updatedKeyPath := keyPath + k + "/"
			val := recursivelyFindKeyPath(queryKey, v.(map[string]interface{}), keysToIgnoreInData, updatedKeyPath)
			if val != "" {
				return val
			}
		default:
			continue
		}
	}
	return ""
}

func updateObjUsingWebFiles(webData *unstructured.Data, currDir string) error {
	htmlFile := filepath.Join(currDir, outFile+".html")
	updateObjFromFile(webData, htmlFile, htmlKey)

	javascriptFile := filepath.Join(currDir, outFile+".js")
	updateObjFromFile(webData, javascriptFile, javascriptKey)

	cssFile := filepath.Join(currDir, outFile+".css")
	updateObjFromFile(webData, cssFile, cssKey)
	return nil
}

func updateObjFromFile(data *unstructured.Data, currFile string, fieldToSet string) error {
	s, err := readFileAsString(currFile)
	if err != nil {
		log.Println("Update obj from file error:", err)
		return err
	}
	data.SetField(fieldToSet, s)
	return nil
}

func processParser(currWidgetDir string, parserObj *unstructured.Data, parserType string) error {
	valueData, err := parserObj.GetByPointer("/value")
	if err != nil {
		return err
	}

	switch valueData.RawValue().(type) {
	case map[string]interface{}:
		currDir := filepath.Join(currWidgetDir, parserType)
		updateObjUsingWebFiles(&valueData, currDir)
	case string:
		currFile := filepath.Join(currWidgetDir, parserType, outFile+".js")
		updateObjFromFile(parserObj, currFile, "value")
	default:

	}
	return nil

}

func mergeMaps(a map[string]interface{}, b map[string]interface{}) {
	for k, v := range b {
		a[k] = v
	}
}

func processCurrInternalResourceDir(path string, allInternalResources *unstructured.Data) error {
	meta, err := getPortalInternalResourceMetaFile(path)
	if err != nil {
		return err
	}

	resourceID := meta["id"].(string)
	if err := allInternalResources.SetField(resourceID, meta); err != nil {
		return err
	}

	myResource, err := allInternalResources.GetByPointer("/" + resourceID)
	if err != nil {
		return err
	}

	resourceName := meta["name"].(string)
	file, err := getPortalInternalResourceCode(path, resourceName)
	if err != nil {
		return err
	}

	return myResource.SetField("file", file)
}

func processCurrWidgetDir(path string, allWidgets *unstructured.Data) error {

	widgetMeta, err := getPortalWidgetMetaFile(path)
	if err != nil {
		return err
	}

	widgetSettings, err := getPortalWidgetSettingsFile(path)
	if err != nil {
		return err
	}

	mergeMaps(widgetMeta, map[string]interface{}{"props": widgetSettings})
	widgetID := widgetMeta["id"].(string)
	if err := allWidgets.SetField(widgetID, widgetMeta); err != nil {
		return err
	}

	myWidget, err := allWidgets.GetByPointer("/" + widgetID)
	if err != nil {
		return err
	}
	return actOnParserSettings(widgetSettings, func(settingName, dataType string) error {
		settingDir := path + "/" + parsersDirectory + "/" + settingName

		if setting, err := myWidget.GetByPointer("/props/" + settingName); err == nil {
			found := false
			if incoming, err := setting.GetByPointer("/" + incomingParserKey); err == nil {
				found = true
				if dataType != dynamicDataType {
					incoming = setting
				}
				if err := processParser(settingDir, &incoming, incomingParserKey); err != nil {
					return err
				}
			}

			if outgoing, err := setting.GetByPointer("/" + outgoingParserKey); err == nil {
				found = true
				if dataType != dynamicDataType {
					outgoing = setting
				}
				if err := processParser(settingDir, &outgoing, outgoingParserKey); err != nil {
					return err
				}
			}

			if !found {
				if setting.HasKey("value") {
					if err := processParser(settingDir, &setting, incomingParserKey); err != nil {
						return err
					}
				}
			}
		} else {
			return err
		}

		return nil
	})
}

func compressWidgets(portal *unstructured.Data, decompressedPortalDir string) error {
	portalConfig, err := portal.GetByPointer(portalConfigPath)
	if err != nil {
		return err
	}
	portalConfig.SetField("widgets", map[string]interface{}{})
	widgets, err := portal.GetByPointer(portalWidgetsPath)
	if err != nil {
		return fmt.Errorf("Couldn't address widgets into my own json")
	}

	widgetsDir := filepath.Join(decompressedPortalDir, "widgets")
	return filepath.Walk(widgetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		split := strings.Split(path, "/")
		if split[len(split)-2] != widgetsDirectory {
			return nil
		}

		return processCurrWidgetDir(path, &widgets)
	})
}

func getDecompressedPortalDir(portalName string) string {
	return filepath.Join(portalsDir, portalName, portalConfigDirectory)
}

func compressInternalResources(portal *unstructured.Data, decompressedPortalDir string) error {
	portalConfig, err := portal.GetByPointer(portalConfigPath)
	if err != nil {
		return err
	}
	portalConfig.SetField("internalResources", map[string]interface{}{})
	resources, err := portal.GetByPointer(portalInternalResourcesPath)
	if err != nil {
		return fmt.Errorf("Couldn't address internal resources into my own json")
	}

	internalResourcesDir := filepath.Join(decompressedPortalDir, "internalResources")
	return filepath.Walk(internalResourcesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		split := strings.Split(path, "/")
		if split[len(split)-2] != internalResourcesDirectory {
			return nil
		}

		return processCurrInternalResourceDir(path, &resources)
	})
}

func compressPortal(name string) (map[string]interface{}, error) {

	decompressedPortalDir := getDecompressedPortalDir(name)

	p, err := getPortal(name)
	if err != nil {
		return nil, err
	}
	portalConfig, err := convertPortalMapToUnstructured(p)
	if err != nil {
		return nil, err
	}

	if err := compressDatasources(portalConfig, decompressedPortalDir); err != nil {
		return nil, err
	}
	if err := compressWidgets(portalConfig, decompressedPortalDir); err != nil {
		return nil, err
	}
	if err := compressInternalResources(portalConfig, decompressedPortalDir); err != nil {
		return nil, err
	}

	return portalConfig.ObValue()
}
