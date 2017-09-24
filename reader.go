package lester

import (
	"bytes"
	"errors"
	"os/exec"
)

var (
	NoMessage = errors.New("No message Received")
)

type CommandRunner interface {
	Run(input []byte) ([]byte, error)
}

type wrapexec struct{}

func (w wrapexec) Run(input []byte) ([]byte, error) {
	c := exec.Command("keybase", "chat", "api")
	c.Stdin = bytes.NewBuffer(input)
	return c.CombinedOutput()
}

type Message struct {
	Destination string
	Body        string
	Source      string
	Attachments []string
}

type Attachment struct {
	ID          uint64
	Size        uint64
	ContentType string
}
