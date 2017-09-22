package main

import (
	"fmt"
	"os"

	"strings"

	"github.com/clearblade/cblib"
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

	fmt.Printf("%v\n", theArgs[2:])

	var newArgs []string

	for _, value := range theArgs[2:] {
		//if the first character of the value is a dash, we are at a new flag
		if len(value) > 0 && string(value[0]) == "-" {
			newArgs = append(newArgs, strings.TrimSpace(value))
		} else {
			//append the value to the previous element
			newArgs[len(newArgs)-1] += strings.TrimSpace(value)
		}
	}

	err = subCommand.Execute( /*client,*/ newArgs)
	if err != nil {
		fmt.Printf("Aborting: %s\n", err.Error())
	}
}
