package installer

import (
	"context"
	"fmt"
	"io/ioutil"
	"mlcgo/downloader"
	"mlcgo/model"
	"mlcgo/utils"
	"os"
	"path/filepath"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

/*
https://meta.fabricmc.net/v2/versions/loader
https://meta.fabricmc.net/v2/versions/game
https://meta.fabricmc.net/v2/versions/loader/1.19.2/0.14.11/profile/json
*/

func InstallFabric(ctx context.Context, gameDir, mc_version, fabric_version string, maxThreads int) error {

	log.Debugln("get loader json")
	resp, err := req.SetContext(ctx).Get(fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json", mc_version, fabric_version))
	if err != nil {
		return err
	}
	loader_json := gjson.ParseBytes(resp.Bytes())
	id := loader_json.Get("id").String()

	log.Debugf("output version json: %s", id)
	path := filepath.Join(gameDir, "versions", id, id+".json")
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, resp.Bytes(), 0777)
	if err != nil {
		return err
	}

	log.Debugf("download libraries")

	d := downloader.NewDownloader(10, maxThreads)

	go func() {
		defer d.CloseJobCh()
		for _, v := range loader_json.Get("libraries").Array() {

			path, err := utils.ArtifactFrom(v.Get("name").String())
			if err != nil {
				log.Error(err)
				continue
			}

			err = d.AddJob(ctx, &model.DownloadFile{
				Url:  "https://maven.fabricmc.net/" + path,
				Path: filepath.Join(gameDir, "libraries", path),
				Sha1: "",
			})
			if err != nil {
				log.Debug(err)
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

	}()
	d.StartDownload(ctx, func(info req.DownloadInfo, job model.DownloadFile, status model.DownloadStatus) {
		switch status {
		case model.Downloading:
			if info.Response.Response != nil {
				log.Debugf("Downloading [%s] %.2f%%", job.Path, float64(info.DownloadedSize)/float64(info.Response.ContentLength)*100.0)
			}
		case model.Downloaded:
			log.Debugf("Downloaded [%s]", job.Url)
		case model.DownloadError:
			log.Errorf("Download Error [%s]", job.Url)
		}
	})
	return nil
}
