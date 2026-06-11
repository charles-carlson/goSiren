package detector

import (
	"ambulance-tracker/config"
	"math"

	"github.com/mjibson/go-dsp/fft"
)

type Detector interface {
	Detect(samples []float32) (bool, config.Direction)
}

type SirenDetector struct {
	chunkSize            int
	sirenEnergyThreshold config.EnergyLevel
	confirmationChunks   int
	silenceChunks        int
	envelopeWindowSecs   int
	confirmationCounter  int
	silenceCounter       int
	energyEnvelope       []float64
}

func NewDetector() *SirenDetector {
	return &SirenDetector{
		chunkSize:            config.ChunkSize,
		sirenEnergyThreshold: config.SirenEnergyThreshold,
		confirmationChunks:   config.ConfirmationChunks,
		silenceChunks:        config.SilenceChunks,
		envelopeWindowSecs:   config.EnvelopeWindowSecs,
		energyEnvelope:       make([]float64, 0, config.EnvelopeWindowSecs*int(config.SampleRate)/config.ChunkSize),
	}
}

func (d *SirenDetector) Detect(samples []float32) (bool, config.Direction) {
	f64 := make([]float64, len(samples))
	for i, v := range samples {
		f64[i] = float64(v)
	}
	bins := fft.FFTReal(f64)
	lowBin := binIndex(float64(config.SirenFreqLow), float64(config.SampleRate), len(samples))
	highBin := binIndex(float64(config.SirenFreqHigh), float64(config.SampleRate), len(samples))
	highBin = min(highBin, len(samples)/2-1)

	energy := bandEnergy(bins, lowBin, highBin)
	if config.EnergyLevel(energy) > d.sirenEnergyThreshold {
		d.confirmationCounter++
		d.silenceCounter = 0
		d.energyEnvelope = append(d.energyEnvelope, energy)
		if len(d.energyEnvelope) > config.EnvelopeWindowSecs*2 {
			d.energyEnvelope = d.energyEnvelope[1:]
		}
	} else {
		d.silenceCounter++
		if d.silenceCounter > d.silenceChunks {
			d.confirmationCounter = 0
			d.energyEnvelope = d.energyEnvelope[:0]
		}
	}
	if d.confirmationCounter >= config.ConfirmationChunks {
		d.confirmationCounter = 0
		return true, d.determineDirection()
	}
	return false, config.DirectionUnknown
}

// bin index = frequnecy * N / sample rate
func binIndex(freqHz float64, sampleRate float64, numSamples int) int {
	return int(freqHz * float64(numSamples) / sampleRate)
}

func bandEnergy(bins []complex128, lowBin int, highBin int) float64 {
	var energy float64
	for i := lowBin; i <= highBin; i++ {
		real := real(bins[i])
		imag := imag(bins[i])
		energy += math.Sqrt(real*real + imag*imag)
	}
	return energy
}

func (d *SirenDetector) determineDirection() config.Direction {
	n := len(d.energyEnvelope)
	if n < 2 {
		return config.DirectionUnknown
	}
	mid := n / 2
	var firstHalfEnergy, secondHalfEnergy float64
	for _, v := range d.energyEnvelope[:mid] {
		firstHalfEnergy += v
	}
	for _, v := range d.energyEnvelope[mid:] {
		secondHalfEnergy += v
	}
	diff := secondHalfEnergy - firstHalfEnergy
	switch {
	case diff > 0.5:
		return config.DirectionApproaching
	case diff < -0.5:
		return config.DirectionReceding
	default:
		return config.DirectionStationary
	}
}
