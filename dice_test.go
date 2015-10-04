package dice

import (
	"fmt"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testRollExpectations struct {
	*Game

	Nonce uint64
	HMAC []byte
	Roll float64
}

var (
	ExampleClientSeed = []byte("ClientSeedForDiceSites.com")
	ExampleServerSeed = []byte("293d5d2ddd365f54759283a8097ab2640cbe6f8864adc2b1b31e65c14c999f04")
)

func TestGame(t *testing.T) {
	for _, testRoll := range newTestRolls(t) {
		assert.Equal(t, testRoll.Nonce, testRoll.Game.Nonce)
		assert.Equal(t, testRoll.HMAC, testRoll.Game.CalculateHMAC())

		roll, err := testRoll.Game.Roll()
		assert.NoError(t, err)
		assert.Equal(t, testRoll.Roll, roll)

		if testing.Verbose() {
			fmt.Println("Client Seed:", string(testRoll.ClientSeed))
			fmt.Println("Server Seed:", string(testRoll.ServerSeed))
			fmt.Println("Blinded Server Seed Hex:", hex.EncodeToString(testRoll.BlindedServerSeed))
			fmt.Println("Nonce:", testRoll.Nonce)
			fmt.Println("HMAC:", string(testRoll.HMAC))
			fmt.Println("Roll:", roll)
		}
	}
}

func TestRandomServerSeed(t *testing.T) {
	game, err := NewGame(ExampleClientSeed, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), game.Nonce)
  assert.Equal(t, ExampleClientSeed, game.ClientSeed)
  assert.Equal(t, 32, len(game.ServerSeed))
  assert.Equal(t, 32, len(game.BlindedServerSeed))
}

func TestVerify(t *testing.T) {
	for _, testRoll := range newTestRolls(t) {
		verified, err := Verfiy(testRoll.ClientSeed, testRoll.ServerSeed, testRoll.Nonce, testRoll.Roll)
		assert.NoError(t, err)
		assert.True(t, verified)
	}
}

func newExampleGame(t *testing.T) *Game {
	game, err := NewGame(ExampleClientSeed, ExampleServerSeed)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), game.Nonce)
  assert.Equal(t, ExampleClientSeed, game.ClientSeed)
  assert.Equal(t, ExampleServerSeed, game.ServerSeed)
  assert.Equal(t, 32, len(game.BlindedServerSeed))
	return game
}

func newTestRolls(t *testing.T) []*testRollExpectations {
	game := newExampleGame(t)
	return []*testRollExpectations{
		{
			Game: game,
			Nonce: 0,
			HMAC: []byte("aa671aad5e4565ebffb8dc5c185e4df1ae6d9aca2578b5c03ec9c7750f881922276d8044e5e3d84f158ce411f667e224e9b0c1ac50fc94e9c5eb883a678f6ca2"),
			Roll: 79.69,
		},
		{
			Game: game,
			Nonce: 1,
			HMAC: []byte("7b9062b1a8188feff82d643c0c8f2883bc744240594952f55126b24c76b05648a73850905e68fe86fe64c9fbd9a9ef9f677264d3771bd98db64b022ad183da53"),
			Roll: 61.18,
		},
		{
			Game: game,
			Nonce: 2,
			HMAC: []byte("a5644976f61b4012c0eb27848bbe3d05d43d34dcb89e2032b8d93ba0992b26ad916223caf9ba5421229508144a370ba053f27893b5e7f6e8283231cce90e1535"),
			Roll: 74.44,
		},
		{
			Game: game,
			Nonce: 3,
			HMAC: []byte("8bd6805955d0ca66fb5eb672b75bd0874bea59ecbe1d21e101ad50faf19e7d67256d6d4714c53fa848d801d92874f72813a78e447431b1fd609ba328d18d3875"),
			Roll: 27.76,
		},
	}

}
