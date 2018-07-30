package cblib

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bgentry/speakeasy"
	cb "github.com/clearblade/Go-SDK"
	"os"
	"strings"
)

const (
	urlPrompt       = "Platform URL"
	msgurlPrompt    = "Messaging URL"
	systemKeyPrompt = "System Key"
	emailPrompt     = "Developer Email"
	passwordPrompt  = "Password: "
)

func init() {
	flag.StringVar(&URL, "platform-url", "", "Clearblade platform url for target system")
	flag.StringVar(&MsgURL, "messaging-url", "", "Clearblade messaging url for target system")
	flag.StringVar(&SystemKey, "system-key", "", "System key for target system")
	flag.StringVar(&Email, "email", "", "Developer email for login")
	flag.StringVar(&Password, "password", "", "Developer password")
}

func getOneItem(prompt string, isASecret bool) string {
	reader := bufio.NewReader(os.Stdin)
	if isASecret {
		pw, err := speakeasy.Ask("Developer password: ")
		fmt.Printf("\n")
		if err != nil {
			fmt.Printf("Error getting password: %s\n", err.Error())
			os.Exit(1)
		}
		thing := string(pw)
		return strings.TrimSpace(thing)
	}
	fmt.Printf("%s: ", prompt)
	thing, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading answer: %s\n", err.Error())
		os.Exit(1)
	}
	return strings.TrimSpace(thing)
}

func buildPrompt(basicPrompt, defaultValue string) string {
	if defaultValue == "" {
		return basicPrompt
	}
	return fmt.Sprintf("%s (%s)", basicPrompt, defaultValue)
}

func getAnswer(entered, defaultValue string) string {
	if entered != "" {
		return entered
	}
	return defaultValue
}

func fillInTheBlanks(defaults *DefaultInfo) {
	var defaultUrl, defaultMsgUrl, defaultEmail, defaultSys string
	if defaults != nil {
		defaultUrl, defaultMsgUrl, defaultEmail, defaultSys = defaults.url, defaults.msgUrl, defaults.email, defaults.systemKey
	}
	if URL == "" {
		URL = getAnswer(getOneItem(buildPrompt(urlPrompt, defaultUrl), false), defaultUrl)
		if MsgURL == "" {
			MsgURL = getAnswer(getOneItem(buildPrompt(msgurlPrompt, defaultMsgUrl), false), defaultMsgUrl)
		}
	}
	setupAddrs(URL, MsgURL)
	if SystemKey == "" {
		SystemKey = getAnswer(getOneItem(buildPrompt(systemKeyPrompt, defaultSys), false), defaultSys)
	}
	if Email == "" {
		Email = getAnswer(getOneItem(buildPrompt(emailPrompt, defaultEmail), false), defaultEmail)
	}
	if Password == "" {
		Password = getOneItem(passwordPrompt, true)
	}
}

func GoToRepoRootDir() error {
	var err error
	whereIReallyAm, _ := os.Getwd()
	for {
		dirname, dirErr := os.Getwd()
		if dirErr != nil {
			return dirErr
		}
		if dirname == "/" || strings.HasSuffix(dirname, ":\\") {
			os.Chdir(whereIReallyAm) //  go back in case this err is ignored
			return fmt.Errorf(SpecialNoCBMetaError)
		}
		if IsInRepo() {
			// Exit
			return nil
		} else if err = os.Chdir(".."); err != nil {
				return fmt.Errorf("Error changing directory: %s", err.Error())
		}
	}
}

func makeClientFromMetaInfo() *cb.DevClient {
	var newSchema bool
	devToken := MetaInfo["token"].(string)
	email, ok := MetaInfo["developerEmail"].(string)
	if !ok {
		email = MetaInfo["developer_email"].(string)
		newSchema = true
	}
	// Checking if meta has messagingURL attribute to support systems that were exported before
	// This code is horrible but needs to be done to maintain backward compatibility with
	// systems that are already exported
	if newSchema {
		messagingURL, ok := MetaInfo["messaging_url"].(string)
		if !ok {
			setupAddrs(MetaInfo["platform_url"].(string), "")
		} else {
			setupAddrs(MetaInfo["platform_url"].(string), messagingURL)
		}
	} else {
		messagingURL, ok := MetaInfo["messagingURL"].(string)
		if !ok {
			setupAddrs(MetaInfo["platformURL"].(string), "")
		} else {
			setupAddrs(MetaInfo["platformURL"].(string), messagingURL)
		}
	}

	return cb.NewDevClientWithToken(devToken, email)
}

func Authorize(defaults *DefaultInfo) (*cb.DevClient, error) {
	var ok bool
	if MetaInfo != nil {
		DevToken = MetaInfo["token"].(string)
		Email, ok = MetaInfo["developerEmail"].(string)
		if !ok {
			Email = MetaInfo["developer_email"].(string)
		}
		URL, ok = MetaInfo["platformURL"].(string)
		if !ok {
			URL = MetaInfo["platform_url"].(string)
		}
		MsgURL, ok = MetaInfo["messagingURL"].(string)
		if !ok {
			MsgURL = MetaInfo["messaging_url"].(string)
		}
		setupAddrs(URL, MsgURL)
		fmt.Printf("Using ClearBlade platform at '%s'\n", cb.CB_ADDR)
		fmt.Printf("Using ClearBlade messaging at '%s'\n", cb.CB_MSG_ADDR)
		return cb.NewDevClientWithToken(DevToken, Email), nil
	}
	// No cb meta file -- get url, syskey, email passwd
	fillInTheBlanks(defaults)
	fmt.Printf("Using ClearBlade platform at '%s'\n", cb.CB_ADDR)
	fmt.Printf("Using ClearBlade messaging at '%s'\n", cb.CB_MSG_ADDR)
	cli := cb.NewDevClient(Email, Password)
	if err := cli.Authenticate(); err != nil {
		fmt.Printf("Authenticate failed: %s\n", err)
		return nil, err
	}
	return cli, nil
}

func checkIfTokenHasExpired(client *cb.DevClient, systemKey string) (*cb.DevClient, error) {
	_, err := client.GetAllRoles(systemKey)
	if err != nil {
		fmt.Printf("Token has probably expired. Please enter details for authentication again...\n")
		MetaInfo = nil
		client, _ = Authorize(nil)
		metaStuff := map[string]interface{}{
			"platform_url":        cb.CB_ADDR,
			"messaging_url":       cb.CB_MSG_ADDR,
			"developer_email":     Email,
			"asset_refresh_dates": []interface{}{},
			"token":               client.DevToken,
		}
		if err = storeCBMeta(metaStuff); err != nil {
			return nil, err
		}
		return client, nil
	}
	return client, nil
}