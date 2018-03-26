package GoSDK

import (
	"fmt"
)

const (
	_EVENTS_DEFS_PREAMBLE  = "/admin/triggers/definitions"
	_EVENTS_HDLRS_PREAMBLE = "/admin/triggers/handlers/"
	_TIMERS_HDLRS_PREAMBLE = "/admin/triggers/timers/"
	_MH_PREAMBLE           = "/api/v/1/message/"
	_EVENTS_V3_PREAMBLE    = "/api/v/3/code/"
)

//GetEventDefinitions returns a slice of the different kinds of events that can be handled.
//Returns a slice of strings
func (d *DevClient) GetEventDefinitions() ([]interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _EVENTS_DEFS_PREAMBLE, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

//GetEventHandlers returns a slice of the event handlers for a system
func (d *DevClient) GetEventHandlers(systemKey string) ([]interface{}, error) {
	return getEventHandlers(d, _EVENTS_HDLRS_PREAMBLE+systemKey)
}

func (u *UserClient) GetEventHandlers(systemKey string) ([]interface{}, error) {
	return getEventHandlers(u, _EVENTS_V3_PREAMBLE+systemKey+"/triggers")
}

func getEventHandlers(c cbClient, endpoint string) ([]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, endpoint, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

// Alias for GetEventHandlers() to better match up with Console terminology
func (d *DevClient) GetTriggers(systemKey string) ([]interface{}, error) {
	return d.GetEventHandlers(systemKey)
}

//GetEventHandler reuturns a single event handler
//Returns an object shaped map[string]interface{}{"system_key":"associated system key","system_secret":"secret","name":"event name","event_definition":map[string]interface{}{"def_module":"module","def_name":"definition name","event_keys":[]string{"event","keys"},"visibility":false|true}, KeyVals:map[string]interface{}{"keys":"values"},"service_name":"corresponding service name"}
func (d *DevClient) GetEventHandler(systemKey, name string) (map[string]interface{}, error) {
	return getEventHandler(d, _EVENTS_HDLRS_PREAMBLE+systemKey+"/"+name)
}

func (u *UserClient) GetEventHandler(systemKey, name string) (map[string]interface{}, error) {
	return getEventHandler(u, _EVENTS_V3_PREAMBLE+systemKey+"/trigger/"+name)
}

func getEventHandler(c cbClient, endpoint string) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, endpoint, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

// Alias for GetEventHandler() to better match up with Console terminology
func (d *DevClient) GetTrigger(systemKey, name string) (map[string]interface{}, error) {
	return d.GetEventHandler(systemKey, name)
}

//CreateEventHandler creates an event handler, otherwise known as a trigger
//Returns the same object as GetEventHandler corresponding to the created event
func (d *DevClient) CreateEventHandler(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return createEventHandler(d, _EVENTS_HDLRS_PREAMBLE+systemKey+"/"+name, data)
}

func (u *UserClient) CreateEventHandler(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return createEventHandler(u, _EVENTS_V3_PREAMBLE+systemKey+"/trigger/"+name, data)
}

func createEventHandler(c cbClient, endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(c, endpoint, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

// Alias for CreateEventHandler() to better match up with Console terminology
func (d *DevClient) CreateTrigger(systemKey, name string,
	data map[string]interface{}) (map[string]interface{}, error) {

	return d.CreateEventHandler(systemKey, name, data)
}

//DeleteEventHandler removes the event handler
func (d *DevClient) DeleteEventHandler(systemKey, name string) error {
	return deleteEventHandler(d, _EVENTS_HDLRS_PREAMBLE+systemKey+"/"+name)
}

func (u *UserClient) DeleteEventHandler(systemKey, name string) error {
	return deleteEventHandler(u, _EVENTS_V3_PREAMBLE+systemKey+"/trigger/"+name)
}

func deleteEventHandler(c cbClient, endpoint string) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(c, endpoint, nil, creds, nil)
	_, err = mapResponse(resp, err)
	return err
}

// Alias for DeleteEventHandler() to better match up with Console terminology
func (d *DevClient) DeleteTrigger(systemKey, name string) error {
	return d.DeleteEventHandler(systemKey, name)
}

//UpdateEventHandler allows the developer to alter the code executed by the event handler
//Returns an object corresponding to GetEventHandler with the altered values
func (d *DevClient) UpdateEventHandler(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return updateEventHandler(d, _EVENTS_HDLRS_PREAMBLE+systemKey+"/"+name, data)
}

func (u *UserClient) UpdateEventHandler(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return updateEventHandler(u, _EVENTS_V3_PREAMBLE+systemKey+"/trigger/"+name, data)
}

func updateEventHandler(c cbClient, endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(c, endpoint, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

// Alias for UpdateEventHandler() to better match up with Console terminology
func (d *DevClient) UpdateTrigger(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return d.UpdateEventHandler(systemKey, name, data)
}

func mapResponse(resp *CbResp, err error) (*CbResp, error) {
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%+v", resp.Body)
	}
	return resp, nil
}

func mapSyncChanges(resources map[string][]string) []map[string]interface{} {
	changes := []map[string]interface{}{}
	for r, a := range resources {
		for _, n := range a {
			changes = append(changes, map[string]interface{}{
				"def_module": string(r),
				"def_name":   n,
			})
		}
	}
	return changes
}

////////////////////////////////////////////////////////////////////////////////
//
//  Timer calls are from here down

//Returns a slice of timer descriptions
//Return value looks like []interface{}[map[string]interface{}{"timer_key":"clearblade generated timer identifier","name":"the name of the timer","start_time":"rfc3339 formatted date string","repeats":0,"frequency":5,"service_name":"name of service executed","system_key":"system key associated with timer","user_id":"userid associated with timer","user_token":"a token the timer runs with"},...]
func (d *DevClient) GetTimers(systemKey string) ([]interface{}, error) {
	return getTimers(d, _TIMERS_HDLRS_PREAMBLE+systemKey)
}

func (u *UserClient) GetTimers(systemKey string) ([]interface{}, error) {
	return getTimers(u, _EVENTS_V3_PREAMBLE+systemKey+"/timers")
}

func getTimers(c cbClient, endpoint string) ([]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, endpoint, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.([]interface{}), nil
}

//GetTimer returns the definition of a single timer
//Returns a single instance of the object described in GetTimers
func (d *DevClient) GetTimer(systemKey, name string) (map[string]interface{}, error) {
	return getTimer(d, _TIMERS_HDLRS_PREAMBLE+systemKey+"/"+name)
}

func (u *UserClient) GetTimer(systemKey, name string) (map[string]interface{}, error) {
	return getTimer(u, _EVENTS_V3_PREAMBLE+systemKey+"/timer/"+name)
}

func getTimer(c cbClient, endpoint string) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(c, endpoint, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

//CreateTimer allows the user to create the timer with code
//Returns a single instance of the object described in GetTimers for the newly created timer
func (d *DevClient) CreateTimer(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return createTimer(d, _TIMERS_HDLRS_PREAMBLE+systemKey+"/"+name, data)
}

func (u *UserClient) CreateTimer(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return createTimer(u, _EVENTS_V3_PREAMBLE+systemKey+"/timer/"+name, data)
}

func createTimer(c cbClient, endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := post(c, endpoint, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

//DeleteTimer removes the timer
func (d *DevClient) DeleteTimer(systemKey, name string) error {
	return deleteTimer(d, _TIMERS_HDLRS_PREAMBLE+systemKey+"/"+name)
}

func (u *UserClient) DeleteTimer(systemKey, name string) error {
	return deleteTimer(u, _EVENTS_V3_PREAMBLE+systemKey+"/timer/"+name)
}

func deleteTimer(c cbClient, endpoint string) error {
	creds, err := c.credentials()
	if err != nil {
		return err
	}
	resp, err := delete(c, endpoint, nil, creds, nil)
	_, err = mapResponse(resp, err)
	return err
}

//UpdateTimer allows the developer to change the code executed with the timer
//Returns an updated version of the timer as described in GetTimer
func (d *DevClient) UpdateTimer(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return updateTimer(d, _TIMERS_HDLRS_PREAMBLE+systemKey+"/"+name, data)
}

func (u *UserClient) UpdateTimer(systemKey, name string, data map[string]interface{}) (map[string]interface{}, error) {
	return updateTimer(u, _EVENTS_V3_PREAMBLE+systemKey+"/timer/"+name, data)
}

func updateTimer(c cbClient, endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	creds, err := c.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := put(c, endpoint, data, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp.Body.(map[string]interface{}), nil
}

//MessageHistory allows the developer to retrieve the message history
//Returns a slice of []map[string]interface{}{map[string]interface{}{"topicid":"/topic/path", "ip":"127.0.0.1", "time":123141244, "payloadsize":12,"payload":"hello world\n","userid":"8675309","qos":0 }}
func (d *DevClient) MessageHistory(systemKey string) (interface{}, error) {
	creds, err := d.credentials()
	if err != nil {
		return nil, err
	}
	resp, err := get(d, _MH_PREAMBLE+systemKey, nil, creds, nil)
	resp, err = mapResponse(resp, err)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
