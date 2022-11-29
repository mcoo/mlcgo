package model

type ClientInfo struct {
	JvmArguments  []string
	GameArguments []string
	Libraries     []Library
	AssetIndex    struct {
		ID        string `json:"id"`
		Sha1      string `json:"sha1"`
		Size      int    `json:"size"`
		TotalSize int    `json:"totalSize"`
		URL       string `json:"url"`
	}
	JavaVersion struct {
		Component    string `json:"component"`
		MajorVersion int    `json:"majorVersion"`
	}
	Logging struct {
		Argument string `json:"argument"`
		File     struct {
			ID   string `json:"id"`
			Sha1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"file"`
		Type string `json:"type"`
	}
	MainClass   string
	Id          string
	VersionType string
}
type Asset struct {
	Url  string `json:"url"`
	Path string `json:"path"`
	Hash string `json:"hash"`
	Size int    `json:"size:"`
}
