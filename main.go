package main

import (
	"fmt"
	"os"
	"slices"

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
	i := 1
	for pkt := range pcap.Packets() {

		i += 1
		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {

			tcp, _ := tcpLayer.(*layers.TCP)
			data := tcp.Payload
			if len(data) >= 10 {

				var d dnp3.DNP3
				err := d.FromBytes(data)
				if err != nil {
					fmt.Println(err)
					continue
				}

				fmt.Printf("Packet: %d\n", i)

				bytes := d.ToBytes()
				if !slices.Equal(bytes, data) {
					fmt.Println("Packet did not match")
				}
				fmt.Printf("Packet raw data:     0x % X\n", data)
				fmt.Printf("Packet dnp3.ToBytes: 0x % X\n", bytes)
				fmt.Println(d.String())
			}
		}
	}

}
