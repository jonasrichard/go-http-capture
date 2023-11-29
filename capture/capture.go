package capture

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type CaptureConfig struct {
	Interface   string
	CaptureFile string
	Pattern     string
}

type Conversation struct {
	// StartTime is the timestamp of the first packet in the flow
	StartTime time.Time
	// RequestID is IP << 16 + port
	RequestID uint32
	// ResponseID is IP << 16 + port
	ResponseID uint32
	// RequestBuffer collects the request side of the traffic/payloads
	RequestBuffer *bytes.Buffer
	// ResponseBuffer collects the response side of the traffic/payloads
	ResponseBuffer *bytes.Buffer
	// HTTPRequest is the parsed HTTP request from the request buffer
	HTTPRequest *http.Request
	// HTTPResponse is the parsed HTTP response from the response buffer
	HTTPResponse *http.Response
	// RequestFIN is true if FIN packet arrived from request side
	RequestFIN bool
	// ResponseFIN is true if FIN packet arrived from response side
	ResponseFIN bool
}

var conversations map[uint32]*Conversation = make(map[uint32]*Conversation)

// FlowKey packs a source and destination port to identify the flow
func FlowKey(srcPort, dstPort uint16) uint32 {
	return uint32(srcPort)<<16 + uint32(dstPort)
}

// Reverse reverses the ports in the flowkey in order to find the other side
// of the conversation.
func Reverse(flowKey uint32) uint32 {
	lo := flowKey & 0xFFFF
	hi := flowKey >> 16

	return lo<<16 + hi
}

// AddPayload appends the payload to the respective buffer.
func (c *Conversation) AddPayload(flowKey uint32, payload []byte) {
	if flowKey == c.RequestID {
		c.RequestBuffer.Write(payload)
	} else {
		c.ResponseBuffer.Write(payload)
	}
}

// HandleFIN sets the respective FIN bool flag.
func (c *Conversation) HandleFIN(flowKey uint32) {
	if flowKey == c.RequestID {
		c.RequestFIN = true
	} else {
		c.ResponseFIN = true
	}
}

// Parse parses the request and response and stores them in this struct.
func (c *Conversation) Parse() error {
	var err error

	if c.RequestBuffer.Len() == 0 {
		return errors.New("empty request")
	}

	c.HTTPRequest, err = http.ReadRequest(bufio.NewReader(c.RequestBuffer))
	if err != nil {
		return err
	}

	if c.HTTPRequest == nil {
		return errors.New("no HTTP request parsed")
	}

	if c.ResponseBuffer.Len() == 0 {
		return errors.New("empty response")
	}

	c.HTTPResponse, err = http.ReadResponse(bufio.NewReader(c.ResponseBuffer), c.HTTPRequest)
	if err != nil {
		return err
	}

	return nil
}

// Capture starts packet capturing and collects the payloads and stores them in
// different Conversation values.
func Capture(config *CaptureConfig, packetSource *gopacket.PacketSource) {
	var r *regexp.Regexp
	var err error

	if config.Pattern != "" {
		r, err = regexp.Compile(config.Pattern)
		if err != nil {
			panic(err)
		}
	}

    fmt.Printf("%#v\n", r)

	count := 0

	for packet := range packetSource.Packets() {
		layer := packet.Layer(layers.LayerTypeTCP)
		if layer == nil {
			continue
		}

		tcpLayer := layer.(*layers.TCP)

		flowKey := FlowKey(uint16(tcpLayer.SrcPort), uint16(tcpLayer.DstPort))
		reverseFlowKey := Reverse(flowKey)

		var conversation *Conversation
		var ok bool

		if conversation, ok = conversations[flowKey]; ok {
			conversation.AddPayload(flowKey, tcpLayer.Payload)
		} else {
			conversation = &Conversation{
				StartTime:      time.Now(),
				RequestID:      flowKey,
				ResponseID:     reverseFlowKey,
				RequestBuffer:  new(bytes.Buffer),
				ResponseBuffer: new(bytes.Buffer),
			}

			conversations[flowKey] = conversation
			conversations[reverseFlowKey] = conversation
		}

		if tcpLayer.FIN {
			conversation.HandleFIN(flowKey)

			if conversation.RequestFIN && conversation.ResponseFIN {
				if err := conversation.Parse(); err == nil {
					if r != nil && r.Match([]byte(conversation.HTTPRequest.URL.Path)) {
						fmt.Println(conversation.HTTPRequest)
						fmt.Println(conversation.HTTPResponse)
					}

					delete(conversations, flowKey)
					delete(conversations, reverseFlowKey)
				} else {
					//fmt.Fprintf(os.Stderr, "Error %d Request:\n%s\nResponse:\n%s\n",
					//    count,
					//    conversation.RequestBuffer.String(),
					//    conversation.ResponseBuffer.String())
				}
			}
		}

		count++
	}
}
