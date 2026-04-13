package live

import (
	"context"
	"fmt"
	"os"
)

// StreamSenderConfig tunes the initial WS sender behavior.
type StreamSenderConfig struct {
	FlushEveryChunk bool
}

// SendAudioFrames pushes replayed audio chunks into the WS transport. For the
// first WS integration, it optionally flushes after each chunk so that final
// word events remain easy to attribute and compare.
func SendAudioFrames(ctx context.Context, client *WSLiveClient, chunks <-chan AudioChunk, finalized chan<- AudioChunk, finalizedAck <-chan struct{}, cfg StreamSenderConfig) error {
	defer close(finalized)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case chunk, ok := <-chunks:
			if !ok {
				if err := client.Stop(); err != nil {
					return fmt.Errorf("send stop event: %w", err)
				}
				return nil
			}
			if err := client.SendAudio(chunk); err != nil {
				return fmt.Errorf("send audio sequence %d: %w", chunk.Sequence, err)
			}
			if cfg.FlushEveryChunk {
				if err := client.Flush(); err != nil {
					return fmt.Errorf("flush sequence %d: %w", chunk.Sequence, err)
				}
				finalized <- chunk
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-finalizedAck:
				}
			}
			if chunk.WAVPath != "" {
				_ = os.Remove(chunk.WAVPath)
			}
		}
	}
}
