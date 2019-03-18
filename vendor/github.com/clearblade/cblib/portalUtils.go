package cblib

import "fmt"

func actOnParserSettings(widgetConfig map[string]interface{}, cb func(string, string, map[string]interface{}) error) error {
	widgetSettings := make(map[string]interface{})
	ok := true
	if widgetSettings, ok = widgetConfig["props"].(map[string]interface{}); !ok {
		return fmt.Errorf("No props key for widget config")
	}
	for settingName, v := range widgetSettings {
		switch v.(type) {
		case map[string]interface{}:
			// if there's a dataType property we know this setting is a parser
			if dataType, ok := v.(map[string]interface{})["dataType"].(string); ok {
				if err := cb(settingName, dataType, v.(map[string]interface{})); err != nil {
					return err
				}
			}
		default:
			continue
		}
	}
	return nil
}
