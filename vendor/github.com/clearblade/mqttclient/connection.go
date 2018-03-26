package mqttclient

import (
	"errors"
	"fmt"
	mqtt "github.com/clearblade/mqtt_parsing"
	"log"
	"strconv"
	"time"
)

const (
	CON_ACCEPT = iota
	CON_REFUSED_BAD_PROTO_VER
	CON_REFUSED_BAD_ID
	CON_REFUSED_DISALLOWED_ID
	CON_REFUSED_BAD_USERNAME_PASSWORD
	CON_REFUSED_NOT_AUTH
)

//SendConnect does precisely what it says on the tin. It returns an error if there is a problem with the authentication
func SendConnect(c *Client, lastWill, lastWillRetain bool, lastWillQOS int,
	lastWillBody, lastWillTopic string) error {
	var username string
	var password string
	var clientid string

	if c.AuthToken != "" && c.SystemKey != "" {
		username, password = c.AuthToken, c.SystemKey
	} else {

		return errors.New(fmt.Sprintf("Systemkey and auth token required, one of those was blank syskey: %s, token: %s\n", c.SystemKey, c.AuthToken))
	}
	if c.Clientid == "" {
		clientid = randStr(c)
	} else {
		clientid = c.Clientid
	}
	var lwTopic mqtt.TopicPath
	if lastWill {
		var valid bool
		lwTopic, valid = mqtt.NewTopicPath(lastWillTopic)
		if !valid {
			return fmt.Errorf("%S is an invalid topic\n", lastWillTopic)
		}
	}

	connect := &mqtt.Connect{
		ProtoName:      "MQTT",
		UsernameFlag:   true,
		Username:       username,
		PasswordFlag:   true,
		Password:       password,
		CleanSeshFlag:  true,
		KeepAlive:      60,
		ClientId:       clientid,
		WillRetainFlag: lastWillRetain,
		WillFlag:       lastWill,
		WillTopic:      lwTopic,
		WillMessage:    lastWillBody,
		WillQOS:        uint8(lastWillQOS),
		Version:        0x4,
	}
	if err := c.sendMessage(connect); err != nil {
		return err
	}
	select {
	case <-c.got_connack:
		return nil
	case <-time.After(time.Second * 30):
		return fmt.Errorf("Timed out waiting for connack")
	}

}

//MakeMeABytePublish is a helper function to create a QOS 0, non retained MQTT publish without the problems of worrying about bad unicode
//if you're not into strings
func MakeMeABytePublish(topic string, msg []byte, mid uint16) (*mqtt.Publish, error) {
	tp, valid := mqtt.NewTopicPath(topic)
	if !valid {
		return nil, fmt.Errorf("invalid topic path %s\n", topic)
	}
	return &mqtt.Publish{
		Header: &mqtt.StaticHeader{
			QOS:    0,
			Retain: false,
			DUP:    false,
		},
		Payload:   msg,
		MessageId: mid,
		Topic:     tp,
	}, nil
}

//MakeMeAPublish is a helper function for creating a publish with a string payload
func MakeMeAPublish(topic, msg string, mid uint16) (*mqtt.Publish, error) {
	tp, valid := mqtt.NewTopicPath(topic)
	if !valid {
		return nil, fmt.Errorf("invalid topic path %s\n", topic)
	}
	return &mqtt.Publish{
		Header: &mqtt.StaticHeader{
			QOS:    0,
			Retain: false,
			DUP:    false,
		},
		Payload:   []byte(msg),
		MessageId: mid,
		Topic:     tp,
	}, nil
}

//PublishFlow allows one use their mqttclient to publish a message
func PublishFlow(c *Client, p *mqtt.Publish) error {
	//make sure we don't have to commit it to storage
	//TODO:Handle high qoses
	//also, TODO::FEATURE::should this block depending on qos?
	// switch p.Header.QOS {
	// case 0:
	// 	return c.sendMessage(p)
	// }
	if p.Header.QOS > 0 && p.MessageId == 0 {
		p.MessageId = randMid(c)
	}
	return c.sendMessage(p)
}

//UnsubscribeFlow sends unsubscribes the client from a topic
func UnsubscribeFlow(c *Client, topic string) error {
	//TODO:BUG:add multiple unsubscribes at once
	//make sure you were subscribed
	if !c.subscriptions.subscription_exists(topic) && !c.waiting_for_subscription.subscription_exists("UNSUBSCRIBE#"+topic) {
		return errors.New("Not subscribed or waiting to be subscribed to topic: " + topic)
	}

	tp, valid := mqtt.NewTopicPath(topic)
	if !valid {
		return errors.New("invalid wildcard topic")
	}
	//ugh gotta do work now
	unsub := &mqtt.Unsubscribe{
		MessageId: randMid(c),
		Header: &mqtt.StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Topics: []mqtt.TopicQOSTuple{
			mqtt.TopicQOSTuple{
				Qos:   0, //qos is irrelevant here
				Topic: tp,
			},
		},
	}

	schan, err := c.waiting_for_subscription.new_subscription("UNSUBSCRIBE#" + topic)
	c.msg_store.add(unsub, unsub.MessageId)
	err = c.sendMessage(unsub)
	if err != nil {
		return err
	}
	select {
	//WARNING: HERE WE ARE ASSUMING ONE UNSUBACK PER SUBSCRIPTION
	case _ = <-schan:
		c.waiting_for_subscription.remove_subscription("UNSUBSCRIBE#" + unsub.Topics[0].Topic.Whole)
		return nil
	case <-time.After(time.Minute * 5):
		//THE ABOVE IS TOTALLY ARBITRARY
		return errors.New("Did not recieve suback after five minutes")
	}

}

//SubscribeFlow allows the client to subscribe to an mqtt topic. it returns a channel that will contain
//the publishes recieved from that particular topic, or an error
//this call blocks until a suback is recieved
func SubscribeFlow(c *Client, topic string, qos int) (<-chan *mqtt.Publish, error) {
	//NOTE: we're only allowing singleton subscriptions right now
	if qos > 2 || qos < 0 {
		return nil, errors.New("Invalid qos: " + strconv.Itoa(qos) + ". Must be less than two.")
	}

	tp, valid := mqtt.NewTopicPath(topic)
	if !valid {
		return nil, errors.New("invalid topic path")
	}
	sub := &mqtt.Subscribe{
		Header: &mqtt.StaticHeader{
			DUP:    false,
			Retain: false,
			QOS:    1,
		},
		MessageId: randMid(c),
		Subscriptions: []mqtt.TopicQOSTuple{
			mqtt.TopicQOSTuple{
				Qos:   uint8(qos),
				Topic: tp,
			},
		},
	}

	c.msg_store.add(sub, sub.MessageId)
	schan, err := c.waiting_for_subscription.new_subscription(topic)
	err = c.sendMessage(sub)
	if err != nil {
		return nil, err
	}
	//it already exists
	if err != nil {
		return nil, err
	}

	//so we're going to retry a few times. it's another register in use, but eh
	//better than blocking for eternity
	retries := 3
	for {
		select {
		//WARNING: HERE WE ARE ASSUMING ONE SUBACK PER SUBSCRIPTION
		case _ = <-schan:
			c.waiting_for_subscription.remove_subscription(topic)
			return c.subscriptions.new_outgoing(topic)
		case <-time.After(c.Timeout / 2):
			//THE ABOVE IS TOTALLY ARBITRARY
			if retries == 0 {
				return nil, errors.New("Did not recieve suback after 3 tries")
			}
			retries--
		}
		if retries > 0 {
			err = c.sendMessage(sub)
			if err != nil {
				log.Println("error resending subscribe", err.Error())
			}
		}
	}
}

//Sends a disconnect
func SendDisconnect(c *Client) error {
	return c.sendMessage(&mqtt.Disconnect{})
}

//allows us to get a random string
func randStr(c *Client) string {
	var out string
	const charset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	order := c.randoPerm(23)
	//permute order of string
	for _, v := range order {
		out += string(charset[v])
	}
	return out
}

func randMid(c *Client) uint16 {
	return uint16(c.getInt())
}
