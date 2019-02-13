package GoSDK

import (
	"fmt"
)

//This file provides the interface for establishing connect collections
//that is to say, collections that are interfaced with non-platform databases
//they have to be treated a little bit differently, because a lot of configuration information
//needs to be trucked across the line during setup. enough that it's more helpful to have it in a
//struct than it is just in a map, or an endless list of function arguments.

type connectCollection interface {
	toMap() map[string]interface{}
	tableName() string
	name() string
}

//MySqlConfig houses configuration information for an MySql-backed collection
type MySqlConfig struct {
	Name, User, Password, Host, Port, DBName, Tablename string
}

func (my MySqlConfig) tableName() string { return my.Tablename }
func (my MySqlConfig) name() string      { return my.Name }

func (my MySqlConfig) toMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["name"] = my.Name
	m["user"] = my.User
	m["password"] = my.Password
	m["address"] = my.Host
	m["port"] = my.Port
	m["dbname"] = my.DBName
	m["tablename"] = my.Tablename
	m["dbtype"] = "mysql"
	return m
}

//MSSqlConfig houses configuration information for an MSSql-backed collection
type MSSqlConfig struct {
	Name, User, Password, Host, Port, DBName, Tablename string
}

func (ms MSSqlConfig) tableName() string { return ms.Tablename }
func (ms MSSqlConfig) name() string      { return ms.Tablename }

func (ms MSSqlConfig) toMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["user"] = ms.User
	m["password"] = ms.Password
	m["address"] = ms.Host
	m["port"] = ms.Port
	m["dbname"] = ms.DBName
	m["tablename"] = ms.Tablename
	m["dbtype"] = "mssql"
	m["name"] = ms.Name
	return m
}

//PostgresqlConfig houses configuration information for an Postgresql-backed collection
type PostgresqlConfig struct {
	Name, User, Password, Host, Port, DBName, Tablename string
}

func (pg PostgresqlConfig) toMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["user"] = pg.User
	m["password"] = pg.Password
	m["address"] = pg.Host
	m["port"] = pg.Port
	m["dbname"] = pg.DBName
	m["tablename"] = pg.Tablename
	m["dbtype"] = "postgres"
	m["name"] = pg.Name
	return m
}

func (pg PostgresqlConfig) tableName() string { return pg.Tablename }
func (pg PostgresqlConfig) name() string      { return pg.Tablename }

type MongoDBConfig struct {
	Name, User, Password, Host, Port, DBName, Tablename string
}

func (mg MongoDBConfig) toMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["user"] = mg.User
	m["password"] = mg.Password
	m["address"] = mg.Host
	m["port"] = mg.Port
	m["dbname"] = mg.DBName
	m["tablename"] = mg.Tablename
	m["dbtype"] = "MongoDB"
	m["name"] = mg.Name
	return m
}

func (mg MongoDBConfig) tableName() string { return mg.Tablename }
func (mg MongoDBConfig) name() string      { return mg.Tablename }

func GenerateConnectCollection(co map[string]interface{}) (connectCollection, error) {
	dbtype, ok := co["dbtype"].(string)
	if !ok {
		return nil, fmt.Errorf("generateConnectCollection: dbtype field missing or is not a string")
	}
	switch dbtype {
	case "mysql":
		return &MySqlConfig{
			User:      co["user"].(string),
			Password:  co["password"].(string),
			Host:      co["address"].(string),
			Port:      co["port"].(string),
			DBName:    co["dbname"].(string),
			Tablename: co["tablename"].(string),
			Name:      co["name"].(string),
		}, nil
	case "mssql":
		return &MSSqlConfig{
			User:      co["user"].(string),
			Password:  co["password"].(string),
			Host:      co["address"].(string),
			Port:      co["port"].(string),
			DBName:    co["dbname"].(string),
			Tablename: co["tablename"].(string),
			Name:      co["name"].(string),
		}, nil
	case "postgresql":
		return &PostgresqlConfig{
			User:      co["user"].(string),
			Password:  co["password"].(string),
			Host:      co["address"].(string),
			Port:      co["port"].(string),
			DBName:    co["dbname"].(string),
			Tablename: co["tablename"].(string),
			Name:      co["name"].(string),
		}, nil
	case "MongoDB":
		return &MongoDBConfig{
			User:      co["user"].(string),
			Password:  co["password"].(string),
			Host:      co["address"].(string),
			Port:      co["port"].(string),
			DBName:    co["dbname"].(string),
			Tablename: co["tablename"].(string),
			Name:      co["name"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("generateConnectCollection: Unknown connect database type: '%s'\n", dbtype)
	}
}
