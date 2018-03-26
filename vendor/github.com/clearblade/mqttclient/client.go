package mqttclient

import (
	"crypto/tls"
	"errors"
	"fmt"
	mqtt "github.com/clearblade/mqtt_parsing"
	"io"
	mrand "math/rand"
	"net"
	"sync"
	"time"
)

type Client struct {

	//this is all clearblade specific
	SystemKey    string
	SystemSecret string
	AuthToken    string
	Clientid     string
	//this will usually be occupied
	//by a net.Conn
	C                   io.ReadWriteCloser
	internalOutgoingBuf chan []byte
	Timeout             time.Duration
	//thou shalt type channels consumed by others
	ClientErrorBuffer chan error

	internalErrorBuffer chan *errWrap
	shutdown_reader     chan struct{}
	shutdown_writer     chan struct{}
	//introduce a sync write mode?
	last_timeout_reccd time.Time

	shutting_down            bool
	got_connack              chan struct{}
	msg_store                *storage
	subscriptions            *outgoing_topics
	waiting_for_subscription *subscription_store
	//TODO:redesign around a thread-local
	//rng
	rando      *mrand.Rand
	randomut   *sync.RWMutex
	resetTimer chan bool
}

var (
	Verbose bool
)

//Start connects to the mqtt broker. It does not send the connect packet. Use the SendConnect function for that.
func (c *Client) Start(addr string, ssl *tls.Config) error {
	var con net.Conn
	var err error
	if ssl != nil {
		con, err = tls.Dial("tcp", addr, ssl)
	} else {
		con, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}
	c.C = con

	go c.connectionWriter()
	go c.connectionListener()
	go c.errorTree()
	return nil
}

//NewClient allocates a new client. It is supplied with the (in order of appearance)
//Token, SystemKey,SystemSecret,Clientid, and the mqtt timeout
//Note that the following combinations of (Token|SystemKey|SystemSecret) are allowed
//(Token && SystemKey), (SystemKey && SystemSecret)
func NewClient(tok, sk, ss, cid string, timeout int) *Client {
	client := &Client{
		msg_store:                newStorage(),
		waiting_for_subscription: newSubscriptionStore(),
		subscriptions:            newOutgoingTopics(),
		SystemSecret:             ss,
		SystemKey:                sk,
		AuthToken:                tok,
		Clientid:                 cid,
		Timeout:                  time.Duration(timeout) * time.Second,
		//internalOutgoingBuf:      make(chan []byte, 30),
		internalOutgoingBuf: make(chan []byte),
		ClientErrorBuffer:   make(chan error, 10),
		internalErrorBuffer: make(chan *errWrap, 2),
		shutdown_reader:     make(chan struct{}, 1),
		shutdown_writer:     make(chan struct{}, 1),
		rando:               mrand.New(mrand.NewSource(time.Now().UnixNano())),
		got_connack:         make(chan struct{}, 1),
		randomut:            new(sync.RWMutex),
		resetTimer:          make(chan bool, 1),
	}
	return client
}

//sendMessage is an internal function that acts as a central point
//of failure for all of the message sending channels
//sort of like a fan-in, except this simply allows us to do
// all of the error handling logic in one place
func (c *Client) sendMessage(m mqtt.Message) error {
	_, err := c.C.Write(m.Encode())
	if err != nil {
		c.internalErrorBuffer <- &errWrap{
			err:      err,
			reciever: _CON_WRITER,
		}
		return err
	}
	select {
	case c.resetTimer <- true:
	default:
	}
	return nil
}

//connectionWriter is an internal function. it essentially sits in a goroutine and writes to the connection
//whenever it recieves a message over the channel
func (c *Client) connectionWriter() {
	if c.C == nil {
		return
	}
	for {
		select {
		case out := <-c.internalOutgoingBuf:
			_, err := c.C.Write(out)
			if err != nil {
				c.internalErrorBuffer <- &errWrap{
					err:      err,
					reciever: _CON_WRITER,
				}
				return
			}
		case <-c.shutdown_writer:
			return
		}
	}
}

//connectionListener is another function that sits on a goroutine.
//DecodePacket blocks until it reads a complete mqtt packet
func (c *Client) connectionListener() {
	if c.C == nil {
		return
	}
	mch, ech, shutdown := make(chan mqtt.Message, 10), make(chan error, 1), false
	//we have to establish an internal chain of goroutines here
	//otherwise we couldn't shutdown the listener on demand
	//since it's really hard to coordinate all the shutting down
	//when a connection drops
	//we're waiting for the connection listener to simply fail
	//this allows us to handle it a bit more gracefully
	//in order to shut down via channels directly we'd have to
	//wait for the read to fail anyway.
	go func(m chan mqtt.Message, e chan error) {
		for {
			msg, err := mqtt.DecodePacket(c.C)
			if err != nil {
				shutdown = true
				e <- err
				return
			}
			m <- msg
		}
	}(mch, ech)

	heardFromServer := true
	for {
		select {
		case <-c.resetTimer:
		case msg := <-mch:
			heardFromServer = true
			c.dispatch(msg)
		case e := <-ech:
			if !shutdown {
				c.internalErrorBuffer <- &errWrap{
					err:      e,
					reciever: _CON_READER,
				}
			}
		case <-c.shutdown_reader:
			shutdown = true
			return
		case <-time.After(c.Timeout):
			heardFromServer = false
		}
		if !heardFromServer {
			c.sendMessage(&mqtt.Pingreq{})
		}
	}
}

//this is the "business logic" of the client. it decides how each packet mutates the
//internal state of the client
func (c *Client) dispatch(msg mqtt.Message) {
	//for example we make a decision if the client sees this request or not
	//or if we have to send another message in a flow
	c.last_timeout_reccd = time.Now()
	switch msg.Type() {
	case mqtt.CONNECT:
		//shouldn't happen?
	case mqtt.CONNACK:
		if msg.(*mqtt.Connack).ReturnCode != 0 {
			c.internalErrorBuffer <- &errWrap{
				err:      fmt.Errorf("Got return type %d instead of 0\n", msg.(*mqtt.Connack).ReturnCode),
				reciever: _OTHER,
			}
		} else {
			c.got_connack <- struct{}{}
		}
	case mqtt.PUBLISH:
		pubMsg := msg.(*mqtt.Publish)
		c.subscriptions.relay_message(pubMsg, pubMsg.Topic.Whole)
		switch msg.(*mqtt.Publish).Header.QOS {
		case 1:
			c.sendMessage(&mqtt.Puback{
				MessageId: pubMsg.MessageId,
			})
		case 2:
			c.sendMessage(&mqtt.Pubrec{
				MessageId: pubMsg.MessageId,
			})
		}
	case mqtt.PUBACK:
		//TODO:handle resend
	case mqtt.PUBREC:
		//discard, store the fact that it was recieved
		c.sendMessage(&mqtt.Pubrel{
			MessageId: msg.(*mqtt.Pubrec).MessageId,
			Header: &mqtt.StaticHeader{
				DUP:    false,
				Retain: false,
				QOS:    1,
			}})
	case mqtt.PUBREL:
		//this shouldn't have happened
	case mqtt.SUBSCRIBE:
		//this is not supposed to happen
	case mqtt.SUBACK:
		//the subscribe call blocks, so we need to forward the message
		//along that the subscribe was acknowleged so we can return
		//control flow to the parent program
		//of course, the problem is that a suback does not have
		//the subscriptions in it by name
		//but it does have the same message id
		//so we have to retrieve that and then match up the subscribe
		//TODO:NOTE THAT WE ARE ONLY USING ONE TOPIC PER SUBSCRIBE MESSSAGE
		//THIS LOGIC WILL NEED TWEAKING IF THAT CHANGES
		msg := c.msg_store.getEntry(msg.(*mqtt.Suback).MessageId)
		if msg == nil {
			//this is a bad thing to happen
			return
		}
		sub, ok := msg.(*mqtt.Subscribe)
		if !ok {
			//this is a worse thing to happen
			return
		}
		//TODO: BUG:: IF WE MAKE MULTISUBSCRIPTION, THIS WILL BREAK
		//we've now released control flow in that channel
		//also it'll allocate the userside channel and all that good stuff
		c.waiting_for_subscription.relay_message(msg, sub.Subscriptions[0].Topic.Whole)

	case mqtt.UNSUBSCRIBE:
		//this shouldn't happen
	case mqtt.UNSUBACK:
		//not gorgeous, but it do the thing
		msg := c.msg_store.getEntry(msg.(*mqtt.Unsuback).MessageId)
		unsub, ok := msg.(*mqtt.Unsubscribe)
		if !ok {
			return
		}
		c.subscriptions.remove_subscription(unsub.Topics[0].Topic.Whole)
		//TODO: do something besides prepend UNSUBSCRIBE# (the hash makes it an invalid mqtt topic, which should prevent collisions) to the front of the topic
		//but also doesn't allocate another big-ole map wrapper
		err := c.waiting_for_subscription.relay_message(msg, "UNSUBSCRIBE#"+unsub.Topics[0].Topic.Whole)
		if err != nil {
			c.ClientErrorBuffer <- err
		}
	case mqtt.PINGREQ:
		//shouldn't happen
	case mqtt.PINGRESP:
		//pingresp will reset the counter elsewhere
	case mqtt.DISCONNECT:
		//shouldn't happen
	default:
		c.ClientErrorBuffer <- fmt.Errorf("Invalid mqtt type recieved %+v", msg)
	}
}

//errorTree is the goroutine that sits on it's own goroutine and waits for
//a message to be recieved on c.internalErrorBuffer. It's our "in case of emergency break glass"
//way of reporting an error, and shutting the entire thing down
func (c *Client) errorTree() {
	//we still need to shutdown the listeners anyway
	//at least write will probably not error out

	//since you're reading the text of this fn, prepare your face for some exposition on how this mechanism works
	//so, we've spread reading and writing to the conn (or whatever) across goroutines, this is great
	//high speed low drag
	//but what happens if one goroutine encounters an error? the goroutines don't know about each other, so what do we do?
	//well, writing, and reading from to a closed connection is an error condition. so if one goroutine dies, then the other
	//will be taken down with it
	//we also use this mechanism for a regular shutdown of the client's connection, simply crashing them both and releasing the resources
	e := <-c.internalErrorBuffer
	if e.reciever != _CON_READER {
		c.shutdown_reader <- struct{}{}
	}
	if e.reciever != _CON_WRITER {
		c.shutdown_writer <- struct{}{}
	}
	if c.C != nil {
		c.C.Close()
	}
	if e.reciever != _REGULAR_SHUTDOWN {
		c.ClientErrorBuffer <- fmt.Errorf("Shutting down: Recieved error %v\n", e.err.Error())
	}
}

//Shutdown sends a disconnect packet (if asked), and then disconnects from the broker after a set time limit
func (c *Client) Shutdown(sendDisconnect bool) error {
	var err error
	if sendDisconnect {
		e := SendDisconnect(c)
		if e != nil {
			//don't return here, wait to finish the flow
			err = errors.New("While sending disconnect: " + e.Error() + "\nNote:connection was shut down anyway")
		}
		<-time.After(time.Second)
	}
	c.internalErrorBuffer <- &errWrap{reciever: _REGULAR_SHUTDOWN}
	return err
}

func (c *Client) randoPerm(i int) []int {
	c.randomut.RLock()
	order := c.rando.Perm(i)
	c.randomut.RUnlock()
	return order
}

func (c *Client) getInt() int {
	c.randomut.RLock()
	num := c.rando.Int()
	c.randomut.RUnlock()
	return num
}
