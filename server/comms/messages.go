package comms

const (
	MSG_HEADER uint8 = 2

	MSG_PLAYER_TURN           uint8 = 2
	MSG_PLAYER_PLAY_CARD      uint8 = 3
	MSG_PLAYER_CALL_SNAP      uint8 = 4
	MSG_PLAYER_WON            uint8 = 5
	MSG_PLAYER_LOST           uint8 = 6
	MSG_PLAYER_SNAP_YOU       uint8 = 7
	MSG_PLAYER_SNAP_OTHER     uint8 = 8
	MSG_PLAYER_ASK_CARD_COUNT uint8 = 9
)
