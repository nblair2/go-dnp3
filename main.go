package main

import (
	"encoding/json"
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
	i := 0
	for pkt := range pcap.Packets() {

		i += 1
		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {

			tcp, _ := tcpLayer.(*layers.TCP)
			data := tcp.Payload
			if len(data) >= 10 {

				fmt.Printf("Packet: %d\n", i)
				var d dnp3.DNP3
				err := d.FromBytes(data)
				if err != nil {
					fmt.Printf("error FromBytes: %v\n", err)
				}

				b, err := d.ToBytes()
				if err != nil {
					fmt.Printf("error ToBytes: %v\n", err)
					continue
				}
				if !slices.Equal(b, data) {
					fmt.Println("error packet did not match")

					continue
				}
				fmt.Printf("Packet raw data:     0x % X\n", data)
				fmt.Printf("Packet dnp3.ToBytes: 0x % X\n", b)
				fmt.Printf("Packet in string format:\n")
				fmt.Println(d.String())
				pretty, err := json.MarshalIndent(d, "", "  ")
				if err != nil {
					fmt.Printf("error: marshaling DNP3 to JSON: %v\n", err)
				} else {
					fmt.Println("Packet in JSON format:")
					fmt.Println(string(pretty))
				}
			}
		}
	}

}
