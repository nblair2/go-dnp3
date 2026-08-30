package corpus_test

import (
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"
	"github.com/nblair2/go-dnp3/v4/dnp3"
)

const (
	corpusDir      = "corpus"
	scoreboardPath = "README.md"
)

var (
	// updateScoreboard regenerates the committed corpus scoreboard instead of
	// comparing against it.
	updateScoreboard = flag.Bool(
		"update-scoreboard",
		false,
		"rewrite "+scoreboardPath+" from the current parse results",
	)

	// customPcaps round-trips ad-hoc pcap files without scoreboard comparison.
	customPcaps = flag.String(
		"pcaps",
		"",
		"comma-separated list of extra pcap files to round-trip",
	)
)

// corpusStats are the per-pcap parse results tracked by the scoreboard.
type corpusStats struct {
	Payloads  int // non-empty TCP/UDP payloads
	Frames    int // DNP3 frames parsed
	Fragments int // application fragments reassembled
	Dropped   int // transport segments the assembler could not use
	Undecoded int // payloads ParseFrames could not fully consume
}

func (stats corpusStats) String() string {
	return fmt.Sprintf("payloads=%d frames=%d fragments=%d dropped=%d undecoded=%d",
		stats.Payloads, stats.Frames, stats.Fragments, stats.Dropped, stats.Undecoded)
}

// Failed reports whether the parser dropped transport segments or left
// payloads undecoded.
func (stats corpusStats) Failed() bool {
	return stats.Dropped > 0 || stats.Undecoded > 0
}

// icon is the scoreboard status glyph for stats.
func (stats corpusStats) icon() string {
	if stats.Failed() {
		return "❌"
	}

	return "✅"
}

// TestCorpus round-trips every DNP3 frame in the pcaps fetched by
// `make corpus`. Re-encoding must be byte-exact for every parsed frame.
// Per-pcap parse statistics are compared against the committed scoreboard so
// a parsing-coverage regression fails, while frames using unsupported
// groups/variations (counted in undecoded) do not.
func TestCorpus(t *testing.T) {
	t.Parallel()

	pcapFiles, err := filepath.Glob(filepath.Join(corpusDir, "*.pcap"))
	if err != nil {
		t.Fatal(err)
	}

	if len(pcapFiles) == 0 {
		t.Skipf("no pcaps in %s; run `make corpus` to fetch them", corpusDir)
	}

	results := make(map[string]corpusStats, len(pcapFiles))

	for _, path := range pcapFiles {
		results[filepath.Base(path)] = roundTripPcap(t, path)
	}

	if *updateScoreboard {
		writeScoreboard(t, results)

		return
	}

	compareScoreboard(t, results)
}

// TestCustomPcaps round-trips any pcap files passed via -pcaps. They are not
// compared against the scoreboard.
func TestCustomPcaps(t *testing.T) {
	t.Parallel()

	if *customPcaps == "" {
		t.Skip("no pcap files passed via -pcaps")
	}

	for pcapFile := range strings.SplitSeq(*customPcaps, ",") {
		pcapFile = strings.TrimSpace(pcapFile)
		if pcapFile == "" {
			continue
		}

		t.Run(filepath.Base(pcapFile), func(t *testing.T) {
			t.Parallel()
			t.Log(roundTripPcap(t, pcapFile))
		})
	}
}

// roundTripPcap parses every TCP/UDP payload in the pcap with ParseFrames,
// re-encodes each frame, and requires the rebuilt bytes to match the
// original payload exactly.
func roundTripPcap(t *testing.T, path string) corpusStats {
	t.Helper()

	handle, err := pcap.OpenOffline(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	defer handle.Close()

	var stats corpusStats

	// Payload remainders are not carried across packets, so a capture that
	// splits a link frame over two TCP segments loses the segment and counts
	// toward Dropped.
	assembler := &dnp3.Assembler{}

	packetIndex := 0
	for pkt := range gopacket.NewPacketSource(handle, handle.LinkType()).Packets() {
		packetIndex++

		payload := transportPayload(pkt)
		if len(payload) == 0 {
			continue
		}

		stats.Payloads++

		frames, remainder, err := dnp3.ParseFrames(payload)
		if err != nil {
			stats.Undecoded++
		}

		var rebuilt []byte

		for _, frame := range frames {
			stats.Frames++

			fragment, assembleErr := assembler.Assemble(frame)

			// A fragment can complete even when its application layer does
			// not decode, so check it before the error.
			switch {
			case fragment != nil:
				stats.Fragments++
			case assembleErr != nil:
				stats.Dropped++
			}

			encoded := serializeFrame(t, frame)
			rebuilt = append(rebuilt, encoded...)
		}

		rebuilt = append(rebuilt, remainder...)
		if !slices.Equal(rebuilt, payload) {
			t.Errorf("%s packet %d: round-trip mismatch\n got: %x\nwant: %x",
				filepath.Base(path), packetIndex, rebuilt, payload)
		}
	}

	return stats
}

// transportPayload returns the TCP or UDP payload of a packet, or nil.
func transportPayload(pkt gopacket.Packet) []byte {
	if tcp, ok := pkt.Layer(layers.LayerTypeTCP).(*layers.TCP); ok {
		return tcp.Payload
	}

	if udp, ok := pkt.Layer(layers.LayerTypeUDP).(*layers.UDP); ok {
		return udp.Payload
	}

	return nil
}

// serializeFrame runs a frame through gopacket.SerializeLayers and returns
// the resulting bytes.
func serializeFrame(t *testing.T, frame *dnp3.Frame) []byte {
	t.Helper()

	buf := gopacket.NewSerializeBuffer()

	err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{}, frame)
	if err != nil {
		t.Fatal("SerializeLayers:", err)
	}

	return buf.Bytes()
}

func writeScoreboard(t *testing.T, results map[string]corpusStats) {
	t.Helper()

	nameWidth := len("**Total**")
	for name := range results {
		nameWidth = max(nameWidth, len(name))
	}

	var (
		total   corpusStats
		passing int
	)

	for _, stats := range results {
		total.Payloads += stats.Payloads
		total.Frames += stats.Frames
		total.Fragments += stats.Fragments
		total.Dropped += stats.Dropped
		total.Undecoded += stats.Undecoded

		if !stats.Failed() {
			passing++
		}
	}

	totalStatus := fmt.Sprintf("%d/%d (%d%%)", passing, len(results), passing*100/len(results))
	statusWidth := max(len("status"), len(totalStatus))

	var builder strings.Builder

	builder.WriteString("# DNP3 corpus scoreboard\n\n")
	builder.WriteString(
		"Per-pcap round-trip results from `TestCorpus`. Regenerate with " +
			"`go test ./test -run TestCorpus -args -update-scoreboard`.\n\n")
	builder.WriteString(
		"`fragments` are application fragments reassembled by `dnp3.Assembler`;\n" +
			"`dropped` are transport segments it could not use. `status` is ❌ when\n" +
			"either is nonzero for that pcap.\n\n")
	fmt.Fprintf(&builder, "| %-*s | payloads | frames | fragments | dropped | undecoded | %-*s |\n",
		nameWidth, "pcap", statusWidth, "status")
	fmt.Fprintf(&builder, "| %s | -------: | -----: | --------: | ------: | --------: | :%s: |\n",
		strings.Repeat("-", nameWidth), strings.Repeat("-", statusWidth-2))

	for _, name := range slices.Sorted(maps.Keys(results)) {
		stats := results[name]
		fmt.Fprintf(&builder, "| %-*s | %8d | %6d | %9d | %7d | %9d | %-*s |\n",
			nameWidth, name, stats.Payloads, stats.Frames,
			stats.Fragments, stats.Dropped, stats.Undecoded, statusWidth, stats.icon())
	}

	fmt.Fprintf(&builder, "| %-*s | %8d | %6d | %9d | %7d | %9d | %-*s |\n",
		nameWidth, "**Total**", total.Payloads, total.Frames,
		total.Fragments, total.Dropped, total.Undecoded, statusWidth, totalStatus)

	err := os.WriteFile(scoreboardPath, []byte(builder.String()), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func compareScoreboard(t *testing.T, results map[string]corpusStats) {
	t.Helper()

	content, err := os.ReadFile(scoreboardPath)
	if err != nil {
		t.Fatalf("%v; run `go test ./test -run TestCorpus -args -update-scoreboard`", err)
	}

	expected := make(map[string]corpusStats, len(results))

	for line := range strings.SplitSeq(string(content), "\n") {
		var (
			name   string
			stats  corpusStats
			status string
		)

		// Only per-pcap rows match: five numeric cells followed by a
		// space-free status token. The header, separator, prose lines, and
		// the Total row (whose status cell contains a space) do not.
		_, scanErr := fmt.Sscanf(line, "| %s | %d | %d | %d | %d | %d | %s |",
			&name, &stats.Payloads, &stats.Frames,
			&stats.Fragments, &stats.Dropped, &stats.Undecoded, &status)
		if scanErr != nil {
			continue
		}

		expected[name] = stats
	}

	for _, name := range slices.Sorted(maps.Keys(results)) {
		want, ok := expected[name]
		if !ok {
			t.Errorf("%s: missing from scoreboard (re-run with -update-scoreboard)", name)

			continue
		}

		if results[name] != want {
			t.Errorf("%s: parse results changed\n got: %s\nwant: %s",
				name, results[name], want)
		}
	}

	for name := range expected {
		if _, ok := results[name]; !ok {
			t.Errorf("%s: in scoreboard but not fetched (re-run `make corpus`)", name)
		}
	}
}
