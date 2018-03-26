package mqttclient

import (
	mqtt "github.com/clearblade/mqtt_parsing"
	"sync"
)

//bog standard wrapped map

//storage is a map with a mutex. it's meant to store in flight messages
type storage struct {
	sto map[uint16]mqtt.Message
	mux *sync.RWMutex
}

//newStorage allocates a new one
func newStorage() *storage {
	return &storage{
		sto: make(map[uint16]mqtt.Message),
		mux: new(sync.RWMutex),
	}
}

//add adds a new (mid,message) tuple
func (s *storage) add(m mqtt.Message, messageid uint16) {
	s.mux.Lock()
	s.sto[messageid] = m
	s.mux.Unlock()
}

//check just checks for existence
func (s *storage) check(mid uint16) bool {
	s.mux.RLock()
	_, extant := s.sto[mid]
	s.mux.RUnlock()
	return extant
}

//removeMessageKeepId
func (s *storage) removeMessageKeepId(mid uint16) {
	s.mux.Lock()
	s.sto[mid] = nil
	s.mux.Unlock()
}

//getEntry
func (s *storage) getEntry(mid uint16) mqtt.Message {
	s.mux.Lock()
	msg, _ := s.sto[mid]
	s.mux.Unlock()
	return msg
}

//removeEntry
func (s *storage) removeEntry(mid uint16) {
	s.mux.Lock()
	delete(s.sto, mid)
	s.mux.Unlock()
}
