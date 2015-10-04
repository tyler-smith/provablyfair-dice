package dice

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
)

// ErrClientSeedBlank is the error returned if the supplied clientSeed is nil
// or a slice of 0 length
var ErrClientSeedBlank = errors.New("Client seed can't be empty")

// ErrInvalidNonce is returned when doesn't create a valid random number
var ErrInvalidNonce = errors.New("Invalid nonce")

// Game reprsents the current state of a dice game for a single client
type Game struct {
	ClientSeed        []byte
	ServerSeed        []byte
	BlindedServerSeed []byte

	Nonce uint64

	RollLock sync.Mutex
}

// NewGame creates a new game from the given seeds. A clientSeed is required
// If the serverSeed is nil we create a random one
func NewGame(clientSeed []byte, serverSeed []byte) (*Game, error) {
	// Validate the clientSeed
	if clientSeed == nil || len(clientSeed) == 0 {
		return nil, ErrClientSeedBlank
	}

	// Generate a random server seed if one isn't provided
	if serverSeed == nil || len(serverSeed) == 0 {
		var err error
		serverSeed, err = newServerSeed(32)
		if err != nil {
			return nil, err
		}
	}

	// Hash the serverSeed to show the client
	blindedSeed := sha256.Sum256(serverSeed)

	return &Game{
		Nonce:             0,
		ClientSeed:        clientSeed,
		ServerSeed:        serverSeed,
		BlindedServerSeed: blindedSeed[:],
	}, nil
}

// Roll calculates the number for the current nonce, then increments the nonce
// Doing it in this order ensures that the first once we use is 0
func (g *Game) Roll() (float64, error) {
	// Lock the RollLock so we can be safe across threads accessing the same game
	g.RollLock.Lock()
	defer g.RollLock.Unlock()

	// Calculate the current number from the current state
	roll, err := g.Calculate()
	if err != nil {
		return roll, err
	}

	// Increment the nonce for next time
	g.Nonce ++

	return roll, nil
}

// Calculate calculates the current value from the current state of the game
// It does not advance the state in anyway; i.e. simply calling Calculate
// multiple times will always result in the same value unless the Nonce changes
func (g *Game) Calculate() (float64, error) {
	// Calculate the HMAC for the current nonce
	ourHMAC := string(g.CalculateHMAC())

	// Find the first 5 character segment that converts to decimal < the max
	var randNum uint64
	var err error
	for i := 0; i < len(ourHMAC)-5; i++ {
		// Get the index for this segment and ensure it doesn't overrun the slice
		idx := i * 5
		if len(ourHMAC) < (idx + 5) {
			break
		}

		// Get 5 characters and convert them to decimal
		randNum, err = strconv.ParseUint(ourHMAC[idx:idx+5], 16, 0)
		if err != nil {
			return 0, err
		}

		// Continue unless our number was greater than our max
		if randNum <= 999999 {
			break
		}
	}

	// If even the last segment was invalid we must give up
	if randNum > 999999 {
		return 0, ErrInvalidNonce
	}

	// Normalize the number to [0,100]
	return float64(randNum%10000) / 100, nil
}

// CalculateHMAC calculates the hmac of "clientseed-nonce" as a hex string
func (g *Game) CalculateHMAC() []byte {
	h := hmac.New(sha512.New, g.ServerSeed)
	h.Write(append(append(g.ClientSeed, '-'), []byte(strconv.FormatUint(g.Nonce, 10))...))

	ourHMAC := make([]byte, 128)
	hex.Encode(ourHMAC, h.Sum(nil))
	return ourHMAC
}

// Verfiy takes a state and checks that the supplied number was fairly generated
func Verfiy(clientSeed []byte, serverSeed []byte, nonce uint64, randNum float64) (bool, error) {
	game, _ := NewGame(clientSeed, serverSeed)
	game.Nonce = nonce

	roll, err := game.Calculate()
	if err != nil {
		return false, err
	}

	return roll == randNum, nil
}

func newServerSeed(byteCount int) ([]byte, error) {
	seed := make([]byte, byteCount)
	_, err := rand.Read(seed)
	return seed, err
}
