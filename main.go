package main

import (
	"bufio"
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
	URL                        string
	shouldImportCollectionRows bool
)

type Role_meta struct {
	Name        string
	Description string
	Permission  []map[string]interface{}
}

type Column struct {
	ColumnName string
	ColumnType string
}

type Collection_meta struct {
	Name          string
	Collection_id string
	Columns       []Column
}

type Service_meta struct {
	Name    string
	Version int
	Hash    string
	Params  []string
}

type System_meta struct {
	Name        string
	Key         string
	Description string
	Services    map[string]Service_meta
	PlatformUrl string
}

func init() {
	flag.StringVar(&URL, "url", "", "Set the URL of the platform you want to use")
	flag.BoolVar(&shouldImportCollectionRows, "importrows", false, "If supplied the import command will transfer collection rows from the old system to the new system")
}

func pull_roles(systemKey string, cli *cb.DevClient) ([]interface{}, error) {
	r, err := cli.GetAllRoles(systemKey)
	if err != nil {
		return nil, err
	}

	return r, nil
}
func store_roles(roles []interface{}, meta *System_meta) error {

	authdir := strings.Replace(meta.Name+"/auth", " ", "_", -1)
	if err := os.MkdirAll(authdir, 0777); err != nil {
		return err
	}
	meta_bytes, err := json.Marshal(roles)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(authdir+"/roles.json", meta_bytes, 0777); err != nil {
		return err
	}

	return nil
}

func pull_colls(systemKey string, cli *cb.DevClient) ([]Collection_meta, error) {
	colls, err := cli.GetAllCollections(systemKey)
	if err != nil {
		return nil, err
	}
	collections := make([]Collection_meta, len(colls))
	for i, col := range colls {
		co := col.(map[string]interface{})
		id, ok := co["collectionID"].(string)
		if !ok {
			return nil, fmt.Errorf("collectionID is not a string")
		}
		columnsResp, err := cli.GetColumns(id)

		if err != nil {
			return nil, err
		}
		columns := make([]Column, len(columnsResp))
		for j := 0; j < len(columnsResp); j++ {
			columns[j] = Column{
				ColumnName: columnsResp[j].(map[string]interface{})["ColumnName"].(string),
				ColumnType: columnsResp[j].(map[string]interface{})["ColumnType"].(string),
			}
		}

		collections[i] = Collection_meta{
			Name:          co["name"].(string),
			Collection_id: co["collectionID"].(string),
			Columns:       columns,
		}
	}

	return collections, nil
}

func store_cols(collections []Collection_meta, meta *System_meta) error {

	datadir := strings.Replace(meta.Name+"/data", " ", "_", -1)
	if err := os.MkdirAll(datadir, 0777); err != nil {
		return err
	}
	meta_bytes, err := json.Marshal(collections)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(datadir+"/collections.json", meta_bytes, 0777); err != nil {
		return err
	}

	return nil
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
			Params:  svc.Params,
		}
	}
	sys_meta := &System_meta{
		Name:        sys.Name,
		Key:         sys.Key,
		Description: sys.Description,
		Services:    serv_metas,
		PlatformUrl: URL,
	}
	return sys_meta, nil
}

func createSystemDirectory(dir string, meta *System_meta) error {
	if err := os.MkdirAll(dir, 0777); err != nil {
		fmt.Println("wtf")
		fmt.Println("dir= ", dir)
		return err
	}
	return nil
}

func store_services(systemKey string, services []*cb.Service, meta *System_meta) error {
	dir := strings.Replace(meta.Name, " ", "_", -1)
	err := createSystemDirectory(dir, meta)
	if err != nil {
		return err
	}
	codedir := strings.Replace(meta.Name+"/code", " ", "_", -1)
	if err := os.MkdirAll(codedir, 0777); err != nil {
		return err
	}
	meta_bytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dir+"/.meta.json", meta_bytes, 0777); err != nil {
		return err
	}

	for _, service := range services {
		if err := ioutil.WriteFile(codedir+"/"+service.Name+".js", []byte(service.Code), 0777); err != nil {
			return err
		}
	}
	return nil
}

// func store_system(systemKey string, meta *System_meta) error {
// 	dir := strings.Replace(meta.Name, " ", "_", -1)
// 	if err := os.MkdirAll(dir, 0777); err != nil {
// 		return err
// 	}
// 	// meta.Services = nil

// 	meta_bytes, err := json.Marshal(meta)
// 	if err != nil {
// 		return err
// 	}
// 	if err := ioutil.WriteFile(dir+"/system.json", meta_bytes, 0777); err != nil {
// 		return err
// 	}
// 	return nil
// }

func store_meta(dir string, meta *System_meta) error {
	meta.PlatformUrl = URL
	meta_bytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dir+"/.meta.json", meta_bytes, 0777); err != nil {
		fmt.Println("fuckkk")
		return err
	}
	return nil
}

func load_sys_meta(dir string) (*System_meta, error) {
	meta_bytes, err := ioutil.ReadFile(dir + "/.meta.json")
	if err != nil {
		return nil, err
	}
	sys_meta := new(System_meta)
	if err := json.Unmarshal(meta_bytes, sys_meta); err != nil {
		return nil, err
	}
	return sys_meta, nil
}

func load_collection_meta(dir string) ([]Collection_meta, error) {
	meta_bytes, err := ioutil.ReadFile(dir + "/data/collections.json")
	if err != nil {
		return nil, err
	}
	var collection_meta []Collection_meta
	if err := json.Unmarshal(meta_bytes, &collection_meta); err != nil {
		return nil, err
	}
	return collection_meta, nil
}

func load_services(dir string, meta *System_meta) ([]cb.Service, error) {
	service_meta := make([]cb.Service, len(meta.Services))

	i := 0
	for k, _ := range meta.Services {
		code_bytes, err := ioutil.ReadFile(dir + "/code/" + k + ".js")
		if err != nil {
			return nil, err
		}
		service_meta[i].Name = meta.Services[k].Name
		service_meta[i].Params = meta.Services[k].Params
		service_meta[i].Code = string(code_bytes)
		i++
	}

	return service_meta, nil
}

func load_roles(dir string) ([]interface{}, error) {
	roles_bytes, err := ioutil.ReadFile(dir + "/auth/roles.json")
	if err != nil {
		return nil, err
	}
	var roles_data []interface{}
	if err := json.Unmarshal(roles_bytes, &roles_data); err != nil {
		return nil, err
	}
	return roles_data, nil
}

func service_hash(dir, name string) (string, error) {
	svc_bytes, err := ioutil.ReadFile(dir + "/" + name + ".js")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(svc_bytes)), nil
}

func service_changed(dir, name string) (bool, error) {
	sys_meta, err := load_sys_meta(dir)
	if err != nil {
		return false, err
	}
	hash, err := service_hash(dir, name)
	if err != nil {
		return false, err
	}
	if sys_meta.Services[name].Hash != hash {
		return true, nil
	} else {
		return false, nil
	}
}

func service_local(dir, name string) bool {
	if _, err := os.Stat(dir + "/" + name + ".js"); err != nil {
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
		dir := strings.Replace(sys_meta.Name, " ", "_", -1)
		if service_local(dir, k) {
			if has_changed, err := service_changed(dir, k); err != nil {
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

func system_diff(systemKey, dir string, cli *cb.DevClient) ([]string, error) {
	sys_meta, err := pull_system_meta(systemKey, cli)
	if err != nil {
		return nil, err
	}
	changed := make([]string, 0)
	for k, _ := range sys_meta.Services {
		if has_changed, err := service_changed(dir, k); err != nil {
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
	cli, err := auth()
	if err != nil {
		return err
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
	dir := strings.Replace(sys_meta.Name, " ", "_", -1)
	fmt.Printf("Code for %s has been successfully pulled and put in a directory %s\n", sysKey, dir)
	return nil
}

func service_params(dir, name string) ([]string, error) {
	meta, err := load_sys_meta(dir)
	if err != nil {
		return nil, err
	}
	if svc_meta, extant := meta.Services[name]; extant {
		return svc_meta.Params, nil
	} else {
		return nil, fmt.Errorf("service does not exist locally")
	}
}

func push(systemKey, dir string, services []string, cli *cb.DevClient) error {
	for _, svc := range services {

		svc_bytes, err := ioutil.ReadFile(dir + "/code/" + svc + ".js")
		if err != nil {
			return err
		}
		svc_bytes = []byte(strings.Replace(strings.Replace(string(svc_bytes), "\\n", "\n", -1), "\n", "\\n", -1))
		params, err := service_params(dir, svc)
		if err != nil {
			return err
		}
		if put_err := cli.UpdateService(systemKey, svc, string(svc_bytes), params); put_err != nil {
			return put_err
		}
	}
	return nil
}

func push_cmd(systemKey, dir string) error {
	cli, err := auth()
	if err != nil {
		return err
	}
	sys_meta, err := pull_system_meta(systemKey, cli)
	if err != nil {
		return err
	}
	if dir == "" {
		dir = strings.Replace(sys_meta.Name, " ", "_", -1)
	}

	if svcs, err := system_diff(systemKey, dir, cli); err != nil {
		return err
	} else if len(svcs) == 0 {
		return fmt.Errorf("No services have changed, nothing to push")
	} else {
		if err := push(systemKey, dir, svcs, cli); err != nil {
			return err
		} else {
			if store_err := store_meta(dir, sys_meta); store_err != nil {
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

func export_cmd(sysKey, dir string) error {
	cli, err := auth()
	if err != nil {
		return err
	}

	sys_meta, err := pull_system_meta(sysKey, cli)
	if err != nil {
		return err
	}

	if err := createSystemDirectory(dir, sys_meta); err != nil {
		return err
	}

	if err := store_meta(dir, sys_meta); err != nil {
		return err
	}

	services, err := pull_services(sysKey, cli)
	if err != nil {
		return err
	}
	if err := store_services(sysKey, services, sys_meta); err != nil {
		return err
	}
	collections, err := pull_colls(sysKey, cli)
	if err != nil {
		return err
	}
	if err := store_cols(collections, sys_meta); err != nil {
		return err
	}

	roles, err := pull_roles(sysKey, cli)
	if err != nil {
		return err
	}
	if err := store_roles(roles, sys_meta); err != nil {
		return err
	}

	dir = strings.Replace(sys_meta.Name, " ", "_", -1)
	fmt.Printf("System %s has been successfully pulled and put in a directory %s\n", sysKey, dir)
	return nil
}

func import_cmd(dir string) error {

	cli, err := auth()
	if err != nil {
		return err
	}
	old_sys_meta, err := load_sys_meta(dir)
	if dir == "" {
		dir = strings.Replace(old_sys_meta.Name, " ", "_", -1)
	}
	if err != nil {
		fmt.Printf("Import failed - loading system metadata\n")
		return err
	}
	sysKey, err := CreateSystem(cli, old_sys_meta)
	if err != nil {
		fmt.Printf("Import failed - uploading system metadata\n")
		return err
	}

	old_collection_meta, err := load_collection_meta(dir)
	if err != nil {
		fmt.Printf("Import failed - loading collection metadata\n")
		return err
	}
	newCollections, err := CreateCollections(cli, sysKey, old_collection_meta)
	if err != nil {
		fmt.Printf("Import failed - uploading collection metadata\n")
		return err
	}

	old_services, err := load_services(dir, old_sys_meta)
	if err != nil {
		fmt.Printf("Import failed - loading service info\n")
		return err
	}

	err = CreateServices(cli, sysKey, old_services)
	if err != nil {
		fmt.Printf("Import failed - uploading service info\n")
		return err
	}

	old_roles, err := load_roles(dir)
	if err != nil {
		fmt.Printf("Import failed - retrieving roles info\n")
	}

	err = CreateRoles(cli, sysKey, old_roles)
	if err != nil {
		fmt.Printf("Import failed - uploading service info\n")
		return err
	}

	if shouldImportCollectionRows {
		err = MigrateRows(cli, old_sys_meta.Key, sysKey, old_collection_meta, newCollections)
		if err != nil {
			fmt.Printf("Import failed - uploading collection rows\n")
			return err
		}
	}

	fmt.Printf("Import successful\n")
	return nil

}

func sys_for_dir() (string, error) {
	if _, err := os.Stat(".meta.json"); os.IsNotExist(err) {
		return "", fmt.Errorf("No system key argument given and not in a system repository")
	}
	sys_meta, err := load_sys_meta(".")
	if err != nil {
		return "", fmt.Errorf("Error loading system meta data")
	}
	return sys_meta.Key, nil
}

func main() {
	flag.Parse()
	if URL != "" {
		cb.CB_ADDR = URL
	} else {
		URL = cb.CB_ADDR
	}
	cmd := strings.ToLower(flag.Arg(0))
	var err error
	var sysKey, dir string

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
			sysKey, err = sys_for_dir()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			dir = "."
		} else {
			dir = ""
			sysKey = flag.Arg(1)
		}
		if err := push_cmd(sysKey, dir); err != nil {
			fmt.Printf("Error pushing: %v\n", err)
		}
	case "export":
		if flag.NArg() != 2 {
			sysKey, err = sys_for_dir()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			dir = "."
		} else {
			dir = ""
			sysKey = flag.Arg(1)
		}
		if err := export_cmd(sysKey, dir); err != nil {
			fmt.Printf("Error export data: %v\n", err)
		}
	case "import":
		var sysKey, dir string
		var err error
		sysKey, err = sys_for_dir()
		if err != nil {
			fmt.Printf("%v\n", err)
			fmt.Printf("%v\n", sysKey)
			return
		}
		dir = "."

		if err := import_cmd(dir); err != nil {
			fmt.Printf("Error import data: %v\n", err)
		}
	default:
		fmt.Printf("Commands: 'auth', 'pull', 'push', 'export', 'import'\n")
	}
}
