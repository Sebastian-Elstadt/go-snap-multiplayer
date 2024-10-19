package game

import (
	"fmt"
	"gosnapserver/comms"
	"sync"
)

var (
	lobbyCleanupMutex = &sync.Mutex{}
)

var lobbyId int = 0
var lobbies []*Lobby

type Lobby struct {
	Players     []*Player
	ID          int
	MaxSize     int
	GameStarted bool

	Deck        *Deck
	PlayedCards []*Card

	playerDisconnectMutex *sync.Mutex
	snapCheckMutex        *sync.Mutex
}

func (l *Lobby) AddPlayer(p *Player) bool {
	if len(l.Players) >= l.MaxSize {
		return false
	}

	if l.GameStarted {
		return false
	}

	l.Players = append(l.Players, p)
	p.Lobby = l

	return true
}

func (l *Lobby) StartGame() {
	l.Deck = NewDeck()
	ShuffleDeck(l.Deck)

	l.Players[0].Cards, l.Players[1].Cards = SplitDeck(l.Deck)
	fmt.Printf("lobby %d deck shuffled and distributed. waiting for player action...\n", l.ID)

	l.Players[0].WriteMessage(comms.MSG_PLAYER_TURN)
}

func (l *Lobby) EndGame() {
	fmt.Printf("ending lobby %d\n", l.ID)
	for i := range l.Players {
		l.Players[i].WriteStatus(comms.STATUS_LOBBY_END)
		l.Players[i].Disconnect()
	}

	l.Players = nil
	CleanUpLobbies()
}

func (l *Lobby) OnPlayerDisconnect(player *Player) {
	l.playerDisconnectMutex.Lock()

	fmt.Printf("player %d disconnected, ending lobby %d...\n", player.ID, l.ID)

	for i := range l.Players {
		if l.Players[i].ID == player.ID {
			continue
		}

		l.Players[i].WriteStatus(comms.STATUS_OPP_DISCON)
		l.Players[i].Disconnect()
	}

	l.Players = nil

	CleanUpLobbies()

	l.playerDisconnectMutex.Unlock()
}

func (l *Lobby) PlayPlayerCard(player *Player) {
	topCard := player.Cards[len(player.Cards)-1]
	player.Cards = player.Cards[:len(player.Cards)-1]
	l.PlayedCards = append(l.PlayedCards, topCard)

	fmt.Printf("player %d has played card %s.\n", player.ID, topCard.ToBytes())

	for i := range l.Players {
		l.Players[i].WriteMessage(comms.MSG_PLAYER_PLAY_CARD)
		l.Players[i].WriteData(topCard.ToBytes())

		if l.Players[i].ID == player.ID {
			continue
		}

		if len(player.Cards) < 1 {
			fmt.Printf("player %d has run out of cards. game is over.\n", player.ID)
			player.WriteMessage(comms.MSG_PLAYER_LOST)
			l.Players[i].WriteMessage(comms.MSG_PLAYER_WON)
			l.EndGame()
		} else {
			l.Players[i].WriteMessage(comms.MSG_PLAYER_TURN)
		}
	}
}

func (l *Lobby) CheckSnap(player *Player) {
	l.snapCheckMutex.Lock()

	if len(l.PlayedCards) < 2 {
		l.snapCheckMutex.Unlock()
		return
	}

	lastCard := l.PlayedCards[len(l.PlayedCards)-1]
	secondLastCard := l.PlayedCards[len(l.PlayedCards)-2]

	if lastCard.Value != secondLastCard.Value {
		l.snapCheckMutex.Unlock()
		return
	}

	fmt.Printf("player %d has correctly called snap.\n", player.ID)

	player.Cards = append(player.Cards, l.PlayedCards...)
	l.PlayedCards = nil

	for i := range l.Players {
		if l.Players[i].ID == player.ID {
			player.WriteMessage(comms.MSG_PLAYER_SNAP_YOU)
		} else {
			l.Players[i].WriteMessage(comms.MSG_PLAYER_SNAP_OTHER)
		}
	}

	l.snapCheckMutex.Unlock()
}

func (l *Lobby) SendPlayerCardCount(player *Player) {
	player.WriteMessage(comms.MSG_PLAYER_ASK_CARD_COUNT)
	player.WriteData([]byte{byte(len(player.Cards))})
}

func GetJoinableLobby() *Lobby {
	for i := range lobbies {
		if !lobbies[i].GameStarted && len(lobbies[i].Players) < lobbies[i].MaxSize {
			return lobbies[i]
		}
	}

	lobbyId++
	lobby := &Lobby{
		ID:                    lobbyId,
		MaxSize:               2,
		playerDisconnectMutex: &sync.Mutex{},
		snapCheckMutex:        &sync.Mutex{},
	}

	lobbies = append(lobbies, lobby)

	return lobby
}

func CleanUpLobbies() {
	lobbyCleanupMutex.Lock()

	for i := len(lobbies) - 1; i >= 0; i-- {
		if len(lobbies[i].Players) == 0 {
			lobbies = append(lobbies[:i], lobbies[i+1:]...)
		}
	}

	lobbyCleanupMutex.Unlock()
}
