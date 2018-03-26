package mqttclient

import (
	"errors"
	mqtt "github.com/clearblade/mqtt_parsing"
	"strings"
	"sync"
)

//subscription_store is another simple threadsafe map
//however this is more specific
//I guess classical oo is missed here, but I probably could've designed this in a better way?
type subscription_store struct {
	//the channels in the values are outgoing channels
	store map[string]chan mqtt.Message
	mux   *sync.RWMutex
}

//newSubscriptionStore allocates a new sub store
func newSubscriptionStore() *subscription_store {
	return &subscription_store{
		store: make(map[string]chan mqtt.Message),
		mux:   new(sync.RWMutex),
	}
}

//relay_message takes a topic and shovels it along to the consumer-owned channel
func (ss *subscription_store) relay_message(msg mqtt.Message, topic string) error {
	topic = strip(topic)
	//drop the lock down if ordering is important
	ss.mux.RLock()
	ch, extant := ss.store[topic]
	ss.mux.RUnlock()
	if !extant {
		//this will probably be an error that is the client's fault
		return errors.New("Recieved message without subscription on topic " + topic)
	} else {
		ch <- msg
		return nil
	}

}

//new_subscription puts a new topic into the map and supplies a new channel
func (ss *subscription_store) new_subscription(topic string) (chan mqtt.Message, error) {
	topic = strip(topic)
	if ss.subscription_exists(topic) {
		return nil, errors.New("Duplicate subscription")
	} else {
		mch := make(chan mqtt.Message, 5)
		ss.mux.Lock()
		ss.store[topic] = mch
		ss.mux.Unlock()
		return mch, nil
	}
}

//subscription_exists checks for membership
func (ss *subscription_store) subscription_exists(topic string) bool {
	topic = strip(topic)
	ss.mux.RLock()
	_, extant := ss.store[topic]
	ss.mux.RUnlock()
	return extant
}

//remove_subscription kills the subscription and then kills the channel
func (ss *subscription_store) remove_subscription(topic string) {
	//the implications of this kind of suck for the reciever if this gets called
	//when they're trying to call the fn
	//could that be a cause of panics later on?
	//not with the `for msg := range mch` idiom
	topic = strip(topic)
	if ss.subscription_exists(topic) {
		ss.mux.Lock()
		mch, _ := ss.store[topic]
		delete(ss.store, topic)
		ss.mux.Unlock()
		close(mch)
	}

}

//strip removes the unsubscribe beacon from the topic
func strip(top string) string {
	//TODO:HACK:replace with a more solid manner of deliniating subscribes from unsubscribes
	if strings.Contains(top, "UNSUBSCRIBE#") {
		tops := strings.Split(top, "UNSUBSCRIBE#")
		if len(tops) > 1 {
			return "UNSUBSCRIBE#" + strings.Trim(tops[1], "/")
		}
	}
	return strings.Trim(top, "/")
}
