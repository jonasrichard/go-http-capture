package main

import (
	"flag"
	"fmt"
	"gohttpcapture/capture"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
    config := new(capture.CaptureConfig)

	flag.StringVar(&config.Interface, "i", "", "The interface from the data are captured")
	flag.StringVar(&config.CaptureFile, "f", "", "Read from capture file (tcpdump)")
	flag.StringVar(&config.Pattern, "p", "", "Capture paths matching to this regexp")
	flag.Parse()

	fmt.Printf("%#v\n", config)

	if config.Interface != "" {
		capture.Capture(config, NewInterfaceSource(config))
	} else if config.CaptureFile != "" {
		capture.Capture(config, NewFileSource(config))
	} else {
		fmt.Fprintln(os.Stderr, "Neither -i interface nor -f capture file specified")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func NewInterfaceSource(config *capture.CaptureConfig) *gopacket.PacketSource {
	if handle, err := pcap.OpenLive(config.Interface, 1600, true, pcap.BlockForever); err != nil {
		panic(err)
	} else if err := handle.SetBPFFilter("tcp and port 80"); err != nil {
		panic(err)
	} else {
		return gopacket.NewPacketSource(handle, handle.LinkType())
	}
}

func NewFileSource(config *capture.CaptureConfig) *gopacket.PacketSource {
	if handle, err := pcap.OpenOffline(config.CaptureFile); err != nil {
		panic(err)
	} else if err := handle.SetBPFFilter("tcp and port 80"); err != nil {
		panic(err)
	} else {
		return gopacket.NewPacketSource(handle, handle.LinkType())
	}
}
