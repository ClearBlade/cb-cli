Go-SDK
======

# API Reference

## Setup
### newUserClient(systemkey, systemsecret, email, password string) *UserClient
		systemKey // string
			Required. The systemKey of the system to connect to. Retrievable from the Console's System Settings.
		systemSecret // string
			Required. The systemSecret of the system to connect to. Retrievable from the Console's System Settings.
		email // string
			Email of non-dev user to connect to system as. If registerUser key is not provided, the user must be registered through the Auth tab of the console, and given appropriate roles.
		password // string
			Password of non-dev user to connect to system as.
## Authentication
### userClient.Authenticate() error
Authenticates credentials set on userClient and sets session token
### userClient.Register(username, password string) error
Register a new user with the platform.

		username // string
			Email of the new user
		password // string
			Password of the new user
### userClient.RegisterUser(username, password string) (map[string]interface{}, error)
Register a new user with the platform, and return response with token

		username // string
			Email of the new user
		password // string
			Password of the new user
### userClient.Logout() error
End session for current user


## Data
### userClient.InsertData(collection_id string, data interface{}) error
Insert new item in the collection

		collection_id // string
			ID assigned to the collection by the system
		data // interface
			Representatation of data object to be inserted
### userClient.CreateData(collection_id string, data interface{}) ([]interface{}, error)
Insert new item in the collection and return the response

		collection_id // string
			ID assigned to the collection by the system
		data // interface
			Representation of data object to be inserted
### userClient.GetData(collectionId string, query *GoSDK.Query) (map[string]interface{}, error)
Retrieve array of items from the collection using the collection ID

		collectionId // string
			ID assigned to the collection by the system
		query // *GoSDK.Query
			Custom query created using this SDK
### userClient.GetDataByName(collectionName string, query *GoSDK.Query) (map[string]interface{}, error)
Retrieve array of items from the collection using the collection name

		collectionName // string
			Name assigned to the collection by the developer
		query // *GoSDK.Query
			Custom query created using this SDK
### userClient.UpdateData(collection_id string, query *GoSDK.Query, changes map[string]interface{}) error
Updates existing items in the collection that match the provided query

		collectionId // string
			ID assigned to the collection by the system
		query // *GoSDK.Query
			Custom query created using this SDK
		changes // map[string]interface{}
			Key-value pairs representing column names to be updated and new values
### userClient.DeleteData(collection_id string, query *GoSDK.Query) error
Removes every item in the collection that match the provided query

		collectionId // string
			ID assigned to the collection by the system
		query // *GoSDK.Query
			Custom query created using this SDK
### userClient.GetColumns(collection_id string) ([]interface{}, error)
Retrieves column names, types and primary keys for a collection

		collectionId // string
			ID assigned to the collection by the system
## Queries
### query := NewQuery() *GoSDK.Query
Returns new Query to be used in Data operations

### query.EqualTo(field string, value interface{})
Select where field is equal to value

		field // string
			Field / column name in collection
		value // interface{}
			Data to match in field
### query.GreaterThan(field string, value interface{})
Select where field is > to value

		field // string
			Field / column name in collection
		value // interface{}
			Data to match in field
### query.GreaterThanEqualTo(field string, value interface{})
Select where field is >= to value

		field // string
			Field / column name in collection
		value // interface{}
			Data to match in field
### query.LessThan(field string, value interface{})
Select where field is < to value

		field // string
			Field / column name in collection
		value // interface{}
			Data to match in field
### query.LessThanEqualTo(field string, value interface{})
Select where field is <= to value

		field // string
			Field / column name in collection
		value // interface{}
			Data to match in field
### query.NotEqualTo(field string, value interface{})
Select where field is != to value

		field // string
			Field / column name in collection
		value // interface{}
			Data to match in field
### query.Matches(field string, value interface{})
Query where field is ~ than value using regex

		field // string
			Field / column name in collection
		value // interface{}
			Data to match in field
### query.Or(orQuery *GoSDK.Query)
Join two queries together with OR condition

		orQuery // *GoSDK.Query
			Second GoSDk.Query
## Code Services
### userClient.CallService(systemKey, name string, params map[string]interface{}) (map[string]interface{}, error)
Run a named code service that exists on the system with options, and returns the response object or an error

		systemKey // string
			Key to system that contains desired code service
		name // string
			Name of code service to be executed
		params // map[string]interface{}
			Request object to be passed to code service function on execution
## Messaging
### userClient.InitializeMQTT(clientid string, ignore string, timeout int) error
Set required MQTT options on user client

		clientid // string
			MQTT client id
		ignore // string
			Not used, default to ""
		timeout // int
			Timeout value for MQTT client in Seconds
### userClient.ConnectMQTT(ssl *tls.Config, lastWill *GoSDK.LastWillPacket) error
Establish MQTT connection for set user

		clientid // *tls.Config,
		lastWill // *GoSDK.LastWillPacket, 
### userClient.Publish(topic string, message []byte, qos int) error
Publish message on MQTT topic

		topic // string
			MQTT topic to publish on
		message // []byte
			Byte-seralized message
		qos // int
			QOS priority for message
### userClient.Subscribe(topic string, qos int) (<-chan *mqtt.Publish, error)
Asynchronoulsy subscribe to MQTT topic and push messages onto channel as they arrive

		topic // string
			MQTT topic to subscribe
		qos // int
			QOS priority for message
### userClient.Unsubscribe(topic string) error
Stop subscription to MQTT topic

		topic // string
			MQTT topic to unsusbcribe
### userClient.Disconnect() error
End MQTT connection for set user

# QuickStart


### Download and Install Go
Follow the instructions for downloading and installing Go from https://golang.org/doc/install  

### Clone the ClearBlade Go-SDK repository    
Do a ``` git clone https://github.com/ClearBlade/Go-SDK.git``` to clone the Go-SDK for the ClearBlade Platform.

### Import the Go-SDK package in your project  
In order to use the Go-SDK in your project, you will need to import the Go-SDK package that you cloned from github, in your project.

# GoDoc

The GoDoc for the Go API can be found at https://docs.clearblade.com/v/3/static/goapi/pkg/github.com/clearblade/Go-SDK/index.html


