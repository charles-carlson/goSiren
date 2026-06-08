package audio

import (
	"fmt"
	"sync"

	"github.com/gordonklaus/portaudio"
)

type Capturer interface {
	Start() error
	Stop() error
	Samples() <-chan []float32
}

type PortAudioCapturer struct {
	SampleRate      float64
	framesPerBuffer int
	channels        int

	stream      *portaudio.Stream
	samplesChan chan []float32
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
	isRunning   bool
}

func NewPortAudioCapturer(sampleRate float64, framesPerBuffer int, channels int) *PortAudioCapturer {
	return &PortAudioCapturer{
		SampleRate:      sampleRate,
		framesPerBuffer: framesPerBuffer,
		channels:        channels,
		samplesChan:     make(chan []float32, 10), // Buffered channel to prevent blocking
		stopChan:        make(chan struct{}),
	}
}

func (p *PortAudioCapturer) Samples() <-chan []float32 {
	return p.samplesChan
}

func (p *PortAudioCapturer) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isRunning {
		return fmt.Errorf("capturer is already running")
	}
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize portaudio: %w", err)
	}
	buffer := make([]float32, p.framesPerBuffer)
	stream, err := portaudio.OpenDefaultStream(
		p.channels,
		0, // Input only
		p.SampleRate,
		p.framesPerBuffer,
		&buffer,
	)
	if err != nil {
		portaudio.Terminate()
		return fmt.Errorf("failed to open stream: %w", err)
	}
	if err := stream.Start(); err != nil {
		stream.Close()
		portaudio.Terminate()
		return fmt.Errorf("failed to start stream: %w", err)
	}

	p.stream = stream
	p.isRunning = true
	p.stopChan = make(chan struct{})

	// Start background thread processing
	p.wg.Add(1)
	go p.captureLoop(buffer)

	return nil
}

func (p *PortAudioCapturer) captureLoop(buffer []float32) {
	defer p.wg.Done()
	for {
		select {
		case <-p.stopChan:
			return
		default:
			if err := p.stream.Read(); err != nil {
				return
			}
			// Send a copy of the buffer to avoid data race
			outBuffer := make([]float32, len(buffer))
			copy(outBuffer, buffer)
			p.samplesChan <- outBuffer
		}
	}
}

func (p *PortAudioCapturer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isRunning {
		return fmt.Errorf("capturer is not running")
	}
	close(p.stopChan)
	p.wg.Wait() // Wait for CaptureLoop to finish
	if err := p.stream.Stop(); err != nil {
		return fmt.Errorf("failed to stop stream: %w", err)
	}
	if err := p.stream.Close(); err != nil {
		return fmt.Errorf("failed to close stream: %w", err)
	}
	if err := portaudio.Terminate(); err != nil {
		return fmt.Errorf("failed to terminate portaudio: %w", err)
	}
	p.isRunning = false
	return nil
}
