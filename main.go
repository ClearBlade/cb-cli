package main

import (
	"flag"
	"fmt"
	cb "github.com/clearblade/Go-SDK"
	"github.com/clearblade/cblib"
	"os"
	"strings"
)

func init() {
	flag.StringVar(&cblib.URL, "url", "", "Platform URL")
	flag.BoolVar(&cblib.Help, "help", false, "Print help message")
	flag.StringVar(&cblib.Email, "email", "", "Developer username")
	flag.StringVar(&cblib.Password, "password", "", "Developer password")
}

func main() {
	if err := setupFromOptsAndCBMeta(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	client, err := doAuth()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	cmd := strings.ToLower(flag.Arg(0))
	subCommand, err := cblib.GetCommand(cmd)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	subArgs := flag.Args()
	err = subCommand.Execute(client, subArgs[1:])
	if err != nil {
		fmt.Printf("Aborting: %s\n", err.Error())
	}
}

func setupFromOptsAndCBMeta() error {
	flag.Parse()
	//cblib.DevToken, _ = cblib.Load_auth_info(cblib.AuthInfoFile)
	err := cblib.GoToRepoRootDir()
	if err != nil && err.Error() != cblib.SpecialNoCBMetaError {
		return err
	}
	if cblib.URL != "" {
		cb.CB_ADDR = cblib.URL
	} else if url, ok := cblib.MetaInfo["platformURL"]; ok {
		cblib.URL = url.(string)
		cb.CB_ADDR = url.(string)
	} else {
		cblib.URL = cb.CB_ADDR
	}

	if cblib.DevToken == "" {
		if tok, ok := cblib.MetaInfo["token"]; ok {
			cblib.DevToken = tok.(string)
		}
	}

	fmt.Printf("Using system at '%s'\n", cb.CB_ADDR)

	if cblib.Email == "" {
		if email, ok := cblib.MetaInfo["developerEmail"]; ok {
			cblib.Email = email.(string)
		}
	} else {
		cblib.CommandLineEmail = true
	}
	return nil
}

func doAuth() (*cb.DevClient, error) {
	if cblib.Email != "" && cblib.Password != "" {
		return cblib.AuthUserPass(cblib.Email, cblib.Password)
	} else if cblib.DevToken != "" {
		return cblib.Auth(cblib.DevToken)
	} else if cblib.CommandLineEmail {
		return cblib.AuthPromptPass(cblib.Email)
	} else {
		return cblib.Auth("")
	}
}
