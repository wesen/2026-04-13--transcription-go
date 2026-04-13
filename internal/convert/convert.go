// Package convert provides pure Go audio conversion from WAV to 16kHz mono.
package convert

import (
	"fmt"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/oov/audio/resampler"
)

// To16kMono reads a WAV file, resamples to 16kHz mono PCM16, writes output.
// Streams in ~100ms chunks — constant memory regardless of input size.
func To16kMono(inputPath, outputPath string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		return fmt.Errorf("not a valid WAV file: %s", inputPath)
	}
	if err := dec.FwdToPCM(); err != nil {
		return fmt.Errorf("seek to PCM data: %w", err)
	}

	inRate := int(dec.SampleRate)
	numChans := int(dec.NumChans)
	bitDepth := int(dec.BitDepth)

	if inRate == 16000 && numChans == 1 {
		// Already in target format — just copy
		f.Close()
		return copyFile(inputPath, outputPath)
	}

	const targetRate = 16000

	outF, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer outF.Close()

	enc := wav.NewEncoder(outF, targetRate, 16, 1, 1)
	rs := resampler.New(1, inRate, targetRate, 0)
	maxVal := math.Pow(2, float64(bitDepth-1)) - 1

	// Process in ~100ms chunks for streaming
	framesPerChunk := inRate / 10
	bufSize := framesPerChunk * numChans

	for {
		buf := &audio.IntBuffer{
			Format: &audio.Format{NumChannels: numChans, SampleRate: inRate},
			Data:   make([]int, bufSize),
		}
		n, err := dec.PCMBuffer(buf)
		if err != nil {
			return fmt.Errorf("read PCM: %w", err)
		}
		if n == 0 {
			break
		}
		buf.Data = buf.Data[:n]

		// Downmix to mono (average channels)
		framesRead := n / numChans
		mono := make([]float64, framesRead)
		for i := 0; i < framesRead; i++ {
			sum := 0.0
			for c := 0; c < numChans; c++ {
				sum += float64(buf.Data[i*numChans+c])
			}
			mono[i] = sum / float64(numChans)
		}

		// Resample
		outSize := int(float64(len(mono))*float64(targetRate)/float64(inRate)) + 256
		outChunk := make([]float64, outSize)
		_, written := rs.ProcessFloat64(0, mono, outChunk)
		if written == 0 {
			continue
		}

		// Convert to int and write
		intOut := make([]int, written)
		for i, v := range outChunk[:written] {
			intOut[i] = int(math.Max(math.Min(v, maxVal), -maxVal-1))
		}

		outBuf := &audio.IntBuffer{
			Data:   intOut,
			Format: &audio.Format{NumChannels: 1, SampleRate: targetRate},
		}
		if err := enc.Write(outBuf); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}

	if err := enc.Close(); err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
