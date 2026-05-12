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

	// ParseFrames handles raw TCP reads that may contain multiple DNP3 frames
	// or a trailing partial frame. It returns all complete frames and any
	// unconsumed bytes so the caller can prepend them to the next read.
	tcpSegment := append(input, input...) // two frames in one segment
	frames, remainder, err := dnp3.ParseFrames(tcpSegment)
	if err != nil {
		log.Fatalf("Failed to parse frames: %v", err)
	}

	fmt.Printf("--- ParseFrames: %d frame(s), %d remainder byte(s) ---\n", len(frames), len(remainder))

	for i, f := range frames {
		fmt.Printf("Frame %d:\n%s\n", i+1, f.String())
	}

	// Parse the input bytes into a single DNP3 Frame
	frame, err := dnp3.NewFrameFromBytes(input)
	if err != nil {
		log.Fatalf("Failed to parse frame: %v", err)
	}

	// Display with String() method
	fmt.Println("--- Before (String) ---")
	fmt.Println(frame.String())

	// Change data
	data := frame.Application.GetData()

	point := data.Objects[0].Points[0].(*dnp3.PointBytes)

	if err := point.SetIndex(0x0201); err != nil {
		log.Fatalf("Failed to set index: %v", err)
	}

	if err := point.SetValue([]byte{0xFF}); err != nil {
		log.Fatalf("Failed to set value: %v", err)
	}

	timestamp := time.Date(2010, time.July, 1, 0, 0, 0, 0, time.UTC)
	absTime := dnp3.AbsoluteTime(timestamp)

	if err := point.SetAbsTime(absTime); err != nil {
		log.Fatalf("Failed to set absolute time: %v", err)
	}

	data.Objects[0].Points[0] = point
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
