package main

import (
	"bufio"
	"cardgameclient/comms"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var conn net.Conn
var isPlayerTurn = false

func awaitServerBytes(size uint) []byte {
	buf := make([]byte, size)

	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Fatalf("an error occurred while talking to the server: %v", err)
		conn.Close()
	}

	return buf
}

func main() {
	fmt.Println("welcome to snapgame.")
	fmt.Println()
	fmt.Println("your controls are:")
	fmt.Println("p - play a card, if it is your turn.")
	fmt.Println("c - see how many cards you still have.")
	fmt.Println("s - call snap.")
	fmt.Println()
	fmt.Println("connecting to server...")

	var err error
	if conn, err = net.Dial("tcp", "localhost:8000"); err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}

	fmt.Println("you are connected to the server.")
	fmt.Println()
	go listenToServer()
	handlePlayerInput()
}

func handlePlayerInput() {
	reader := bufio.NewReader(os.Stdin)

	for {
		char, err := reader.ReadByte()
		if err != nil {
			fmt.Printf("something went wrong with your keypress: %v", err)
		}

		switch char {
		case 'p':
			playCard()
		case 's':
			callSnap()
		case 'c':
			askCardCount()
		}
	}
}

func playCard() {
	if !isPlayerTurn {
		fmt.Println("it is not your turn to play.")
		return
	}

	_, err := conn.Write([]byte{comms.MSG_HEADER, comms.MSG_PLAYER_PLAY_CARD})
	if err != nil {
		fmt.Println("failed to talk to server. disconnecting...")
		stopGame()
	}

	isPlayerTurn = false
}

func callSnap() {
	_, err := conn.Write([]byte{comms.MSG_HEADER, comms.MSG_PLAYER_CALL_SNAP})
	if err != nil {
		fmt.Println("failed to talk to server. disconnecting...")
		stopGame()
	}
}

func askCardCount() {
	_, err := conn.Write([]byte{comms.MSG_HEADER, comms.MSG_PLAYER_ASK_CARD_COUNT})
	if err != nil {
		fmt.Println("failed to talk to server. disconnecting...")
		stopGame()
	}
}

func stopGame() {
	conn.Close()
	os.Exit(0)
}

func listenToServer() {
	for {
		header := awaitServerBytes(1)

		switch header[0] {
		case comms.STATUS_HEADER:
			handleIncomingStatus()
		case comms.MSG_HEADER:
			handleIncomingMessage()
		}
	}
}

func handleIncomingStatus() {
	statusCode := awaitServerBytes(1)

	switch statusCode[0] {
	case comms.STATUS_LOBBY_END:
		fmt.Println("the game has ended. thanks for playing.")
		stopGame()
	case comms.STATUS_LOBBY_ERR:
		fmt.Println("a problem occurred while trying to find an opponent. please try again later.")
		stopGame()
	case comms.STATUS_LOBBY_JOINED:
		fmt.Println("waiting for opponent to join....")
	case comms.STATUS_OPP_DISCON:
		fmt.Println("your opponent has disconnected. the game will now end.")
		stopGame()
	}
}

func handleIncomingMessage() {
	msgCode := awaitServerBytes(1)

	switch msgCode[0] {
	case comms.MSG_PLAYER_TURN:
		fmt.Println("it is your turn to play!")
		isPlayerTurn = true
	case comms.MSG_PLAYER_PLAY_CARD:
		cardBytes := awaitServerBytes(2)
		fmt.Printf("card played: %s\n", cardBytes)
	case comms.MSG_PLAYER_LOST:
		fmt.Println("you have lost!")
	case comms.MSG_PLAYER_WON:
		fmt.Println("you have won!")
	case comms.MSG_PLAYER_SNAP_YOU:
		fmt.Println("you have successfully called snap! you have received all cards on the table.")
	case comms.MSG_PLAYER_SNAP_OTHER:
		fmt.Println("your opponent has successfully called snap. they have received all cards on the table.")
	case comms.MSG_PLAYER_ASK_CARD_COUNT:
		cardCountBytes := awaitServerBytes(1)
		fmt.Printf("you have %d cards.\n", cardCountBytes[0])
	}
}
