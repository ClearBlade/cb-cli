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