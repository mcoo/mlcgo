package model

type LauncherProfiles struct {
	Profiles struct {
		Mlcgo struct {
			GameDir       string `json:"gameDir"`
			LastVersionID string `json:"lastVersionId"`
			Name          string `json:"name"`
		} `json:"MLCGO"`
	} `json:"profiles"`
	SelectedProfileName string `json:"selectedProfileName"`
}
