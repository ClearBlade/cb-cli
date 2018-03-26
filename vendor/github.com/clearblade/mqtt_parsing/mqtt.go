package mqtt_parsing

import (
	"bytes"
	"fmt"
	"io"
	//	"log"
)

//Message is the interface all our packets will be implementing
type Message interface {
	Encode() []byte
	Type() uint8
}

/*
How feasible is it to do a no-copy version of this. So many inward ptrs later on.
*/

const (
	CONNECT = uint8(iota + 1)
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

var (
	_SIZE_LIMIT int64 = 0x100000000
)

//SetMaxPacketSize allows us to limit the size of an incoming packet below the mqtt minimum 256MB
func SetMaxPacketSize(lim int64) {
	_SIZE_LIMIT = lim
}

//StaticHeader as defined in http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html#fixed-header
//Though it's fixed header there
type StaticHeader struct {
	DUP    bool
	Retain bool
	QOS    uint8
}

///Connect is a struct for the packet of the same name
type Connect struct {
	ProtoName      string
	Version        uint8
	UsernameFlag   bool
	PasswordFlag   bool
	WillRetainFlag bool
	WillQOS        uint8
	WillFlag       bool
	CleanSeshFlag  bool
	KeepAlive      uint16

	ClientId    string
	WillTopic   TopicPath
	WillMessage string
	Username    string
	Password    string
}

//Connack struct, the return codes are as follows
// 0x00 connection accepted
// 0x01 refused: unacceptable proto version
// 0x02 refused: identifier rejected
// 0x03 refused server unavailiable
// 0x04 bad user or password
// 0x05 not authorized
type Connack struct {
	ReturnCode uint8
}

//Pubit
type Publish struct {
	Header    *StaticHeader
	Topic     TopicPath
	MessageId uint16
	Payload   []byte
}

//Puback is sent for QOS level one to verify the reciept of a publish
//Qoth the spec: "A PUBACK message is sent by a server in response to a PUBLISH message from a publishing client, and by a subscriber in response to a PUBLISH message from the server."
type Puback struct {
	MessageId uint16
}

//Pubrec is for verifying the reciept of a publish
//Qoth the spec:"It is the second message of the QoS level 2 protocol flow. A PUBREC message is sent by the server in response to a PUBLISH message from a publishing client, or by a subscriber in response to a PUBLISH message from the server."
type Pubrec struct {
	MessageId uint16
}

//Pubrelis third, a response to pubrec from either the client or server.
type Pubrel struct {
	MessageId uint16
	//QOS1
	Header *StaticHeader
}

//Pubcomp is for saying is in response to a pubrel sent by the publisher
//the final member of the QOS2 flow. both sides have said "hey, we did it!"
type Pubcomp struct {
	MessageId uint16
}

//Subscribe tells the server which topics the client would like to subscribe to
type Subscribe struct {
	Header        *StaticHeader
	MessageId     uint16
	Subscriptions []TopicQOSTuple
}

//Suback is to say "hey, you got it buddy. I will send you messages that fit this pattern"
type Suback struct {
	MessageId uint16
	Qos       []uint8
}

//Unsubscribe is the message to send if you don't want to subscribe to a topic anymore
type Unsubscribe struct {
	Header    *StaticHeader
	MessageId uint16
	Topics    []TopicQOSTuple
}

//Unsuback is to unsubscribe as suback is to subscribe
type Unsuback struct {
	MessageId uint16
}

//Pingreq is a keepalive
type Pingreq struct {
}

//Pingresp is for saying "hey, the server is alive"
type Pingresp struct {
}

//Disconnect is to signal you want to cease communications with the server
type Disconnect struct {
}

//TopicQOSTuple is a struct for pairing the Qos and topic together
//for the QOS' pairs in unsubscribe and subscribe
type TopicQOSTuple struct {
	Qos   uint8
	Topic TopicPath
}

//DecodePacket allows us to
func DecodePacket(rdr io.Reader) (Message, error) {
	hdr, sizeOf, messageType, err := decodeStaticHeader(rdr)

	if err != nil {
		return nil, err
	}
	//don't bother doing extra work
	//if there's no extra work to be done
	switch messageType {
	case PINGREQ:
		return &Pingreq{}, nil
	case PINGRESP:
		return &Pingresp{}, nil
	case DISCONNECT:
		return &Disconnect{}, nil
	}

	//check to make sure packet isn't above size limit
	if int64(sizeOf) > _SIZE_LIMIT {
		return nil, fmt.Errorf("packet too large, %d byte maximum.", _SIZE_LIMIT)
	}

	//now we're going to allocate a buffer to hold the rest of the packet
	//TODO: see if there is a way to use sync.Pool that isn't insane.
	//if only we had something akin to rust's Drop trait
	//	log.Printf("gonna decode %d bytes", sizeOf)
	motherBuffer := make([]byte, sizeOf)
	_, err = io.ReadFull(rdr, motherBuffer)
	//	log.Printf("we decided to do %d instead", n)
	if err != nil {

		return nil, err
	}

	//the meat, we're now decoding the body of the packet
	var msg Message
	switch messageType {
	case CONNECT:
		msg = decodeConnect(motherBuffer, hdr)
	case CONNACK:
		msg = decodeConnack(motherBuffer, hdr)
	case PUBLISH:
		msg = decodePublish(motherBuffer, hdr)
	case PUBACK:
		msg = decodePuback(motherBuffer, hdr)
	case PUBREC:
		msg = decodePubrec(motherBuffer, hdr)
	case PUBREL:
		msg = decodePubrel(motherBuffer, hdr)
	case PUBCOMP:
		msg = decodePubcomp(motherBuffer, hdr)
	case SUBSCRIBE:
		msg = decodeSubscribe(motherBuffer, hdr)
	case SUBACK:
		msg = decodeSuback(motherBuffer, hdr)
	case UNSUBSCRIBE:
		msg = decodeUnsubscribe(motherBuffer, hdr)
	case UNSUBACK:
		msg = decodeUnsuback(motherBuffer, hdr)
	default:
		return nil, fmt.Errorf("Have recievd an invalid zero-length packet positing it is type %d", messageType)
	}

	return msg, nil
}

/***** ENCODERS AND TYPE ACCESSORS LIE BELOW GODSPEED *******************/
//general encoding strategy works like this
//encode body, then encode the header and length, then cat the two
//buffers together. This is due to needing the entire packet length

//encodeParts sews the whole packet together
func encodeParts(msgType uint8, buf *bytes.Buffer, h *StaticHeader) []byte {
	var firstByte byte = 0
	firstByte |= msgType << 4
	if h != nil {
		firstByte |= boolToUInt8(h.DUP) << 3
		firstByte |= h.QOS << 1
		firstByte |= boolToUInt8(h.Retain)
	}
	//get the legnth first
	numBytes, bitField := encodeLength(uint32(buf.Len()) - 6)
	offset := 6 - numBytes - 1 //to account for the first byte
	byteBuf := buf.Bytes()
	//now we blit it in
	byteBuf[offset] = byte(firstByte)

	for t_off := offset + 1; t_off < 6; t_off++ {
		//coercing to byte selects the last 8 bits
		byteBuf[t_off] = byte(bitField >> ((numBytes - 1) * 8))
		numBytes--
	}
	//and return a slice from the offset
	return byteBuf[offset:]
}

//Encode for *Connect. now this actually implements the Message interface
func (c *Connect) Encode() []byte {
	buf := new(bytes.Buffer)

	//write some buffer space in the beginning for the maximum number of bytes static header + legnth encoding can take
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	//pack in the protonames and the flags, etc
	//the packer functions are in the bottom of the file
	packString(buf, c.ProtoName)
	buf.WriteByte(c.Version)
	var flagByte byte
	//decent explination in the decode function
	flagByte |= byte(boolToUInt8(c.UsernameFlag)) << 7
	flagByte |= byte(boolToUInt8(c.PasswordFlag)) << 6
	flagByte |= byte(boolToUInt8(c.WillRetainFlag)) << 5
	flagByte |= byte(c.WillQOS) << 3
	flagByte |= byte(boolToUInt8(c.WillFlag)) << 2
	flagByte |= byte(boolToUInt8(c.CleanSeshFlag)) << 1
	buf.WriteByte(flagByte)
	//keepalive and id
	packUInt16(buf, c.KeepAlive)
	packString(buf, c.ClientId)

	if c.WillFlag {
		packString(buf, c.WillTopic.Whole)
		packString(buf, c.WillMessage)
	}
	if c.UsernameFlag {
		packString(buf, c.Username)
	}
	if c.PasswordFlag {
		packString(buf, c.Password)
	}
	//we encode the header in a seperate buffer
	//the problem with this approach is that it requires two
	buffo := encodeParts(CONNECT, buf, nil)
	return buffo
}

//Type returns the type from the "enum"
func (c *Connect) Type() uint8 {
	return CONNECT
}

//Encode (ing)  connacks is relatively easy
func (c *Connack) Encode() []byte {
	buf := new(bytes.Buffer)
	//write padding
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	buf.WriteByte(byte(0))
	buf.WriteByte(byte(c.ReturnCode))
	return encodeParts(CONNACK, buf, nil)
}

//Type returns enum val
func (c *Connack) Type() uint8 {
	return CONNACK
}

//Encode writes the publish to a buffer
func (p *Publish) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packString(buf, p.Topic.Whole)
	if p.Header.QOS > 0 {
		packUInt16(buf, p.MessageId)
	}
	buf.Write(p.Payload)
	return encodeParts(PUBLISH, buf, p.Header)
}

//Type returns the enum value
func (p *Publish) Type() uint8 {
	return PUBLISH
}

//Encode encodes the puback, simple from here on out
func (p *Puback) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, p.MessageId)
	return encodeParts(PUBACK, buf, nil)
}

//Type returns the enum value
func (p *Puback) Type() uint8 {
	return PUBACK
}

//Encode encodes that sweet qos2 packet
func (p *Pubrec) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, p.MessageId)
	return encodeParts(PUBREC, buf, nil)
}

//Type returns the enum value
func (p *Pubrec) Type() uint8 {
	return PUBREC
}

//Encode encodes that pubrel sound you're so familiar with
func (p *Pubrel) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, p.MessageId)
	return encodeParts(PUBREL, buf, p.Header)
}

//Type returns the enum value
func (p *Pubrel) Type() uint8 {
	return PUBREL
}

//Encode encodes the final member of the QOS2 quartet
func (p *Pubcomp) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, p.MessageId)
	return encodeParts(PUBCOMP, buf, nil)
}

//Type returns the enum value
func (p *Pubcomp) Type() uint8 {
	return PUBCOMP
}

//Encode encodes the subscribe packet
func (s *Subscribe) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, s.MessageId)
	for _, t := range s.Subscriptions {
		top := t.Topic
		if top.Whole == "" {
			continue
		}
		packString(buf, top.Whole)
		buf.WriteByte(byte(t.Qos))

	}
	return encodeParts(SUBSCRIBE, buf, s.Header)
}

//Type returns the enum value of a subscribe
func (s *Subscribe) Type() uint8 {
	return SUBSCRIBE
}

//Encode suback encodes the suback for great good
func (s *Suback) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, s.MessageId)
	for _, q := range s.Qos {
		buf.WriteByte(byte(q))
	}
	return encodeParts(SUBACK, buf, nil)
}

//Type returns the enum value
func (s *Suback) Type() uint8 {
	return SUBACK
}

//Encode encodes the unsubscribe. All it is is the bodies of the subscribes
func (u *Unsubscribe) Encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, u.MessageId)
	for _, toptup := range u.Topics {
		packString(buf, toptup.Topic.Whole)
	}
	return encodeParts(UNSUBSCRIBE, buf, u.Header)
}

//Type returns the enum value
func (u *Unsubscribe) Type() uint8 {
	return UNSUBSCRIBE
}

//Unsuback merely acks the fact we recieved a message
func (u *Unsuback) Encode() []byte {
	//another deterministic one
	buf := new(bytes.Buffer)
	buf.Write([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	packUInt16(buf, u.MessageId)
	return encodeParts(UNSUBACK, buf, nil)
}

//Type returns the enum value
func (u *Unsuback) Type() uint8 {
	return UNSUBACK
}

//Encode for pingreq is simply the static header
func (p *Pingreq) Encode() []byte {
	//first nibble is the message id, second is nothing
	//another byte to signal no length
	return []byte{0xc0, 0x0}
}

//Type retursn the enum value for pingreq
func (p *Pingreq) Type() uint8 {
	return PINGREQ
}

//Encode returns the encoded pingresp packet
func (p *Pingresp) Encode() []byte {
	//same comment as pingreq
	return []byte{0xd0, 0x0}
}

//Type returns the enum value
func (p *Pingresp) Type() uint8 {
	return PINGRESP
}

//Encode for disconnect reutrns the encoded disconnect value
func (d *Disconnect) Encode() []byte {
	return []byte{0xe0, 0x0}
}

//Type returns the the enum value
func (d *Disconnect) Type() uint8 {
	return DISCONNECT
}

/***** DECODERS ARE GONNA LIVE HERE *****************************/

/**
decodeStaticHeader does what it says on the tin.
*/
func decodeStaticHeader(rdr io.Reader) (*StaticHeader, uint32, uint8, error) {
	var b [1]byte
	_, err := io.ReadFull(rdr, b[:])

	if err != nil {
		return nil, 0, 0, err
	}

	firstByte := b[0]
	messageType := (firstByte & 0xf0) >> 4
	var hdr *StaticHeader
	switch messageType {
	case PUBLISH, SUBSCRIBE, UNSUBSCRIBE, PUBREL:
		DUP := firstByte&0x08 > 0
		QOS := firstByte & 0x06 >> 1
		retain := firstByte&0x01 > 0

		hdr = &StaticHeader{
			DUP:    DUP,
			QOS:    QOS,
			Retain: retain,
		}

	}

	b[0] = 0x0 //b[0] ^= b[0]
	multiplier := uint32(1)
	length := uint32(0)
	digit := byte(0x80)
	//from the psudeocode in  the spec
	for (digit & 0x80) != 0 {
		_, err = io.ReadFull(rdr, b[:])

		if err != nil {
			return nil, 0, 0, err
		}
		digit = b[0]
		length += uint32(digit&0x7f) * multiplier
		multiplier *= 128
	}

	return hdr, uint32(length), messageType, nil
}

func decodeConnect(data []byte, hdr *StaticHeader) Message {
	//TODO: Decide how to recover rom invalid packets (offsets don't equal actual reading?)
	bookmark := uint32(0)

	protoname := getString(data, &bookmark)
	ver := uint8(data[bookmark])
	bookmark++
	flags := data[bookmark]
	bookmark++
	keepalive := getUInt16(data, &bookmark)
	cliID := getString(data, &bookmark)
	connect := &Connect{
		ProtoName: protoname,
		Version:   ver,
		KeepAlive: keepalive,
		ClientId:  cliID,
		//flagbyte is a byte that contains all the flags
		//we're going from left to right
		//LET THE GAMES BEGIN

		//username exists
		//* _ _ _ _ _ _ _
		UsernameFlag: flags&(1<<7) > 0,
		//password exists
		//_ * _ _ _ _ _ _
		PasswordFlag: flags&(1<<6) > 0,
		//this is for the lastwillandtestament meessage
		//this is asking for the retain flag,
		//_ _ * _ _ _ _ _
		WillRetainFlag: flags&(1<<5) > 0,
		//qos for lastwillandtestament
		//_ _ _ * * _ _ _
		WillQOS: (flags & (1 << 4)) + (flags & (1 << 3)),
		//actually to see if we're doing the last will and testament at all
		//_ _ _ _ _ * _ _
		WillFlag: flags&(1<<2) > 0,
		//is the session clean?
		//_ _ _ _ _ _ * _
		CleanSeshFlag: flags&(1<<1) > 0,
		//the last bit is reserved
	}
	if connect.WillFlag {
		connect.WillTopic, _ = NewTopicPath(getString(data, &bookmark))
		connect.WillMessage = getString(data, &bookmark)
	}
	if connect.UsernameFlag {
		connect.Username = getString(data, &bookmark)
	}
	if connect.PasswordFlag {
		connect.Password = getString(data, &bookmark)
	}
	return connect
}
func decodeConnack(data []byte, hdr *StaticHeader) Message {
	//first byte is weird in connack
	bookmark := uint32(1)
	retcode := data[bookmark]

	return &Connack{
		ReturnCode: retcode,
	}
}
func decodePublish(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	topic, _ := NewTopicPath(getString(data, &bookmark))
	var msgID uint16
	if hdr.QOS > 0 {
		msgID = getUInt16(data, &bookmark)
	}
	payload := data[bookmark:len(data)]
	return &Publish{
		Topic:     topic,
		Header:    hdr,
		Payload:   payload,
		MessageId: msgID,
	}
}
func decodePuback(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgId := getUInt16(data, &bookmark)
	return &Puback{
		MessageId: msgId,
	}
}
func decodePubrec(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := getUInt16(data, &bookmark)
	return &Pubrec{
		MessageId: msgID,
	}
}
func decodePubrel(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := getUInt16(data, &bookmark)
	return &Pubrel{
		Header:    hdr,
		MessageId: msgID,
	}
}
func decodePubcomp(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := getUInt16(data, &bookmark)
	return &Pubcomp{
		MessageId: msgID,
	}
}
func decodeSubscribe(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := getUInt16(data, &bookmark)
	var topics []TopicQOSTuple
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		topic := getString(data, &bookmark)
		qos := data[bookmark]
		bookmark++
		t.Topic, _ = NewTopicPath(topic)
		t.Qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Subscribe{
		Header:        hdr,
		MessageId:     msgID,
		Subscriptions: topics,
	}
}
func decodeSuback(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := getUInt16(data, &bookmark)
	var qoses []uint8
	maxlen := uint32(len(data))
	//is this efficent
	for bookmark < maxlen {
		qos := data[bookmark]
		bookmark++
		qoses = append(qoses, qos)
	}
	return &Suback{
		MessageId: msgID,
		Qos:       qoses,
	}
}
func decodeUnsubscribe(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	var topics []TopicQOSTuple
	msgID := getUInt16(data, &bookmark)
	maxlen := uint32(len(data))
	for bookmark < maxlen {
		var t TopicQOSTuple
		topic, _ := NewTopicPath(getString(data, &bookmark))
		//		qos := data[bookmark]
		//		bookmark++
		t.Topic = topic
		//		t.qos = uint8(qos)
		topics = append(topics, t)
	}
	return &Unsubscribe{
		Header:    hdr,
		MessageId: msgID,
		Topics:    topics,
	}
}
func decodeUnsuback(data []byte, hdr *StaticHeader) Message {
	bookmark := uint32(0)
	msgID := getUInt16(data, &bookmark)
	return &Unsuback{
		MessageId: msgID,
	}

}
func decodePingreq(data []byte, hdr *StaticHeader) Message {
	return &Pingreq{}
}
func decodePingresp(data []byte, hdr *StaticHeader) Message {
	return &Pingresp{}
}
func decodeDisconnect(data []byte, hdr *StaticHeader) Message {
	return &Disconnect{}
}

/*********** Helpers *********/
func packString(buf *bytes.Buffer, s string) {
	strlen := uint16(len(s))
	packUInt16(buf, strlen)
	buf.WriteString(s)
}

func packUInt16(buf *bytes.Buffer, tupac uint16) {
	buf.WriteByte(byte((tupac & 0xff00) >> 8))
	buf.WriteByte(byte(tupac & 0x00ff))
}

//TODO: investigate the effiency of passing pointer to slice rather than slice
//think we'd only be saving 1 ptr of space vs 3 ptrs
func getString(b []byte, startsAt *uint32) string {
	//	beginning := *startsAt
	l := getUInt16(b, startsAt)
	retval := string(b[*startsAt : uint32(l)+*startsAt]) //uint32(beginning)])
	*startsAt += uint32(l)
	return retval
}

func getUInt16(b []byte, startsAt *uint32) uint16 {
	fst := uint16(b[*startsAt])
	*startsAt++
	snd := uint16(b[*startsAt])
	*startsAt++
	return (fst << 8) + snd
}

//raisePanic raises a custom panic
//for usage with new messageHandler
func raisePanic(err error) {
	if err != nil {
		panic("error parsing packet")
	}
}

//boolToUint8 is a silly function to blit a boolean
//into a byte because go
func boolToUInt8(truth bool) uint8 {
	if truth {
		return 0x1
	} else {
		return 0x0
	}
}

//encodeLength encodes the length formatting
// http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html#fixed-header
//and tells us how many bytes it takes up
func encodeLength(bodyLength uint32) (uint8, uint32) {
	var bitField uint32 = 0
	var numBytes uint8 = 0

	if bodyLength == 0 {
		return 1, 0
	}
	for bodyLength > 0 {
		//I actually checked to see if the mod gets compiled away
		//and it does
		//hurrah
		bitField <<= 8
		//grab lowest 8 bits
		dig := uint8(bodyLength % 128)
		bodyLength /= 128
		if bodyLength > 0 {
			dig = dig | 0x80
		}
		bitField |= uint32(dig)
		numBytes++
	}
	return numBytes, bitField
}
