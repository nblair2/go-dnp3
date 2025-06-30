package main

import (
	"fmt"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/nblair2/go-dnp3/dnp3"
)

func main() {
	handle, err := pcap.OpenOffline(os.Args[1])
	if err != nil {
		fmt.Println("Error opening PCAP:", err)
		return
	}
	defer handle.Close()

	pcap := gopacket.NewPacketSource(handle, handle.LinkType())
	for pkt := range pcap.Packets() {
		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			data := tcp.Payload
			if len(data) >= 10 {
				var d dnp3.DNP3
				err := d.DecodeFromBytes(data)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(d)
			}
		}
	}

}
