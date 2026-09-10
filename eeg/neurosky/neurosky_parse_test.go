package neurosky

import (
	"testing"

	"go.uber.org/zap"
)

// asicEEGPowerPacket builds the payload half of an ASIC_EEG_POWER packet:
// the 0x83 code, a length byte, and eight 3-byte big-endian band values.
func asicEEGPowerPacket(bands [8]int) []byte {
	out := []byte{EEG_POWER, 24}
	for _, v := range bands {
		out = append(out, byte(v>>16), byte(v>>8), byte(v))
	}
	return out
}

func drain(events chan MindflexEvent) []MindflexEvent {
	close(events)
	var got []MindflexEvent
	for e := range events {
		got = append(got, e)
	}
	return got
}

func testNeurosky() *Neurosky {
	return &Neurosky{name: "test", log: zap.NewNop().Sugar()}
}

// The band payload is packet-relative. Reading it from the head of the buffer
// instead made delta the [pad, 0x83, 0x18] header - a constant 33560 - and
// shifted every real band one slot late, dropping high gamma off the end.
func TestParseMindflexEEGPowerBandAlignment(t *testing.T) {
	want := [8]int{100000, 90000, 80000, 70000, 60000, 50000, 40000, 30000}

	// A leading POOR_SIGNAL packet puts the band packet at a non-zero offset,
	// which is exactly the case the old absolute indexing got wrong.
	data := []byte{POOR_SIGNAL, 0}
	data = append(data, asicEEGPowerPacket(want)...)

	events := make(chan MindflexEvent, 16)
	testNeurosky().ParseMindflex(data, events)

	var power *MindflexEEGPower
	for _, e := range drain(events) {
		if e.Type == EEG_POWER {
			p := e.EEGPower
			power = &p
		}
	}
	if power == nil {
		t.Fatal("no EEG_POWER event emitted")
	}

	got := [8]int{
		power.Delta, power.Theta, power.Low_Alpha, power.High_Alpha,
		power.Low_Beta, power.High_Beta, power.Low_Gamma, power.High_Gamma,
	}
	names := []string{"delta", "theta", "lowAlpha", "highAlpha",
		"lowBeta", "highBeta", "lowGamma", "highGamma"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s = %d, want %d", names[i], got[i], want[i])
		}
	}

	// 33560 is 0x008318 - the header bytes. Seeing it means we regressed to
	// reading from the head of the buffer.
	if power.Delta == 33560 {
		t.Error("delta is the packet header (0x008318): band offset regressed")
	}
}

// A packet cut short mid-band must not panic the reader.
func TestParseMindflexEEGPowerTruncated(t *testing.T) {
	full := asicEEGPowerPacket([8]int{1, 2, 3, 4, 5, 6, 7, 8})
	for cut := 2; cut < len(full); cut++ {
		events := make(chan MindflexEvent, 16)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic on %d-byte packet: %v", cut, r)
				}
			}()
			testNeurosky().ParseMindflex(full[:cut], events)
		}()
	}
}
