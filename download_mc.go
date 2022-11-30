package mlcgo

import (
	"context"
	"mlcgo/model"
	"mlcgo/resolver"
	"os"
	"path/filepath"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

func (c *Core) DownloadGame(ctx context.Context, version model.Version) error {
	path := filepath.Join(c.minecraftPath, "versions", version.ID)
	os.MkdirAll(path, 0755)
	_, err := req.R().SetOutputFile(filepath.Join(path, version.ID+".json")).Get(version.URL)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Join(path, version.ID+".json"))
	if err != nil {
		return err
	}
	g := gjson.ParseBytes(b)
	client := g.Get("downloads.client")
	c.version = version.ID
	c.clientInfo, err = resolver.ResolverClient(b, c.isDemo, c.isCustomResolution, c.minecraftPath)
	if err != nil {
		return err
	}
	clientJar := []model.DownloadFile{
		{
			Url:  client.Get("url").String(),
			Path: filepath.Join(path, version.ID+".jar"),
			Sha1: client.Get("sha1").String(),
		},
	}
	log.Debugln(clientJar)
	c.autoCompletion(ctx, clientJar)

	return nil
}
