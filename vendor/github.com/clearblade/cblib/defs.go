package cblib

//
//  These are variables that can be used as
//  flags to a main package using this library, or
//  they can be set directly by unit tests, etc.
//  My, there are a lot of these...
//

const (
	NotExistErrorString    = "Does not exist"
	SpecialNoCBMetaError   = "No cbmeta file"
	ExportItemIdDefault    = true
	SortCollectionsDefault = false
	DataPageSizeDefault    = 100
)

var (
	URL                        string
	MsgURL                     string
	SystemKey                  string
	DevToken                   string
	ShouldImportCollectionRows bool
	ImportRows                 bool
	ExportRows                 bool
	ExportItemId               bool
	ImportUsers                bool
	ExportUsers                bool
	CleanUp                    bool
	EdgeSchema                 bool
	DeviceSchema               bool
	DeviceSchemaPresent        bool
	UserSchema                 bool
	DataPageSize               int
	Email                      string
	Password                   string
	ServiceName                string
	LibraryName                string
	CollectionName             string
	CollectionId               string
	SortCollections            bool
	User                       string
	UserId                     string
	RoleName                   string
	TriggerName                string
	TimerName                  string
	DeviceName                 string
	EdgeName                   string
	PortalName                 string
	PluginName                 string
	AdaptorName                string
	Message                    bool
	Topic                      string
	Payload                    string
	Help                       bool
	Params                     string
	Push                       bool
	AllServices                bool
	AllLibraries               bool
	AllDevices                 bool
	AllEdges                   bool
	AllPortals                 bool
	AllPlugins                 bool
	AllAdaptors                bool
	TempDir                    string
)

var (
	systemDotJSON map[string]interface{}
	libCode       map[string]interface{}
	svcCode       map[string]interface{}
	MetaInfo      map[string]interface{}
)

var (
	DefaultEdgeColumns   = []string{"edge_key", "novi_system_key", "system_key", "system_secret", "token", "name", "description", "location", "mac_address", "public_addr", "public_port", "local_addr", "local_port", "broker_port", "broker_tls_port", "broker_ws_port", "broker_wss_port", "broker_auth_port", "broker_ws_auth_port", "first_talked", "last_talked", "communication_style", "last_seen_version", "policy_name", "resolver_func", "sync_edge_tables"}
	DefaultDeviceColumns = []string{"device_key", "name", "system_key", "type", "state", "description", "enabled", "allow_key_auth", "active_key", "keys", "allow_certificate_auth", "certificate", "created_date", "last_active_date"}
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

type User_meta struct {
	Columns []Column
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
	Secret      string
	Description string
	Services    map[string]Service_meta
	PlatformUrl string
	MessageUrl  string
}
