package model

type McServerStatusResponse struct {
	Version       string
	PlayersNum    int
	PlayersNumMax int
	MOTD          string
	Favicon       string
}
