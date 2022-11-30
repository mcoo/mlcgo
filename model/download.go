package model

type DownloadStatus int

const (
	Downloading DownloadStatus = iota
	Downloaded
	DownloadError
)

type DownloadFile struct {
	Url  string `json:"url"`
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
}
