package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/everadaptive/mindlights/eeg/neurosky"
	"go.uber.org/zap"
)

type replayRow struct {
	timestamp  int64
	eegPower   neurosky.MindflexEEGPower
	attention  int
	meditation int
}

// replayCSV reads a mindlights CSV log and replays it onto the event channel,
// paced by the recorded timestamps divided by speed. start skips into the
// recording; limit > 0 truncates the replay to the first limit of recorded time.
func replayCSV(path string, source string, speed float64, start, limit time.Duration, log *zap.SugaredLogger) (chan neurosky.MindflexEvent, error) {
	rows, err := loadReplayCSV(path)
	if err != nil {
		return nil, err
	}
	if start > 0 {
		from := rows[0].timestamp + int64(start/time.Second)
		i := 0
		for ; i < len(rows) && rows[i].timestamp < from; i++ {
		}
		rows = rows[i:]
	}
	if limit > 0 {
		end := rows[0].timestamp + int64(limit/time.Second)
		trimmed := rows[:1]
		for _, r := range rows[1:] {
			if r.timestamp > end {
				break
			}
			trimmed = append(trimmed, r)
		}
		rows = trimmed
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("replay window contains no samples")
	}
	log.Infof("replaying %d samples from %s (speed %fx)", len(rows), path, speed)

	events := make(chan neurosky.MindflexEvent, 16)
	go func() {
		defer close(events)
		var prev int64
		for _, r := range rows {
			if prev != 0 {
				delay := time.Duration(float64(r.timestamp-prev) * float64(time.Second) / speed)
				if delay > 0 {
					time.Sleep(delay)
				}
			}
			prev = r.timestamp

			events <- neurosky.MindflexEvent{
				Type:          neurosky.POOR_SIGNAL,
				Source:        source,
				SignalQuality: 0,
			}
			events <- neurosky.MindflexEvent{
				Type:     neurosky.EEG_POWER,
				Source:   source,
				EEGPower: r.eegPower,
			}
			if r.attention >= 0 {
				events <- neurosky.MindflexEvent{
					Type:      neurosky.ATTENTION,
					Source:    source,
					Attention: r.attention,
				}
			}
			if r.meditation >= 0 {
				events <- neurosky.MindflexEvent{
					Type:       neurosky.MEDITATION,
					Source:     source,
					Meditation: r.meditation,
				}
			}
		}
		log.Info("replay finished")
	}()
	return events, nil
}

// legacyBandShift corrects the column->band mapping for CSVs recorded before the
// EEG_POWER parser offset bug was fixed.
//
// ParseMindflex indexed the band payload as data[k:k+3] instead of
// data[n+k:n+k+3] - absolute buffer offset rather than packet-relative. The
// "eegdelta" column is therefore not delta at all: it is the three header bytes
// [pad, 0x83, 0x18] = 33560, constant on every sample. Every real band landed in
// the next column along, and the true high gamma fell off the end and was never
// recorded.
//
// So a legacy CSV column maps to the band one position lower, and high gamma is
// simply unavailable. Set --legacy-band-shift=false for CSVs captured after the
// parser fix.
var legacyBandShift = true

// readBands pulls the eight band powers out of one CSV row, applying the legacy
// column shift when it is in effect.
func readBands(atoi func(string) int) neurosky.MindflexEEGPower {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		return v
	}
	if !legacyBandShift {
		return neurosky.MindflexEEGPower{
			Delta:      clamp(atoi("eegdelta")),
			Theta:      clamp(atoi("eegtheta")),
			Low_Alpha:  clamp(atoi("eeglowalpha")),
			High_Alpha: clamp(atoi("eeghighalpha")),
			Low_Beta:   clamp(atoi("eeglowbeta")),
			High_Beta:  clamp(atoi("eeghighbeta")),
			Low_Gamma:  clamp(atoi("eeglowgamma")),
			High_Gamma: clamp(atoi("eeghighgamma")),
		}
	}
	return neurosky.MindflexEEGPower{
		Delta:      clamp(atoi("eegtheta")),
		Theta:      clamp(atoi("eeglowalpha")),
		Low_Alpha:  clamp(atoi("eeghighalpha")),
		High_Alpha: clamp(atoi("eeglowbeta")),
		Low_Beta:   clamp(atoi("eeghighbeta")),
		High_Beta:  clamp(atoi("eeglowgamma")),
		Low_Gamma:  clamp(atoi("eeghighgamma")),
		High_Gamma: 0, // never made it into the log
	}
}

// loadReplayCSV parses the CSV, collapsing duplicate timestamps into a single
// sample and merging any attention/meditation values found on any of the rows.
func loadReplayCSV(path string) ([]replayRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	col := map[string]int{}
	for i, h := range header {
		col[h] = i
	}
	for _, need := range []string{"timestamp", "eegdelta", "eegtheta", "eeglowalpha", "eeghighalpha", "eeglowbeta", "eeghighbeta", "eeglowgamma", "eeghighgamma"} {
		if _, ok := col[need]; !ok {
			return nil, fmt.Errorf("csv missing column %q", need)
		}
	}

	var rows []replayRow
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		ts, err := strconv.ParseInt(rec[col["timestamp"]], 10, 64)
		if err != nil {
			continue
		}

		atoi := func(name string) int {
			idx, ok := col[name]
			if !ok || idx >= len(rec) || rec[idx] == "" {
				return -1
			}
			v, err := strconv.Atoi(rec[idx])
			if err != nil {
				return -1
			}
			return v
		}

		if n := len(rows); n > 0 && rows[n-1].timestamp == ts {
			// duplicate sample for the same second: merge att/med if present
			if a := atoi("attention"); a >= 0 {
				rows[n-1].attention = a
			}
			if m := atoi("meditation"); m >= 0 {
				rows[n-1].meditation = m
			}
			continue
		}

		rows = append(rows, replayRow{
			timestamp:  ts,
			eegPower:   readBands(atoi),
			attention:  atoi("attention"),
			meditation: atoi("meditation"),
		})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no replay samples found in %s", path)
	}
	return rows, nil
}
