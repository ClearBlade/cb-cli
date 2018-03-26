package cbjson

import (
	"bytes"
)

type StateProcessor interface {
	Process(c byte, buf *bytes.Buffer) (bool, StateProcessor)
	Name() string
}

type Normal struct{}
type MightBeComment struct{}
type MightEndComment struct{}
type DoubleSlashComment struct{}
type SlashStarComment struct{}
type InString struct{}
type InStringEscape struct{}

type Preprocessor struct {
	inBuf    *bytes.Buffer
	outBuf   *bytes.Buffer
	curState StateProcessor
}

var (
	normal             = Normal{}
	mightBeComment     = MightBeComment{}
	mightEndComment    = MightEndComment{}
	doubleSlashComment = DoubleSlashComment{}
	slashStarComment   = SlashStarComment{}
	inString           = InString{}
	inStringEscape     = InStringEscape{}
)

func NewPreprocessor(byts []byte) *Preprocessor {
	return &Preprocessor{
		inBuf:    bytes.NewBuffer(byts),
		outBuf:   &bytes.Buffer{},
		curState: normal,
	}
}

func (p *Preprocessor) preprocess() (*bytes.Buffer, error) {
	var writeByte bool
	for {

		// Read the byte. Process EOF on successful scan or error
		// Note -- this is a bit obscure. We should check the error
		// but it is most likely EOF
		c, err := p.inBuf.ReadByte()
		if err != nil {
			if p.curState == normal {
				return p.outBuf, nil
			}
			return nil, err
		}

		// Process the byte we just read.
		if writeByte, p.curState = p.curState.Process(c, p.outBuf); writeByte {
			p.outBuf.WriteByte(c)
		}
	}
}

// The normal procsssing state. Everything is cool unless we see the
// potential beginning of a comment (/) or a string beginning (")
func (s Normal) Process(c byte, buf *bytes.Buffer) (bool, StateProcessor) {
	if c == '/' {
		return false, mightBeComment
	} else if c == '"' {
		return true, inString
	}
	return true, normal
}

func (s Normal) Name() string {
	return "Normal"
}

// We've seen a slash (/), is the next char a star (*) or another slash (/)
// If not, we have to write the original slash and the current char...
func (s MightBeComment) Process(c byte, buf *bytes.Buffer) (bool, StateProcessor) {
	if c == '/' {
		return false, doubleSlashComment
	} else if c == '*' {
		return false, slashStarComment
	}
	buf.WriteByte('/')
	return true, normal
}

func (s MightBeComment) Name() string {
	return "MightBeComment"
}

func (s MightEndComment) Process(c byte, buf *bytes.Buffer) (bool, StateProcessor) {
	if c == '/' {
		return false, normal
	}
	if c == '*' {
		return false, mightEndComment
	}
	return false, slashStarComment
}

func (s MightEndComment) Name() string {
	return "MightEndComment"
}

func (s DoubleSlashComment) Process(c byte, buf *bytes.Buffer) (bool, StateProcessor) {
	if c == '\n' {
		return true, normal
	}
	return false, doubleSlashComment
}

func (s DoubleSlashComment) Name() string {
	return "DoubleSlashComment"
}

func (s SlashStarComment) Process(c byte, buf *bytes.Buffer) (bool, StateProcessor) {
	if c == '*' {
		return false, mightEndComment
	}
	return false, slashStarComment
}

func (s SlashStarComment) Name() string {
	return "SlashStarComment"
}

func (s InString) Process(c byte, buf *bytes.Buffer) (bool, StateProcessor) {
	if c == '"' {
		return true, normal
	} else if c == '\\' {
		return true, inStringEscape
	}
	return true, inString
}

func (s InString) Name() string {
	return "InString"
}

func (s InStringEscape) Process(c byte, buf *bytes.Buffer) (bool, StateProcessor) {
	return true, inString
}

func (s InStringEscape) Name() string {
	return "InStringEscape"
}
