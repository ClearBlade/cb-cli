package cblib

import (
	"flag"
	"fmt"
	"strings"

	cb "github.com/clearblade/Go-SDK"
)

type SubCommand struct {
	name            string
	usage           string
	needsAuth       bool
	mustBeInRepo    bool
	mustNotBeInRepo bool
	flags           flag.FlagSet
	run             func(cmd *SubCommand, client *cb.DevClient, args ...string) error
	example         string
}

var (
	subCommands = map[string]*SubCommand{}
)

func (c *SubCommand) Execute( /*client *cb.DevClient,*/ args []string) error {
	var client *cb.DevClient
	var err error
	c.flags.Parse(args)

	if URL != "" && MsgURL != "" {
		setupAddrs(URL, MsgURL)
	}
	if err = GoToRepoRootDir(); err != nil {
		if c.mustBeInRepo {
			return fmt.Errorf("You must be in an initialized repo to run the '%s' command\n", c.name)
		}
		if err.Error() != SpecialNoCBMetaError {
			return err
		}
	} else if c.mustNotBeInRepo {
		return fmt.Errorf("You cannot run the '%s' command in an existing ClearBlade repository", c.name)
	}

	SetRootDir(".")
	// This is the most important part of initialization
	MetaInfo, _ = getCbMeta()

	if MetaInfo != nil {
		client = makeClientFromMetaInfo()
		if c.needsAuth {
			err := client.CheckAuth()
			if err != nil {
				return fmt.Errorf("Check authentication step failed. Please make sure that your token is valid. Hint: if you're inside an existing system you might want to run the 'target' command to re-authenticate. Error - %s", err.Error())
			}
		}
	} else if c.needsAuth {
		client, err = Authorize(nil)
		if err != nil {
			return err
		}
	}
	RootDirIsSet = false
	return c.run(c, client, c.flags.Args()...)
}

func PrintHelpFor(c *SubCommand, args ...string) {
	fmt.Printf("Usage: %s\n", c.usage)
	c.flags.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println(c.example)
}

func GetCommand(commandName string) (*SubCommand, error) {
	commandName = strings.ToLower(commandName)
	if theSubCommand, ok := subCommands[commandName]; ok {
		return theSubCommand, nil
	}
	return nil, fmt.Errorf("Subcommand %s not found", commandName)
}

func AddCommand(commandName string, stuff *SubCommand) {
	commandName = strings.ToLower(commandName)
	if _, alreadyExists := subCommands[commandName]; alreadyExists {
		fmt.Printf("Trying to add command %s but it already exists, ignoring...\n", commandName)
		return
	}
	subCommands[commandName] = stuff
}

func PrintRootHelp() {
	var usage = `
The cb-cli (ClearBlade CLI) provides methods for interacting with ClearBlade platform

Usage: cb-cli <command>

Commands:

`
	for cmd, _ := range subCommands {
		usage += fmt.Sprintf("\t%v\n", cmd)
	}
	usage += `

Examples:

	cb-cli init 							# inits your local workspace
	cb-cli export							# inits and exports your systems
	cb-cli push -service=Service1			# pushs an individual service up to Platform
	cb-cli pull -collection=Collection2	# pulls an individual collection to filesystem
	 `
	fmt.Println(usage)
}
