package models

import (
	"encoding/base64"
	"fmt"
	cb "github.com/clearblade/Go-SDK"
)

type Adaptor struct {
	Name             string
	Info             map[string]interface{}
	InfoForFiles     []interface{}
	ContentsForFiles []map[string]interface{}
	client           *cb.DevClient
	systemKey        string
}

func InitializeAdaptor(name, systemKey string, client *cb.DevClient) *Adaptor {
	return &Adaptor{
		Name:      name,
		client:    client,
		systemKey: systemKey,
	}
}

func (a *Adaptor) FetchAllInfo() error {
	adaptorMeta, err := a.client.GetAdaptor(a.systemKey, a.Name)
	if err != nil {
		return err
	}
	a.Info = adaptorMeta

	adaptorFiles, err := a.client.GetAdaptorFiles(a.systemKey, a.Name)
	if err != nil {
		return err
	}
	a.InfoForFiles = adaptorFiles

	a.ContentsForFiles = make([]map[string]interface{}, 0)
	for i := 0; i < len(a.InfoForFiles); i++ {
		currentAdaptorFile := a.InfoForFiles[i].(map[string]interface{})
		content, err := a.client.GetAdaptorFile(a.systemKey, currentAdaptorFile["adaptor_name"].(string), currentAdaptorFile["name"].(string))
		if err != nil {
			return err
		}
		a.ContentsForFiles = append(a.ContentsForFiles, content)
	}
	return nil
}

func (a *Adaptor) UploadAllInfo() error {
	delete(a.Info, "version")
	if _, err := a.client.CreateAdaptor(a.systemKey, a.Name, a.Info); err != nil {
		return err
	}
	for i := 0; i < len(a.ContentsForFiles); i++ {
		currentFile := a.ContentsForFiles[i]
		currentFileName := currentFile["name"].(string)
		delete(currentFile, "version")
		if _, err := a.client.CreateAdaptorFile(a.systemKey, a.Name, currentFileName, currentFile); err != nil {
			return err
		}
	}

	return nil
}

func (a *Adaptor) UpdateAllInfo() error {
	delete(a.Info, "version")
	delete(a.Info, "name")
	if _, err := a.client.UpdateAdaptor(a.systemKey, a.Name, a.Info); err != nil {
		return err
	}
	for i := 0; i < len(a.ContentsForFiles); i++ {
		currentFile := a.ContentsForFiles[i]
		currentFileName := currentFile["name"].(string)
		delete(currentFile, "version")
		delete(currentFile, "adaptor_name")
		if _, err := a.client.GetAdaptorFile(a.systemKey, a.Name, currentFileName); err != nil {
			// adaptor file DNE
			if _, err := a.client.CreateAdaptorFile(a.systemKey, a.Name, currentFileName, currentFile); err != nil {
				return err
			}
		} else {
			if _, err := a.client.UpdateAdaptorFile(a.systemKey, a.Name, currentFileName, currentFile); err != nil {
				return err
			}
		}

	}

	return nil
}

func (a *Adaptor) DecodeFileByName(fileName string) ([]byte, error) {
	for i := 0; i < len(a.ContentsForFiles); i++ {
		if a.ContentsForFiles[i]["name"] == fileName {
			decoded, err := a.DecodeFile(a.ContentsForFiles[i]["file"].(string))
			if err != nil {
				fmt.Printf("Unable to decode file for '%s'", fileName)
				return nil, err
			}
			return decoded, nil
		}
	}
	return nil, fmt.Errorf("No adaptor file with name '%s'", fileName)
}

func (a *Adaptor) EncodeFile(contents []byte) string {
	return base64.StdEncoding.EncodeToString(contents)
}

func (a *Adaptor) DecodeFile(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
