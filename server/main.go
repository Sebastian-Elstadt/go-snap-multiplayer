package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"slices"
	"sync"
)

const (
	COMMAND_RECEIVE_CARDS = "RC"
)

var suits = []rune{'H', 'D', 'C', 'S'}
var cardValues = []rune{'2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'} // T being 10

type Card struct {
	suit  rune
	value rune
}

func createDeck() []*Card {
	var deck []*Card

	for _, suit := range suits {
		for _, val := range cardValues {
			deck = append(deck, &Card{suit: suit, value: val})
		}
	}

	return deck
}

func shuffleDeck(deck []*Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

func splitDeck(deck []*Card) ([]*Card, []*Card) {
	half := len(deck) / 2
	return deck[:half], deck[half:]
}

type Player struct {
	conn net.Conn
	id   string
}

type Lobby struct {
	players     []*Player
	maxSize     int
	gameStarted bool
}

func (l *Lobby) addPlayer(p *Player) bool {
	if len(l.players) >= l.maxSize || l.gameStarted {
		return false
	}

	l.players = append(l.players, p)

	fmt.Printf("lobby player size: %d\n", len(l.players))

	if len(l.players) >= l.maxSize {
		l.startGame()
	}

	return true
}

func (l *Lobby) startGame() {
	l.gameStarted = true
	fmt.Println("enough players for lobby. starting game...")

	deck := createDeck()
	shuffleDeck(deck)

	player1Deck, player2Deck := splitDeck(deck)
	sendDeckToPlayer(l.players[0], player1Deck)
	sendDeckToPlayer(l.players[1], player2Deck)
}

func sendDeckToPlayer(player *Player, deck []*Card) {
	player.conn.Write([]byte(COMMAND_RECEIVE_CARDS))
	player.conn.Write([]byte{byte(len(deck))})

	for i := range deck {
		player.conn.Write([]byte{byte(deck[i].value), byte(deck[i].suit)})
	}
}

func newLobby(maxSize int) *Lobby {
	lobby := &Lobby{
		maxSize: maxSize,
	}

	lobbies = append(lobbies, lobby)
	return lobby
}

func addPlayerToLobby(player *Player) *Lobby {
	for i := range lobbies {
		if lobbies[i].addPlayer(player) {
			fmt.Println("added player to lobby.")
			return lobbies[i]
		}
	}

	newLobby := newLobby(2)
	newLobby.addPlayer(player)
	fmt.Println("added player to new lobby.")
	return newLobby
}

func getLobbyIndex(lobby *Lobby) int {
	return slices.IndexFunc(lobbies, func(l *Lobby) bool {
		for i := range lobby.players {
			foundMatch := false

			for j := range l.players {
				if lobby.players[i].id == l.players[j].id {
					foundMatch = true
					break
				}
			}

			if !foundMatch {
				return false
			}
		}

		return true
	})
}

func killLobby(lobby *Lobby) {
	for i := range lobby.players {
		lobby.players[i].conn.Close()
	}

	lobbyIndex := getLobbyIndex(lobby)
	lobbies = append(lobbies[:lobbyIndex], lobbies[lobbyIndex+1:]...)
}

var lobbies []*Lobby
var mutex *sync.Mutex

func main() {
	mutex = &sync.Mutex{}

	listener, err := net.Listen("tcp", "localhost:5000")
	if err != nil {
		log.Fatalf("failed to start tcp server: %v", err)
	}

	defer listener.Close()
	log.Println("server started on port 5000")

	for {
		client, err := listener.Accept()
		if err != nil {
			fmt.Println("conn err: ", err)
			continue
		}

		go handleClient(client)
	}
}

func handleClient(client net.Conn) {
	player := &Player{
		conn: client,
		id:   client.RemoteAddr().String(),
	}

	fmt.Println("player connected: ", player.id)

	mutex.Lock()
	lobby := addPlayerToLobby(player)
	mutex.Unlock()

	incoming := make([]byte, 128)
	for {
		_, err := client.Read(incoming)
		if err != nil {
			killLobby(lobby)
		}
	}
}
