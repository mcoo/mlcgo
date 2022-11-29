package model

type Library struct {
	Name       string `json:"name"`
	Sha1       string `json:"sha1"`
	Size       int64  `json:"size"`
	Url        string `json:"url"`
	Path       string `json:"path"`
	Native     bool   `json:"native"`
	NativeURL  string `json:"native-url"`
	NativeSha1 string `json:"native-sha1"`
	NativePath string `json:"native-path"`
	NativeSize int64  `json:"native-size"`
}
