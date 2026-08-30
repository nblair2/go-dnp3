package dnp3_test

import (
	"fmt"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

// ExampleAssembler reassembles an application fragment split across two link
// frames. In practice the frame stream comes from a TCP or serial capture;
// see test/stream_test.go for wiring into gopacket/tcpassembly.
func ExampleAssembler() {
	// Two link frames carrying one Read request split across two transport
	// segments (FIR SEQ 0, then FIN SEQ 1).
	wire := []byte{
		0x05, 0x64, 0x0a, 0xc4, 0x03, 0x00, 0x04, 0x00,
		0x08, 0xcf, 0x40, 0xc5, 0x01, 0x3c, 0x02, 0x6f,
		0xda, 0x05, 0x64, 0x10, 0xc4, 0x03, 0x00, 0x04,
		0x00, 0xa2, 0x0b, 0x81, 0x06, 0x3c, 0x03, 0x06,
		0x3c, 0x04, 0x06, 0x3c, 0x01, 0x06, 0x74, 0xa9,
	}

	frames, _, err := dnp3.ParseFrames(wire)
	if err != nil {
		panic(err)
	}

	var assembler dnp3.Assembler

	for _, frame := range frames {
		fragment, err := assembler.Assemble(frame)
		if err != nil || fragment == nil {
			continue
		}

		fmt.Printf("fragment: %s, %d segments, %d bytes, FC %d\n",
			fragment.Session,
			fragment.Segments,
			len(fragment.Data),
			fragment.Application.GetFunctionCode(),
		)
	}

	// Output:
	// fragment: 4 -> 3 (request), 2 segments, 14 bytes, FC 1
}
