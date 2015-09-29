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
	flag.Parse()
	if cblib.URL != "" {
		cb.CB_ADDR = cblib.URL
	} else {
		cblib.URL = cb.CB_ADDR
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
		fmt.Printf("MAJOR FAIL: %s\n", err.Error())
	}
}

func doAuth() (*cb.DevClient, error) {
	if cblib.Email != "" && cblib.Password != "" {
		return cblib.AuthUserPass(cblib.Email, cblib.Password)
	} else {
		return cblib.Auth("")
	}
}
