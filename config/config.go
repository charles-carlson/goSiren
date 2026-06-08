package config

// ---- Audio ----
const (
	SampleRate        = 44100
	ChunkDurationSecs = 0.5
	ChunkSize         = 22050
)

// --- Frequency ---
type FrequencyHz float64

const (
	SirenFreqLow  FrequencyHz = 700.0
	SirenFreqHigh FrequencyHz = 1700.0
)

// --- Detection thresholds ---
type EnergyLevel float64

const (
	SirenEnergyThreshold EnergyLevel = 0.15
	ConfirmationChunks   int         = 4
	SilenceChunks        int         = 6
	EnvelopeWindowSecs   int         = 20
)

// --- Direction (lives in detector package, not config) ---
type Direction uint8

const (
	DirectionUnknown     Direction = iota // 0
	DirectionApproaching                  // 1
	DirectionReceding                     // 2
	DirectionStationary                   // 3
)

// String makes Direction print nicely in logs and Sheets
func (d Direction) String() string {
	switch d {
	case DirectionApproaching:
		return "approaching"
	case DirectionReceding:
		return "receding"
	case DirectionStationary:
		return "stationary"
	default:
		return "unknown"
	}
}
