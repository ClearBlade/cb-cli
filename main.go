package main

import (
	"flag"
	"fmt"
	cb "github.com/clearblade/Go-SDK"
	"github.com/clearblade/cblib"
	"strings"
)

func init() {
	flag.StringVar(&cblib.URL, "url", "", "Set the URL of the platform you want to use")
	flag.BoolVar(&cblib.ShouldImportCollectionRows, "importrows", false, "If supplied the import command will transfer collection rows from the old system to the new system")
	flag.IntVar(&cblib.ImportPageSize, "pagesize", 100, "If supplied the import command will migrate the specified number of rows per request")
}

func main() {
	flag.Parse()
	if cblib.URL != "" {
		cb.CB_ADDR = cblib.URL
	} else {
		cblib.URL = cb.CB_ADDR
	}
	cmd := strings.ToLower(flag.Arg(0))
	var err error
	var sysKey, dir string

	switch cmd {
	case "auth":
		if err := cblib.Auth_cmd(); err != nil {
			fmt.Printf("Error authenticated: %v\n", err)
		}
	case "pull":
		if flag.NArg() != 2 {
			fmt.Printf("pull requires the systemKey as an argument\n")
		}
		if err := cblib.Pull_cmd(flag.Arg(1)); err != nil {
			fmt.Printf("Error pulling data: %v\n", err)
		}
	case "push":
		if flag.NArg() != 2 {
			sysKey, err = cblib.Sys_for_dir()
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			dir = "."
		} else {
			dir = ""
			sysKey = flag.Arg(1)
		}
		if err := cblib.Push_cmd(sysKey, dir); err != nil {
			fmt.Printf("Error pushing: %v\n", err)
		}
	case "export":
		if flag.NArg() != 2 {
			fmt.Printf("export requires the systemKey as an argument\n")
		}
		if err := cblib.Export_cmd(flag.Arg(1), flag.Arg(2)); err != nil {
			fmt.Printf("Error export data: %v\n", err)
		}
	case "import":
		var sysKey, dir string
		var err error
		sysKey, err = cblib.Sys_for_dir()
		if err != nil {
			fmt.Printf("%v\n", err)
			fmt.Printf("%v\n", sysKey)
			return
		}
		dir = "."

		if err := cblib.Import_cmd(dir, flag.Arg(2)); err != nil {
			fmt.Printf("Error import data: %v\n", err)
		}
	default:
		fmt.Printf("Commands: 'auth', 'pull', 'push', 'export', 'import'\n")
	}
}
