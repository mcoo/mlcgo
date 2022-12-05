package mlcgo

import (
	"context"
	"mlcgo/downloader"
	"mlcgo/model"
	"mlcgo/resolver"
	"mlcgo/utils"
	"path/filepath"
	"runtime"

	"github.com/imroc/req/v3"
)

func (c *Core) autoCompletion(ctx context.Context, extraJobs []model.DownloadFile) error {
	if c.stepCh != nil {
		c.stepCh <- model.CompleteFilesStep
	}
	if c.maxDownloadCount == 0 {
		c.maxDownloadCount = runtime.NumCPU()
	}
	assets, err := resolver.ResolverAssets(c.clientInfo, c.minecraftPath, c.mirrorSource)
	if err != nil {
		return err
	}

	d := downloader.NewDownloader(10, c.maxDownloadCount)
	go func() {
		defer d.CloseJobCh()
		for _, v := range c.clientInfo.Libraries {
			if v.Native {
				libPath := filepath.Join(c.minecraftPath, "libraries", v.NativePath)
				job := CheckDownloadJob(libPath, v.NativeSha1, v.NativeURL)
				if job != nil {
					err = d.AddJob(ctx, c.replaceDownloadJob(job))
					if err != nil {
						return
					}
					if v.Path == v.NativePath {
						continue
					}
				}
			}
			libPath := filepath.Join(c.minecraftPath, "libraries", v.Path)
			job := CheckDownloadJob(libPath, v.Sha1, v.Url)
			if job != nil {
				err = d.AddJob(ctx, c.replaceDownloadJob(job))
				if err != nil {
					return
				}
			}
		}
		for _, v := range assets {
			job := CheckDownloadJob(v.Path, v.Hash, v.Url)
			if job != nil {
				err = d.AddJob(ctx, c.replaceDownloadJob(job))
				if err != nil {
					return
				}
			}
		}
		for _, v := range extraJobs {
			err = d.AddJob(ctx, c.replaceDownloadJob(&v))
			if err != nil {
				return
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
			log.Debugf("Download Error [%s]", job.Url)
		}

	})

	// 解压 native
	for _, v := range c.clientInfo.Libraries {
		if v.Native {
			utils.UnzipNative(filepath.Join(c.minecraftPath, "libraries", v.NativePath), c.nativeDir)
		}
	}
	return nil
}
