package mcserver

import (
	mlclog "mlcgo/log"
	"mlcgo/model"

	"github.com/PassTheMayo/mcstatus/v4"
)

var log = mlclog.Log

func PingServer(server string, port uint16) (serverInfo *model.McServerStatusResponse, e error) {
	serverInfo = &model.McServerStatusResponse{}
	r, err := mcstatus.Status("jmcraft.co", port)
	if err != nil {
		return nil, err
	}
	serverInfo.Version = r.Version.Clean
	serverInfo.MOTD = r.MOTD.HTML
	serverInfo.PlayersNum = r.Players.Online
	serverInfo.PlayersNumMax = r.Players.Max
	if r.Favicon != nil {
		serverInfo.Favicon = *r.Favicon
	}
	return
}
