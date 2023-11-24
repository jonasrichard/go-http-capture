package main

import (
	"flag"
	"gohttpcapture/capture"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
    var iface string

    flag.StringVar(&iface, "i", "en0", "Name of the interface the data are captured")
    flag.Parse()

	if handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever); err != nil {
		panic(err)
	} else if err := handle.SetBPFFilter("tcp and port 80"); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

        capture.Capture(packetSource)
	}

}
