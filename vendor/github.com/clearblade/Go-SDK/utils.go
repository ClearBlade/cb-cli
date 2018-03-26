package GoSDK

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	mqtt "github.com/clearblade/mqtt_parsing"
	"github.com/clearblade/mqttclient"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var (
	//CB_ADDR is the address of the ClearBlade Platform you are speaking with
	CB_ADDR = "https://platform.clearblade.com"
	//CB_MSG_ADDR is the messaging address you wish to speak to
	CB_MSG_ADDR = "platform.clearblade.com:1883"

	_HEADER_KEY_KEY    = "ClearBlade-SystemKey"
	_HEADER_SECRET_KEY = "ClearBlade-SystemSecret"
)

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
}

const (
	createDevUser = iota
	createUser
)

//Client is a convience interface for API consumers, if they want to use the same functions for both developer users and unprivleged users
type Client interface {
	//session bookkeeping calls
	Authenticate() error
	Logout() error

	//data calls
	InsertData(string, interface{}) error
	UpdateData(string, *Query, map[string]interface{}) error
	GetData(string, *Query) (map[string]interface{}, error)
	GetDataByName(string, *Query) (map[string]interface{}, error)
	GetDataByKeyAndName(string, string, *Query) (map[string]interface{}, error)
	DeleteData(string, *Query) error

	//mqtt calls
	InitializeMQTT(string, string, int) error
	ConnectMQTT(*tls.Config, *LastWillPacket) error
	Publish(string, []byte, int) error
	Subscribe(string, int) (<-chan *mqtt.Publish, error)
	Unsubscribe(string) error
	Disconnect() error
}

//cbClient will supply various information that differs between privleged and unprivleged users
//this interface is meant to be unexported
type cbClient interface {
	credentials() ([][]string, error) //the inner slice is a tuple of "Header":"Value"
	preamble() string
	setToken(string)
	getToken() string
	getSystemInfo() (string, string)
	getMessageId() uint16
	getHttpAddr() string
	getMqttAddr() string
}

//UserClient is the type for users
type UserClient struct {
	UserToken    string
	mrand        *rand.Rand
	MQTTClient   *mqttclient.Client
	SystemKey    string
	SystemSecret string
	Email        string
	Password     string
	HttpAddr     string
	MqttAddr     string
}

//DevClient is the type for developers
type DevClient struct {
	DevToken   string
	mrand      *rand.Rand
	MQTTClient *mqttclient.Client
	Email      string
	Password   string
	HttpAddr   string
	MqttAddr   string
}

//CbReq is a wrapper around an HTTP request
type CbReq struct {
	Body        interface{}
	Method      string
	Endpoint    string
	QueryString string
	Headers     map[string][]string
	HttpAddr    string
	MqttAddr    string
}

//CbResp is a wrapper around an HTTP response
type CbResp struct {
	Body       interface{}
	StatusCode int
}

func (u *UserClient) getHttpAddr() string {
	return u.HttpAddr
}

func (d *DevClient) getHttpAddr() string {
	return d.HttpAddr
}

func (u *UserClient) getMqttAddr() string {
	return u.MqttAddr
}

func (d *DevClient) getMqttAddr() string {
	return d.MqttAddr
}

//NewUserClient allocates a new UserClient struct
func NewUserClient(systemkey, systemsecret, email, password string) *UserClient {
	return &UserClient{
		UserToken:    "",
		mrand:        rand.New(rand.NewSource(time.Now().UnixNano())),
		MQTTClient:   nil,
		SystemSecret: systemsecret,
		SystemKey:    systemkey,
		Email:        email,
		Password:     password,
		HttpAddr:     CB_ADDR,
		MqttAddr:     CB_MSG_ADDR,
	}
}

//NewDevClient allocates a new DevClient struct
func NewDevClient(email, password string) *DevClient {
	return &DevClient{
		DevToken:   "",
		mrand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		MQTTClient: nil,
		Email:      email,
		Password:   password,
		HttpAddr:   CB_ADDR,
		MqttAddr:   CB_MSG_ADDR,
	}
}

func NewDevClientWithToken(token, email string) *DevClient {
	return &DevClient{
		DevToken:   token,
		mrand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		MQTTClient: nil,
		Email:      email,
		Password:   "",
		HttpAddr:   CB_ADDR,
		MqttAddr:   CB_MSG_ADDR,
	}
}

func NewUserClientWithAddrs(httpAddr, mqttAddr, systemKey, systemSecret, email, password string) *UserClient {
	return &UserClient{
		UserToken:    "",
		mrand:        rand.New(rand.NewSource(time.Now().UnixNano())),
		MQTTClient:   nil,
		SystemSecret: systemSecret,
		SystemKey:    systemKey,
		Email:        email,
		Password:     password,
		HttpAddr:     httpAddr,
		MqttAddr:     mqttAddr,
	}
}
func NewDevClientWithAddrs(httpAddr, mqttAddr, email, password string) *DevClient {
	return &DevClient{
		DevToken:   "",
		mrand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		MQTTClient: nil,
		Email:      email,
		Password:   password,
		HttpAddr:   httpAddr,
		MqttAddr:   mqttAddr,
	}
}

func NewDevClientWithTokenAndAddrs(httpAddr, mqttAddr, token, email string) *DevClient {
	return &DevClient{
		DevToken:   token,
		mrand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		MQTTClient: nil,
		Email:      email,
		Password:   "",
		HttpAddr:   httpAddr,
		MqttAddr:   mqttAddr,
	}
}

//Authenticate retrieves a token from the specified Clearblade Platform
func (u *UserClient) Authenticate() error {
	return authenticate(u, u.Email, u.Password)
}

func (u *UserClient) AuthAnon() error {
	return authAnon(u)
}

//Authenticate retrieves a token from the specified Clearblade Platform
func (d *DevClient) Authenticate() error {
	return authenticate(d, d.Email, d.Password)
}

//Register creates a new user
func (u *UserClient) Register(username, password string) error {
	if u.UserToken == "" {
		return fmt.Errorf("Must be logged in to create users")
	}
	_, err := register(u, createUser, username, password, u.SystemKey, u.SystemSecret, "", "", "")
	return err
}

//RegisterUser creates a new user, returning the body of the response.
func (u *UserClient) RegisterUser(username, password string) (map[string]interface{}, error) {
	if u.UserToken == "" {
		return nil, fmt.Errorf("Must be logged in to create users")
	}
	resp, err := register(u, createUser, username, password, u.SystemKey, u.SystemSecret, "", "", "")
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//Registers a new developer
func (d *DevClient) Register(username, password, fname, lname, org string) error {
	resp, err := register(d, createDevUser, username, password, "", "", fname, lname, org)
	if err != nil {
		return err
	} else {
		d.DevToken = resp["dev_token"].(string)
		return nil
	}
}

func (d *DevClient) RegisterNewUser(username, password, systemkey, systemsecret string) (map[string]interface{}, error) {
	if d.DevToken == "" {
		return nil, fmt.Errorf("Must authenticate first")
	}
	return register(d, createUser, username, password, systemkey, systemsecret, "", "", "")

}

//Register creates a new developer user
func (d *DevClient) RegisterDevUser(username, password, fname, lname, org string) (map[string]interface{}, error) {
	resp, err := register(d, createDevUser, username, password, "", "", fname, lname, org)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//Logout ends the session
func (u *UserClient) Logout() error {
	return logout(u)
}

//Logout ends the session
func (d *DevClient) Logout() error {
	return logout(d)
}

//Below are some shared functions

func authenticate(c cbClient, username, password string) error {
	var creds [][]string
	switch c.(type) {
	case *UserClient:
		var err error
		creds, err = c.credentials()
		if err != nil {
			return err
		}
	case *DevClient:
	}
	resp, err := post(c, c.preamble()+"/auth", map[string]interface{}{
		"email":    username,
		"password": password,
	}, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error in authenticating, Status Code: %d, %v\n", resp.StatusCode, resp.Body)
	}

	var token string = ""
	switch c.(type) {
	case *UserClient:
		token = resp.Body.(map[string]interface{})["user_token"].(string)
	case *DevClient:
		token = resp.Body.(map[string]interface{})["dev_token"].(string)
	}
	if token == "" {
		return fmt.Errorf("Token not present i response from platform %+v", resp.Body)
	}
	c.setToken(token)
	return nil
}

func authAnon(c cbClient) error {
	creds, err := c.credentials()
	if err != nil {
		return fmt.Errorf("Invalid client: %+s", err.Error())
	}
	resp, err := post(c, c.preamble()+"/anon", nil, creds, nil)
	if err != nil {
		return fmt.Errorf("Error retrieving anon user token: %s", err.Error())
	}
	token := resp.Body.(map[string]interface{})["user_token"].(string)
	if token == "" {
		return fmt.Errorf("Token not present in response from platform %+v", resp.Body)
	}
	c.setToken(token)
	return nil
}

func register(c cbClient, kind int, username, password, syskey, syssec, fname, lname, org string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"email":    username,
		"password": password,
	}
	var endpoint string
	headers := make(map[string][]string)
	var creds [][]string
	switch kind {
	case createDevUser:
		endpoint = "/admin/reg"
		payload["fname"] = fname
		payload["lname"] = lname
		payload["org"] = org
	case createUser:
		switch c.(type) {
		case *DevClient:
			if syskey == "" {
				return nil, fmt.Errorf("System key required")
			}
			endpoint = fmt.Sprintf("/admin/user/%s", syskey)
		case *UserClient:
			if syskey == "" {
				return nil, fmt.Errorf("System key required")
			}
			if syssec == "" {
				return nil, fmt.Errorf("System secret required")
			}
			endpoint = "/api/v/1/user/reg"
			headers["Clearblade-Systemkey"] = []string{syskey}
			headers["Clearblade-Systemsecret"] = []string{syssec}
		default:
			return nil, fmt.Errorf("unreachable code detected")
		}
		var err error
		creds, err = c.credentials()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Cannot create that kind of user")
	}
	resp, err := post(c, endpoint, payload, creds, headers)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Status code: %d, Error in authenticating, %v\n", resp.StatusCode, resp.Body)
	}
	var token string = ""
	switch kind {
	case createDevUser:
		token = resp.Body.(map[string]interface{})["dev_token"].(string)
	case createUser:
		token = resp.Body.(map[string]interface{})["user_id"].(string)
	}

	if token == "" {
		return nil, fmt.Errorf("Token not present in response from platform %+v", resp.Body)
	}
	return resp.Body.(map[string]interface{}), nil
}

func logout(c cbClient) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := post(c, c.preamble()+"/logout", nil, creds, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error in authenticating %v\n", resp.Body)
	}
	return nil
}

func do(c cbClient, r *CbReq, creds [][]string) (*CbResp, error) {
	var bodyToSend *bytes.Buffer
	if r.Body != nil {
		b, jsonErr := json.Marshal(r.Body)
		if jsonErr != nil {
			return nil, fmt.Errorf("JSON Encoding Error: %v", jsonErr)
		}
		bodyToSend = bytes.NewBuffer(b)
	} else {
		bodyToSend = nil
	}
	url := c.getHttpAddr() + r.Endpoint
	if r.QueryString != "" {
		url += "?" + r.QueryString
	}
	var req *http.Request
	var reqErr error
	if bodyToSend != nil {
		req, reqErr = http.NewRequest(r.Method, url, bodyToSend)
	} else {
		req, reqErr = http.NewRequest(r.Method, url, nil)
	}
	if reqErr != nil {
		return nil, fmt.Errorf("Request Creation Error: %v", reqErr)
	}
	if r.Headers != nil {
		for hed, val := range r.Headers {
			for _, vv := range val {
				req.Header.Add(hed, vv)
			}
		}
	}
	for _, c := range creds {
		if len(c) != 2 {
			return nil, fmt.Errorf("Request Creation Error: Invalid credential header supplied")
		}
		req.Header.Add(c[0], c[1])
	}

	cli := &http.Client{Transport: tr}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error Making Request: %v", err)
	}
	defer resp.Body.Close()
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("Error Reading Response Body: %v", readErr)
	}
	var d interface{}
	if len(body) == 0 {
		return &CbResp{
			Body:       nil,
			StatusCode: resp.StatusCode,
		}, nil
	}
	buf := bytes.NewBuffer(body)
	dec := json.NewDecoder(buf)
	decErr := dec.Decode(&d)
	var bod interface{}
	if decErr != nil {
		//		return nil, fmt.Errorf("JSON Decoding Error: %v\n With Body: %v\n", decErr, string(body))
		bod = string(body)
	}
	switch d.(type) {
	case []interface{}:
		bod = d
	case map[string]interface{}:
		bod = d
	default:
		bod = string(body)
	}
	return &CbResp{
		Body:       bod,
		StatusCode: resp.StatusCode,
	}, nil
}

//standard http verbs

func get(c cbClient, endpoint string, query map[string]string, creds [][]string, headers map[string][]string) (*CbResp, error) {
	req := &CbReq{
		Body:        nil,
		Method:      "GET",
		Endpoint:    endpoint,
		QueryString: query_to_string(query),
		Headers:     headers,
	}
	return do(c, req, creds)
}

func post(c cbClient, endpoint string, body interface{}, creds [][]string, headers map[string][]string) (*CbResp, error) {
	req := &CbReq{
		Body:        body,
		Method:      "POST",
		Endpoint:    endpoint,
		QueryString: "",
		Headers:     headers,
	}
	return do(c, req, creds)
}

func put(c cbClient, endpoint string, body interface{}, heads [][]string, headers map[string][]string) (*CbResp, error) {
	req := &CbReq{
		Body:        body,
		Method:      "PUT",
		Endpoint:    endpoint,
		QueryString: "",
		Headers:     headers,
	}
	return do(c, req, heads)
}

func delete(c cbClient, endpoint string, query map[string]string, heds [][]string, headers map[string][]string) (*CbResp, error) {
	req := &CbReq{
		Body:        nil,
		Method:      "DELETE",
		Endpoint:    endpoint,
		Headers:     headers,
		QueryString: query_to_string(query),
	}
	return do(c, req, heds)
}

func query_to_string(query map[string]string) string {
	qryStr := ""
	for k, v := range query {
		qryStr += k + "=" + v + "&"
	}
	return strings.TrimSuffix(qryStr, "&")
}
