package comms

const (
	STATUS_HEADER uint8 = 1

	STATUS_LOBBY_ERR    uint8 = 1
	STATUS_LOBBY_JOINED uint8 = 2
	STATUS_OPP_DISCON   uint8 = 3
	STATUS_LOBBY_END    uint8 = 4
)
