package streamtools

import (
	"github.com/bitly/go-simplejson"
	"github.com/bitly/nsq/nsq"
	"log"
)

var (
	lookupdHTTPAddrs = "127.0.0.1:4161"
	nsqdHTTPAddrs    = "127.0.0.1:4150"
	nsqdTCPAddrs     = "127.0.0.1:4150"
)

type SyncHandler struct {
	msgChan chan simplejson.Json
}

func (self *SyncHandler) HandleMessage(m *nsq.Message) error {
	blob, err := simplejson.NewJson(m.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	self.msgChan <- *blob
	return nil
}

func nsqReader(topic string, channel string, writeChan chan simplejson.Json) {
	r, err := nsq.NewReader(topic, channel)
	if err != nil {
		log.Fatal(err.Error())
	}
	sh := SyncHandler{
		msgChan: writeChan,
	}
	r.AddHandler(&sh)
	_ = r.ConnectToLookupd(lookupdHTTPAddrs)
	<-r.ExitChan
}

func nsqWriter(topic string, channel string, readChan chan simplejson.Json) {

	w := nsq.NewWriter(0)
	err := w.ConnectToNSQ(nsqdHTTPAddrs)
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		select {
		case msg := <-readChan:
			outMsg, _ := msg.Encode()
			frameType, data, err := w.Publish(topic, outMsg)
			if err != nil {
				log.Fatalf("frametype %d data %s error %s", frameType, string(data), err.Error())
			}
		}
	}
}