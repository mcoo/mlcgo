package downloader

import (
	"context"
	"errors"
	"fmt"
	mlclog "mlcgo/log"
	"mlcgo/model"
	"mlcgo/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
)

var log = mlclog.Log

type Downloader struct {
	downloadCh chan model.DownloadFile
	maxThreads int
	Callback   func(req.DownloadInfo)
}

func NewDownloader(downloadChBuffer, maxThreads int) *Downloader {
	return &Downloader{
		downloadCh: make(chan model.DownloadFile, downloadChBuffer),
		maxThreads: maxThreads,
	}
}

func (d *Downloader) AddJob(ctx context.Context, job *model.DownloadFile) error {
	select {
	case <-ctx.Done():
		log.Debugln("接收到结束指令")
		return errors.New("context canceled")
	case d.downloadCh <- *job:
		return nil
	}
}

func (d *Downloader) CloseJobCh() {
	close(d.downloadCh)
}

func (d *Downloader) StartDownload(ctx context.Context, callback func(req.DownloadInfo, model.DownloadFile, model.DownloadStatus)) {
	var wg sync.WaitGroup
	for i := 0; i < d.maxThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			DownloadWork(ctx, d.downloadCh, callback)
		}()
	}
	wg.Wait()
}
func DownloadWork(ctx context.Context, workCh chan model.DownloadFile, callback func(req.DownloadInfo, model.DownloadFile, model.DownloadStatus)) error {
	for job := range workCh {
		dir := filepath.Dir(job.Path)
		os.MkdirAll(dir, 0755)
		_, err := req.
			R().SetOutputFile(job.Path).SetDownloadCallbackWithInterval(func(info req.DownloadInfo) {
			if info.Response.Response != nil {
				callback(info, job, model.Downloading)
			}
		}, time.Second).Get(job.Url)
		if err != nil {
			callback(req.DownloadInfo{}, job, model.DownloadError)
			log.Errorln(job.Path, job.Url, err)
		}
		if job.Sha1 != "" {
			sha1, err := utils.SHA1File(job.Path)
			if err != nil {
				callback(req.DownloadInfo{}, job, model.DownloadError)
				log.Errorln(job.Path, job.Url, err)
				continue
			}
			if !strings.EqualFold(sha1, job.Sha1) {
				callback(req.DownloadInfo{}, job, model.DownloadError)
				log.Errorln(job.Path, job.Url, fmt.Errorf("check [%s] sha1 fails:%s not equal %s", job.Path, sha1, job.Sha1))
				continue
			}
		}

		callback(req.DownloadInfo{}, job, model.Downloaded)
	}
	return nil
}
func ReplaceDownloadUrl(url string, mirror model.DownloadMirrorSource) string {
	switch mirror {
	case model.Mojang:
	case model.BMCL:
		url = strings.ReplaceAll(url, "launchermeta.mojang.com", "bmclapi2.bangbang93.com")
		url = strings.ReplaceAll(url, "launcher.mojang.com", "bmclapi2.bangbang93.com")
		url = strings.ReplaceAll(url, "resources.download.minecraft.net", "bmclapi2.bangbang93.com/assets")
		url = strings.ReplaceAll(url, "libraries.minecraft.net", "bmclapi2.bangbang93.com/maven")
		url = strings.ReplaceAll(url, "files.minecraftforge.net/maven", "bmclapi2.bangbang93.com/maven")
		url = strings.ReplaceAll(url, "dl.liteloader.com/versions/versions.json", "bmclapi.bangbang93.com/maven/com/mumfrey/liteloader/versions.json")
		url = strings.ReplaceAll(url, "authlib-injector.yushi.moe", "bmclapi2.bangbang93.com/mirrors/authlib-injector")
		url = strings.ReplaceAll(url, "meta.fabricmc.net", "bmclapi2.bangbang93.com/fabric-meta")
		url = strings.ReplaceAll(url, "maven.fabricmc.net", "bmclapi2.bangbang93.com/maven")

	case model.Mcbbs:
		url = strings.ReplaceAll(url, "launchermeta.mojang.com", "download.mcbbs.net")
		url = strings.ReplaceAll(url, "launcher.mojang.com", "download.mcbbs.net")
		url = strings.ReplaceAll(url, "resources.download.minecraft.net", "download.mcbbs.net/assets")
		url = strings.ReplaceAll(url, "libraries.minecraft.net", "download.mcbbs.net/maven")
		url = strings.ReplaceAll(url, "files.minecraftforge.net/maven", "download.mcbbs.net/maven")
		url = strings.ReplaceAll(url, "dl.liteloader.com/versions/versions.json", "bmclapi.bangbang93.com/maven/com/mumfrey/liteloader/versions.json")
		url = strings.ReplaceAll(url, "authlib-injector.yushi.moe", "download.mcbbs.net/mirrors/authlib-injector")
		url = strings.ReplaceAll(url, "meta.fabricmc.net", "download.mcbbs.net/fabric-meta")
		url = strings.ReplaceAll(url, "maven.fabricmc.net", "download.mcbbs.net/maven")
	}
	return url

}
