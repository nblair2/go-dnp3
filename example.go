package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nblair2/go-dnp3/dnp3"
)

func main() {
	// DNP3 Application Response G2 V2 - 2 byte prefix, 1 byte value, 6 byte abs time
	input := []byte{
		0x05, 0x64, 0x2a, 0x44, 0x01, 0x00, 0x00, 0x04,
		0xe5, 0x79, 0xc1, 0xe2, 0x81, 0x90, 0x00, 0x02,
		0x02, 0x28, 0x03, 0x00, 0x00, 0x00, 0x81, 0xda,
		0x33, 0xd2, 0xdf, 0xe5, 0x64, 0x71, 0x01, 0x00,
		0x00, 0x01, 0xda, 0x33, 0xd2, 0x64, 0x71, 0x01,
		0xff, 0xff, 0x81, 0xdb, 0xdd, 0x14, 0x33, 0xd2,
		0x64, 0x71, 0x01, 0x38, 0x5d,
	}

	// Parse the input bytes into a DNP3 Frame
	frame := dnp3.Frame{}
	if err := frame.FromBytes(input); err != nil {
		log.Fatalf("Failed to parse frame: %v", err)
	}

	// Display with String() method
	fmt.Println("--- Before (String) ---")
	fmt.Println(frame.String())

	// Change data
	data := frame.Application.GetData()
	data.Objects[0].Points[0] = &dnp3.PointNBytesAbsTime{
		Prefix:       []byte{0x01, 0x02},
		Value:        []byte{0xFF},
		AbsoluteTime: time.Date(2010, time.July, 1, 0, 0, 0, 0, time.UTC),
	}
	frame.Application.SetData(data)

	// Convert back to bytes (forces CRC recalculation)
	output, err := frame.ToBytes()
	if err != nil {
		log.Fatalf("Failed to convert frame to bytes: %v", err)
	}

	// Display as JSON
	fmt.Println("--- After (json) ---")

	jsonOutput, _ := json.MarshalIndent(frame, "", "  ")
	fmt.Println(string(jsonOutput))

	// Display as bytes
	fmt.Println("--- Compare ([]byte) ---")
	fmt.Printf("INPUT : 0x % X\nOUTPUT: 0x % X\n", input, output)
}
