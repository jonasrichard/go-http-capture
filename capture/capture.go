package capture

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Conversation struct {
	StartTime time.Time
	// RequestID is IP << 16 + Port
	RequestID      uint32
	ResponseID     uint32
	RequestBuffer  *bytes.Buffer
	ResponseBuffer *bytes.Buffer
	HTTPRequest    *http.Request
	HTTPResponse   *http.Response
	RequestFIN     bool
	ResponseFIN    bool
}

// var flows map[uint32]*gopacket.Flow = make(map[uint32]*gopacket.Flow)
// var streams map[uint32]*bytes.Buffer = make(map[uint32]*bytes.Buffer)
// var requests map[uint32]*http.Request = make(map[uint32]*http.Request)
var conversations map[uint32]*Conversation = make(map[uint32]*Conversation)

// FlowKey packs a source and destination port to identify the flow
func FlowKey(srcPort, dstPort uint16) uint32 {
	return uint32(srcPort)<<16 + uint32(dstPort)
}

func Reverse(flowKey uint32) uint32 {
	lo := flowKey & 0xFFFF
	hi := flowKey >> 16

	return lo<<16 + hi
}

func (c *Conversation) AddPayload(flowKey uint32, payload []byte) {
	// if this is the first, this flow key is request id
	//if c.RequestID == 0 && c.ResponseID == 0 {
	//    c.StartTime = time.Now()
	//    c.RequestID = flowKey
	//    c.ResponseID = Reverse(flowKey)
	//    c.RequestBuffer = new(bytes.Buffer)
	//    c.ResponseBuffer = new(bytes.Buffer)
	//}

	if flowKey == c.RequestID {
		c.RequestBuffer.Write(payload)
	} else {
		c.ResponseBuffer.Write(payload)
	}
}

func (c *Conversation) HandleFIN(flowKey uint32) {
	if flowKey == c.RequestID {
		c.RequestFIN = true
	} else {
		c.ResponseFIN = true
	}
}

func (c *Conversation) Parse() error {
	var err error

	c.HTTPRequest, err = http.ReadRequest(bufio.NewReader(c.RequestBuffer))
	if err != nil {
		return err
	}

	if c.HTTPRequest == nil {
		return errors.New("no HTTP request parsed")
	}

	c.HTTPResponse, err = http.ReadResponse(bufio.NewReader(c.ResponseBuffer), c.HTTPRequest)
	if err != nil {
		return err
	}

	return nil
}

func Capture(packetSource *gopacket.PacketSource) {
	for packet := range packetSource.Packets() {

		//if layer := packet.Layer(layers.LayerTypeIPv4); layer != nil {
		//	ipLayer := layer.(*layers.IPv4)

		//} else {
		//	continue
		//}

		layer := packet.Layer(layers.LayerTypeTCP)
		if layer == nil {
			continue
		}

		tcpLayer := layer.(*layers.TCP)

		flowKey := FlowKey(uint16(tcpLayer.SrcPort), uint16(tcpLayer.DstPort))

		var conversation *Conversation
		var ok bool

		if conversation, ok = conversations[flowKey]; ok {
			conversation.AddPayload(flowKey, tcpLayer.Payload)
		} else {
			conversation = &Conversation{
				StartTime:      time.Now(),
				RequestID:      flowKey,
				ResponseID:     Reverse(flowKey),
				RequestBuffer:  new(bytes.Buffer),
				ResponseBuffer: new(bytes.Buffer),
			}

			conversations[flowKey] = conversation
			conversations[Reverse(flowKey)] = conversation
		}

		if tcpLayer.FIN {
			conversation.HandleFIN(flowKey)

			if conversation.RequestFIN && conversation.ResponseFIN {
				if err := conversation.Parse(); err == nil {
					fmt.Println(conversation.HTTPRequest)
					fmt.Println(conversation.HTTPResponse)

					// TODO delete conversations from the map
				} else {
					fmt.Println(err)
				}
			}
		}
	}
}
