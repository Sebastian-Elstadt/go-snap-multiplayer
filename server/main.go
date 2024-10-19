package main

import (
	"fmt"
	"gosnapserver/comms"
	"gosnapserver/game"
	"log"
	"net"
	"sync"
)

var (
	playerConnectionMutex = &sync.Mutex{}
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Fatal("failed to setup tcp listener:", err)
	}

	defer listener.Close()
	fmt.Println("server started on port 8000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error on client connection:", err)
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	playerConnectionMutex.Lock()

	player := game.NewPlayer(conn)
	fmt.Printf("received connnection: %s ID: %d\n", conn.RemoteAddr().String(), player.ID)
	lobby := game.GetJoinableLobby()

	if !lobby.AddPlayer(player) {
		fmt.Printf("failed to add player %d to a lobby. disconnecting...\n", player.ID)
		player.WriteStatus(comms.STATUS_LOBBY_ERR)
		conn.Close()
		game.CleanUpLobbies()
		playerConnectionMutex.Unlock()
		return
	}

	playerConnectionMutex.Unlock()

	fmt.Printf("added player %d to lobby.\n", player.ID)
	player.WriteStatus(comms.STATUS_LOBBY_JOINED)
	go player.HandleComms()

	if len(lobby.Players) == lobby.MaxSize {
		fmt.Printf("starting lobby %d\n", lobby.ID)
		lobby.GameStarted = true
		lobby.StartGame()
	}
}
