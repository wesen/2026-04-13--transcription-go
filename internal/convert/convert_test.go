package convert

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func TestTo16kMono_AlreadyTarget(t *testing.T) {
	// Create a minimal 16kHz mono WAV file
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.wav")
	outputPath := filepath.Join(dir, "output.wav")

	// Write a tiny 16kHz mono file
	f, err := os.Create(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	enc := wav.NewEncoder(f, 16000, 16, 1, 1)
	silence := &audio.IntBuffer{
		Format: &audio.Format{NumChannels: 1, SampleRate: 16000},
		Data:   make([]int, 1600), // 100ms of silence
	}
	enc.Write(silence)
	enc.Close()
	f.Close()

	err = To16kMono(inputPath, outputPath)
	if err != nil {
		t.Fatalf("To16kMono failed: %v", err)
	}

	// Verify output
	outF, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer outF.Close()

	dec := wav.NewDecoder(outF)
	if !dec.IsValidFile() {
		t.Fatal("output is not a valid WAV")
	}
	if dec.SampleRate != 16000 {
		t.Errorf("expected sample rate 16000, got %d", dec.SampleRate)
	}
	if dec.NumChans != 1 {
		t.Errorf("expected 1 channel, got %d", dec.NumChans)
	}
}

func TestTo16kMono_Stereo48k(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.wav")
	outputPath := filepath.Join(dir, "output.wav")

	// Write a 48kHz stereo file (1 second)
	f, err := os.Create(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	enc := wav.NewEncoder(f, 48000, 16, 2, 1)
	data := make([]int, 48000*2) // 1 second stereo
	for i := 0; i < len(data); i++ {
		data[i] = i % 1000 // some non-zero data
	}
	stereoBuf := &audio.IntBuffer{
		Format: &audio.Format{NumChannels: 2, SampleRate: 48000},
		Data:   data,
	}
	enc.Write(stereoBuf)
	enc.Close()
	f.Close()

	err = To16kMono(inputPath, outputPath)
	if err != nil {
		t.Fatalf("To16kMono failed: %v", err)
	}

	outF, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer outF.Close()

	dec := wav.NewDecoder(outF)
	if !dec.IsValidFile() {
		t.Fatal("output is not a valid WAV")
	}
	if dec.SampleRate != 16000 {
		t.Errorf("expected sample rate 16000, got %d", dec.SampleRate)
	}
	if dec.NumChans != 1 {
		t.Errorf("expected 1 channel, got %d", dec.NumChans)
	}
}

func TestTo16kMono_InvalidFile(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "notawav.txt")
	outputPath := filepath.Join(dir, "output.wav")

	os.WriteFile(inputPath, []byte("not a wav file"), 0644)

	err := To16kMono(inputPath, outputPath)
	if err == nil {
		t.Fatal("expected error for non-WAV input")
	}
}
