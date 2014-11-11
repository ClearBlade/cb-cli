package main

import (
	"bufio"
	"code.google.com/p/gopass"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	cb "github.com/clearblade/Go-SDK"
	"io/ioutil"
	"os"
	"strings"
)

var (
	URL          string
	AuthInfoFile string
)

type Service_meta struct {
	Name    string
	Version int
	Hash    string
}

type System_meta struct {
	Name        string
	Key         string
	Description string
	Services    map[string]Service_meta
}

func init() {
	flag.StringVar(&URL, "url", "", "Set the URL of the platform you want to use")
	flag.StringVar(&AuthInfoFile, "authinfo", "", "File in which you wish to store auth info")
}

func auth() (*cb.DevClient, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your email: ")
	email, _ := reader.ReadString('\n')
	pass, pass_err := gopass.GetPass("Enter your password: ")
	if pass_err != nil {
		return nil, pass_err
	}
	cli := cb.NewDevClient(email, pass)
	if err := cli.Authenticate(); err != nil {
		return nil, err
	} else {
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

func pull_services(systemKey string, cli *cb.DevClient) ([]*cb.Service, error) {
	svcs, err := cli.GetServiceNames(systemKey)
	if err != nil {
		return nil, err
	}
	services := make([]*cb.Service, len(svcs))
	for i, svc := range svcs {
		service, err := cli.GetService(systemKey, svc)
		if err != nil {
			return nil, err
		}
		service.Code = strings.Replace(service.Code, "\\n", "\n", -1)
		services[i] = service
	}
	return services, nil
}

func pull_system_meta(systemKey string, cli *cb.DevClient) (*System_meta, error) {
	sys, err := cli.GetSystem(systemKey)
	if err != nil {
		return nil, err
	}
	svcs, err := pull_services(systemKey, cli)
	if err != nil {
		return nil, err
	}
	serv_metas := make(map[string]Service_meta)
	for _, svc := range svcs {
		serv_metas[svc.Name] = Service_meta{
			Name:    svc.Name,
			Version: svc.Version,
			Hash:    fmt.Sprintf("%x", md5.Sum([]byte(svc.Code))),
		}
	}
	sys_meta := &System_meta{
		Name:        sys.Name,
		Key:         sys.Key,
		Description: sys.Description,
		Services:    serv_metas,
	}
	return sys_meta, nil
}

func store_services(systemKey string, services []*cb.Service, meta *System_meta) error {
	if err := os.MkdirAll(systemKey, 0777); err != nil {
		return err
	}
	meta_bytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(systemKey+"/.meta.json", meta_bytes, 0777); err != nil {
		return err
	}

	for _, service := range services {
		if err := ioutil.WriteFile(systemKey+"/"+service.Name+".js", []byte(service.Code), 0777); err != nil {
			return err
		}
	}
	return nil
}

func store_meta(systemKey string, meta *System_meta) error {
	meta_bytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(systemKey+"/.meta.json", meta_bytes, 0777); err != nil {
		return err
	}
	return nil
}

func load_sys_meta(systemKey string) (*System_meta, error) {
	meta_bytes, err := ioutil.ReadFile(systemKey + "/.meta.json")
	if err != nil {
		return nil, err
	}
	sys_meta := new(System_meta)
	if err := json.Unmarshal(meta_bytes, sys_meta); err != nil {
		return nil, err
	}
	return sys_meta, nil
}

func service_hash(systemKey, name string) (string, error) {
	svc_bytes, err := ioutil.ReadFile(systemKey + "/" + name + ".js")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(svc_bytes)), nil
}

func service_changed(systemKey, name string) (bool, error) {
	sys_meta, err := load_sys_meta(systemKey)
	if err != nil {
		return false, err
	}
	hash, err := service_hash(systemKey, name)
	if err != nil {
		return false, err
	}
	if sys_meta.Services[name].Hash != hash {
		return true, nil
	} else {
		return false, nil
	}
}

func service_local(systemKey, name string) bool {
	if _, err := os.Stat(systemKey + "/" + name + ".js"); err != nil {
		return false
	} else {
		return true
	}
}

func ok_to_pull(systemKey string, cli *cb.DevClient) (bool, string, error) {
	sys_meta, err := pull_system_meta(systemKey, cli)
	if err != nil {
		return false, "", err
	}
	for k, _ := range sys_meta.Services {
		if service_local(systemKey, k) {
			if has_changed, err := service_changed(systemKey, k); err != nil {
				return false, "", err
			} else if has_changed {
				return false, "You have made changes to a service since the last pull", nil
			} else {
				continue
			}
		}
	}
	return true, "", nil
}

func system_diff(systemKey string, cli *cb.DevClient) ([]string, error) {
	sys_meta, err := pull_system_meta(systemKey, cli)
	if err != nil {
		return nil, err
	}
	changed := make([]string, 0)
	for k, _ := range sys_meta.Services {
		if has_changed, err := service_changed(systemKey, k); err != nil {
			return nil, err
		} else if has_changed {
			changed = append(changed, k)
		}
	}
	return changed, nil
}

func prompt(msg string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", msg)
	response, _ := reader.ReadString('\n')
	return response
}

func pull_cmd(sysKey string) error {
	var cli *cb.DevClient
	var err error
	if AuthInfoFile == "" {
		cli, err = auth()
		if err != nil {
			return err
		}
	} else {
		tok, load_err := load_auth_info(AuthInfoFile)
		if load_err != nil {
			return load_err
		}
		cli = &cb.DevClient{
			DevToken: tok,
		}
	}
	if ok, msg, err := ok_to_pull(sysKey, cli); !ok && err != nil {
		return err
	} else if !ok {
		resp := strings.TrimSpace(strings.ToLower(prompt(msg)))
		switch resp {
		case "yes", "y", "ye":

		case "no", "n":
			return fmt.Errorf("Did not pull because of unresolved changes")
		default:
			return fmt.Errorf("Invalid response: '%s', looking for 'yes' or 'no'", resp)
		}
	}

	sys_meta, err := pull_system_meta(sysKey, cli)
	if err != nil {
		return err
	}
	services, svcErr := pull_services(sysKey, cli)
	if svcErr != nil {
		return svcErr
	}
	if err := store_services(sysKey, services, sys_meta); err != nil {
		return err
	}
	fmt.Printf("Code for %s has been successfully pulled and put in a directory %s\n", sysKey, sysKey)
	return nil
}

func push(systemKey string, services []string, cli *cb.DevClient) error {
	for _, svc := range services {
		svc_bytes, err := ioutil.ReadFile(systemKey + "/" + svc + ".js")
		if err != nil {
			return err
		}
		if put_err := cli.UpdateService(systemKey, svc, string(svc_bytes)); put_err != nil {
			return put_err
		}
	}
	return nil
}

func push_cmd(systemKey string) error {
	var cli *cb.DevClient
	var err error
	if AuthInfoFile == "" {
		cli, err = auth()
		if err != nil {
			return err
		}
	} else {
		tok, load_err := load_auth_info(AuthInfoFile)
		if load_err != nil {
			return load_err
		}
		cli = &cb.DevClient{
			DevToken: tok,
		}
	}

	if svcs, err := system_diff(systemKey, cli); err != nil {
		return err
	} else if len(svcs) == 0 {
		return fmt.Errorf("No services have changed, nothing to push")
	} else {
		if err := push(systemKey, svcs, cli); err != nil {
			return err
		} else {
			meta, meta_err := pull_system_meta(systemKey, cli)
			if meta_err != nil {
				return meta_err
			}
			if store_err := store_meta(systemKey, meta); store_err != nil {
				return store_err
			}
			fmt.Printf("Push successful\n")
			return nil
		}
	}
}

func auth_cmd() error {
	cli, auth_err := auth()
	if auth_err != nil {
		return auth_err
	}
	if AuthInfoFile == "" {
		fmt.Printf("UserToken: %s\n", cli.DevToken)
		return nil
	}
	return save_auth_info(AuthInfoFile, cli.DevToken)
}

func main() {
	flag.Parse()

	if URL != "" {
		cb.CB_ADDR = URL
	}
	cmd := strings.ToLower(flag.Arg(0))
	switch cmd {
	case "auth":
		if err := auth_cmd(); err != nil {
			fmt.Printf("Error authenticated: %v\n", err)
		}
	case "pull":
		if flag.NArg() != 2 {
			fmt.Printf("pull requires the systemKey as an argument\n")
		}
		if err := pull_cmd(flag.Arg(1)); err != nil {
			fmt.Printf("Error pulling data: %v\n", err)
		}
	case "push":
		if flag.NArg() != 2 {
			fmt.Printf("push requires the systemKey as an argument\n")
		}
		if err := push_cmd(flag.Arg(1)); err != nil {
			fmt.Printf("Error pushing: %v\n", err)
		}
	default:
		fmt.Printf("Commands: 'auth', 'pull', 'push'\n")
	}
}
