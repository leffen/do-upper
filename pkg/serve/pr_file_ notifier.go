package serve

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type pingResponseFileAppender struct {
	fileName string
	f        *os.File
}

func newPingResponseFileAppender(fileName string) *pingResponseFileAppender {
	return &pingResponseFileAppender{fileName: fileName}
}

func (p *pingResponseFileAppender) OnStart() error {
	f, err := os.OpenFile(p.fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("Unable to open %s  with error :%s", p.fileName, err)
	}
	p.f = f
	return nil
}

func (p *pingResponseFileAppender) OnEnd() error {
	return p.f.Close()
}

func (p *pingResponseFileAppender) OnNewMeasurement(resp pingResponse) error {
	data, err := resp.toJSON()
	if err != nil {
		return fmt.Errorf("Unable to marshal to json with error %s", err)
	}

	logrus.Debugf("RESP :%#v\n", resp)

	_, err = p.f.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("Unable to marshal to json with error %s", err)
	}
	p.f.Write([]byte("\n"))
	return nil
}
