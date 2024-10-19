package game

import "math/rand"

var suits = []rune{'D', 'H', 'S', 'C'}
var values = []rune{'A', '2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K'}

type Card struct {
	Value rune
	Suit  rune
}

func (c *Card) ToBytes() []byte {
	return []byte{byte(c.Value), byte(c.Suit)}
}

type Deck struct {
	Cards []*Card
}

func NewDeck() *Deck {
	deck := &Deck{}

	for _, suit := range suits {
		for _, value := range values {
			deck.Cards = append(deck.Cards, &Card{Value: value, Suit: suit})
		}
	}

	return deck
}

func ShuffleDeck(deck *Deck) {
	for i := range deck.Cards {
		j := rand.Intn(i + 1)
		deck.Cards[i], deck.Cards[j] = deck.Cards[j], deck.Cards[i]
	}
}

func SplitDeck(deck *Deck) (firstHalf, secondHalf []*Card) {
	half := len(deck.Cards) / 2
	return deck.Cards[:half], deck.Cards[half:]
}
