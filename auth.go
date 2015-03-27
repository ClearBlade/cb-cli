package CBCLI

import (
	"bufio"
	"code.google.com/p/gopass"
	"flag"
	"fmt"
	cb "github.com/clearblade/Go-SDK"
	"io/ioutil"
	"os"
	"os/user"
)

var (
	AuthInfoFile string
)

func init() {
	flag.StringVar(&AuthInfoFile, "authinfo", homedir()+"/.cbauth", "File in which you wish to store auth info")
}

func homedir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

func auth_prompt() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your email: ")
	email, _ := reader.ReadString('\n')
	pass, pass_err := gopass.GetPass("Enter your password: ")
	if pass_err != nil {
		return "", "", pass_err
	}
	return email, pass, nil
}

func auth(devToken string) (*cb.DevClient, error) {
	var cli *cb.DevClient
	if devToken != "" {
		cli = &cb.DevClient{
			DevToken: devToken,
		}
		return cli, nil
	}
	if _, err := os.Stat(AuthInfoFile); os.IsNotExist(err) {
		email, pass, prompt_err := auth_prompt()
		if prompt_err != nil {
			return nil, prompt_err
		}
		cli = cb.NewDevClient(email, pass)

		if err := cli.Authenticate(); err != nil {
			return nil, err
		} else {
			return cli, save_auth_info(AuthInfoFile, cli.DevToken)
		}
	} else {
		token, err := load_auth_info(AuthInfoFile)
		if err != nil {
			return nil, err
		}
		cli = &cb.DevClient{
			DevToken: token,
		}
		return cli, nil
	}
}

func save_auth_info(filename, token string) error {
	return ioutil.WriteFile(filename, []byte(token), 0600)
}

func load_auth_info(filename string) (string, error) {
	if data, err := ioutil.ReadFile(filename); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}
