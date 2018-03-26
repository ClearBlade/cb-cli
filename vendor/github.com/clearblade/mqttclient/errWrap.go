package mqttclient

const (
	_CON_WRITER = iota
	_CON_READER
	_REGULAR_SHUTDOWN
	_OTHER
)

type errWrap struct {
	err      error
	reciever int
}
