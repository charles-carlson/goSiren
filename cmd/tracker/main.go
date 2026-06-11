package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"ambulance-tracker/config"
	"ambulance-tracker/internal/audio"
	"ambulance-tracker/internal/detector"
)

func main() {
	capturer := audio.NewPortAudioCapturer(float64(config.SampleRate), config.ChunkSize, 1)
	det := detector.NewDetector()

	if err := capturer.Start(); err != nil {
		log.Fatalf("failed to start capturer: %v", err)
	}
	defer capturer.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case samples := <-capturer.Samples():
			detected, direction := det.Detect(samples)
			if detected {
				log.Printf("siren detected: %s", direction)
			}
		case <-sig:
			log.Println("shutting down")
			return
		}
	}
}