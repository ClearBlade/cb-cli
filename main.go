package main

import (
	"fmt"
	"github.com/clearblade/cblib"
	"os"
)

func main() {
	theArgs := os.Args
	if len(theArgs) < 2 {
		fmt.Printf("No command provided\n")
		os.Exit(1)
	}
	// Special case version command for cb-cli only
	if theArgs[1] == "version" {
		fmt.Printf("%s\n", cbCliVersion)
		os.Exit(0)
	}

	subCommand, err := cblib.GetCommand(theArgs[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = subCommand.Execute( /*client,*/ theArgs[2:])
	if err != nil {
		fmt.Printf("Aborting: %s\n", err.Error())
	}
}
