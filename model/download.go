package model

type DownloadFile struct {
	Url  string `json:"url"`
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
}
