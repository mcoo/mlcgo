package model

type UserInfo struct {
	UUID        string `json:"uuid"`
	AccessToken string `json:"access_token"`
	Skins       []struct {
		ID      string `json:"id"`
		State   string `json:"state"`
		URL     string `json:"url"`
		Variant string `json:"variant"`
		Alias   string `json:"alias"`
	} `json:"skins"`
	Name string `json:"name"`
}
