package bench

import "time"
import "sync"

// TokenStream generates a stream of tokens
type TokenStream struct {
	rps int
	S chan bool
	closeCh chan bool
	wg *sync.WaitGroup
} 

// Stop stops the token stream
// It doesn't close channel S to prevent a read from the channel succeeding incorrectly.
func (t *TokenStream) Stop() {
	close(t.closeCh)
}

// TODO: generation is at a constant rate and doesn't accomodate spikes.
func (t *TokenStream) generate() {
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(t.rps))
		defer ticker.Stop()
		for {
			select {
			case <-t.closeCh:
				return
			case <-ticker.C:
				t.S <- true
			}
		}
	}()
}

// NewTokenStream creates a new token stream
func NewTokenStream(rps int) * TokenStream{
	t := TokenStream{rps:rps}
	t.closeCh = make(chan bool)
	t.S = make(chan bool, 10*rps)
	t.generate()
	return &t
}