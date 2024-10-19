package game

import (
	"fmt"
	"gosnapserver/comms"
	"io"
	"net"
)

var playerId int = 0

type Player struct {
	conn net.Conn
	ID   int

	Lobby *Lobby
	Cards []*Card

	intentionallyDisconnected bool
}

func (p *Player) HandleComms() {
	for {
		header := make([]byte, 1)
		_, err := io.ReadFull(p.conn, header)
		if err != nil {
			if !p.intentionallyDisconnected {
				p.OnCommsError(err)
			}

			return
		}

		switch header[0] {
		case comms.MSG_PLAYER_PLAY_CARD:
			fmt.Printf("player %d wants to play a card.\n", p.ID)
			p.Lobby.PlayPlayerCard(p)
		case comms.MSG_PLAYER_CALL_SNAP:
			fmt.Printf("player %d is calling snap.\n", p.ID)
			p.Lobby.CheckSnap(p)
		case comms.MSG_PLAYER_ASK_CARD_COUNT:
			fmt.Printf("player %d is asking their card count.\n", p.ID)
			p.Lobby.SendPlayerCardCount(p)
		}
	}
}

func (p *Player) Disconnect() {
	p.intentionallyDisconnected = true
	p.conn.Close()
}

func (p *Player) WriteData(data []byte) {
	_, err := p.conn.Write(data)
	if err != nil {
		p.OnCommsError(err)
	}
}

func (p *Player) WriteStatus(status uint8) {
	_, err := p.conn.Write([]byte{comms.STATUS_HEADER, status})
	if err != nil {
		p.OnCommsError(err)
	}
}

func (p *Player) WriteMessage(message uint8) {
	_, err := p.conn.Write([]byte{comms.MSG_HEADER, message})
	if err != nil {
		p.OnCommsError(err)
	}
}

func (p *Player) OnCommsError(err error) {
	fmt.Printf("ERR: %v\n", err)
	fmt.Printf("player %d has encountered a comms error. disconnecting...\n", p.ID)
	p.conn.Close()
	p.Lobby.OnPlayerDisconnect(p)
}

func NewPlayer(conn net.Conn) *Player {
	playerId++

	return &Player{
		conn: conn,
		ID:   playerId,
	}
}
