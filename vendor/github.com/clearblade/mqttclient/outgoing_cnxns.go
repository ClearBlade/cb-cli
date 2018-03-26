package mqttclient

import (
	"errors"
	mqtt "github.com/clearblade/mqtt_parsing"
	"sync"
)

//patnote: I understand that right tool for the right job, but
//generics would be a handy thing to have at this moment
//we have to do this so that we can make outgoing channels of type
//*mqtt.publish, so that we can then hoist the type out in importing packages
//so someone consuming say, they go-sdk doesn't have to import mqtt_parsing
//as well just to make use of the package
type outgoing_topics struct {
	store map[string]chan *mqtt.Publish
	mux   *sync.RWMutex
}

func newOutgoingTopics() *outgoing_topics {
	return &outgoing_topics{
		store: make(map[string]chan *mqtt.Publish),
		mux:   new(sync.RWMutex),
	}
}

func (ot *outgoing_topics) relay_message(msg *mqtt.Publish, topic string) error {
	topic = strip(topic)
	//drop the lock down if ordering is important
	ot.mux.RLock()
	//ch, extant := ot.store[topic]
	ch, extant := ot.store[ot.bestTopicMatch(topic)]
	ot.mux.RUnlock()
	if !extant {
		//this will probably be an error that is the client's fault
		return errors.New("Recieved message without subscription on topic " + topic)
	} else {
		ch <- msg
		return nil
	}
}

func (ot *outgoing_topics) new_outgoing(topic string) (chan *mqtt.Publish, error) {
	topic = strip(topic)
	if ot.subscription_exists(topic) {
		return nil, errors.New("Duplicate, will not overwrite")
	} else {
		mch := make(chan *mqtt.Publish, 5)
		ot.mux.Lock()
		ot.store[topic] = mch
		ot.mux.Unlock()
		return mch, nil
	}
}

func (ot *outgoing_topics) subscription_exists(topic string) bool {
	topic = strip(topic)
	ot.mux.RLock()
	_, extant := ot.store[topic]
	ot.mux.RUnlock()
	return extant
}

//remove_subscription kills the subscription and then kills the channel
func (ot *outgoing_topics) remove_subscription(topic string) {
	topic = strip(topic)
	if ot.subscription_exists(topic) {
		ot.mux.Lock()
		mch, _ := ot.store[topic]
		delete(ot.store, topic)
		ot.mux.Unlock()
		close(mch)
	}
}
