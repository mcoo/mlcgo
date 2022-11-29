package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"mlcgo/model"
	"mlcgo/utils"
	"net/url"
	"os"
	"path/filepath"

	"github.com/imroc/req/v3"
)

type AuthlibInjectorAuth struct {
	AccessToken string `json:"access_token"`
	ClientToken string `json:"client_token"`
	Name        string `json:"name"`
}

type AuthlibInjectorJarDownload struct {
	BuildNumber int    `json:"build_number"`
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	Checksums   struct {
		Sha256 string `json:"sha256"`
	} `json:"checksums"`
}

func (al *AuthlibInjectorAuth) GetAuthlibInjectorJar(path string) error {
	resp, err := req.R().Get("https://authlib-injector.yushi.moe/artifact/latest.json")
	if err != nil {
		return err
	}
	var r AuthlibInjectorJarDownload
	err = resp.UnmarshalJson(&r)
	if err != nil {
		return err
	}
	if s, err := utils.Sha256(path); err == nil && strings.ToLower(s) == strings.ToLower(r.Checksums.Sha256) {
		return nil
	}
	_, err = req.R().SetOutputFile(path).Get(r.DownloadURL)
	return err
}

func (al *AuthlibInjectorAuth) Auth(loginInfo map[string]string) (*model.UserInfo, error) {
	rootUrl, ok := loginInfo["root_url"]
	if !ok {
		return nil, errors.New("root_url is not set")
	}
	var configPath, email, password, authJarPath string
	var configPassword []byte
	authJarPath, ok = loginInfo["authJarPath"]
	if !ok {
		return nil, errors.New("authJarPath is not set")
	}
	if v, ok := loginInfo["configPath"]; ok {
		configPath = v
	} else {
		return nil, errors.New("need configPath")
	}
	if v, ok := loginInfo["configPassword"]; ok {
		configPassword = []byte(v)
	} else {
		configPassword = PwdKey
	}
	al.readConfig(configPath, configPassword)

	if al.AccessToken != "" && al.ClientToken != "" {
		ok, err := verifyToken(rootUrl, al.AccessToken, al.ClientToken)
		if err == nil && ok {
			return &model.UserInfo{Name: al.Name, AccessToken: al.AccessToken, UUID: al.ClientToken}, nil
		}
		// refresh
		authResponse, err := refresh(rootUrl, al.AccessToken, al.ClientToken)
		if err == nil {
			return &model.UserInfo{Name: authResponse.SelectedProfile.Name, AccessToken: authResponse.AccessToken, UUID: authResponse.ClientToken}, nil
		}
	}
	email, ok = loginInfo["email"]
	if !ok {
		return nil, errors.New("email is not set")
	}
	password, ok = loginInfo["password"]
	if !ok {
		return nil, errors.New("password is not set")
	}

	a, err := login(rootUrl, email, password)
	if err != nil {
		return nil, err
	}
	al.AccessToken = a.AccessToken
	al.ClientToken = a.ClientToken
	al.Name = a.SelectedProfile.Name
	al.saveConfig(configPath, configPassword)
	err = al.GetAuthlibInjectorJar(authJarPath)
	if err != nil {
		return nil, err
	}
	return &model.UserInfo{Name: a.SelectedProfile.Name, AccessToken: a.AccessToken, UUID: a.ClientToken}, nil

}
func (al *AuthlibInjectorAuth) Logout() error {
	return nil
}

func (al *AuthlibInjectorAuth) saveConfig(path string, password []byte) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}
	config := *al
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configBytesEncrypt, err := utils.AesEncrypt(configBytes, password)
	if err != nil {
		return err
	}

	return os.WriteFile(path, configBytesEncrypt, 0777)
}
func (al *AuthlibInjectorAuth) readConfig(path string, password []byte) error {
	if utils.PathExists(path) {
		configBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		configBytesDecrypt, err := utils.AesDecrypt(configBytes, password)
		if err != nil {
			return err
		}
		var config AuthlibInjectorAuth
		err = json.Unmarshal(configBytesDecrypt, &config)
		if err != nil {
			return err
		}
		al.AccessToken = config.AccessToken
		al.Name = config.Name
		al.ClientToken = config.ClientToken
		return nil
	}
	return errors.New("Config file not found")
}
func login(rootUrl, email, password string) (*AuthResponse, error) {
	u, err := url.JoinPath(rootUrl, "/authserver/authenticate")
	if err != nil {
		return nil, err
	}
	resp, err := req.R().SetBodyJsonString(fmt.Sprintf(`{
		"username":"%s",
		"password":"%s",
		"clientToken":"",
		"requestUser":true,
		"agent":{
			"name":"Minecraft",
			"version":1
		}
	}`, email, password)).Post(u)

	if err != nil {
		return nil, err
	}
	var r AuthResponse
	err = resp.UnmarshalJson(&r)
	if err != nil {
		return nil, err
	}
	return &r, err
}

func refresh(rootUrl, accessToken, clientToken string) (*AuthResponse, error) {
	u, err := url.JoinPath(rootUrl, "/authserver/refresh")
	if err != nil {
		return nil, err
	}
	resp, err := req.R().SetBodyJsonString(fmt.Sprintf(`{
		"accessToken":"%s",
		"clientToken":"%s",
		"requestUser":true
	}`, accessToken, clientToken)).Post(u)

	if err != nil {
		return nil, err
	}
	var r AuthResponse
	err = resp.UnmarshalJson(&r)
	if err != nil {
		return nil, err
	}
	return &r, err
}

func verifyToken(rootUrl, accessToken, clientToken string) (bool, error) {
	u, err := url.JoinPath(rootUrl, "/authserver/validate")
	if err != nil {
		return false, err
	}
	resp, err := req.R().SetBodyJsonString(fmt.Sprintf(`{
		"accessToken":"%s",
		"clientToken":"%s"
	}`, accessToken, clientToken)).Post(u)

	if err != nil {
		return false, err
	}
	if resp.StatusCode == 204 {
		return true, nil
	} else {
		return false, nil
	}
}

type AuthResponse struct {
	AccessToken       string `json:"accessToken"`
	ClientToken       string `json:"clientToken"`
	AvailableProfiles []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Properties []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Signature string `json:"signature"`
		} `json:"properties"`
	} `json:"availableProfiles"`
	SelectedProfile struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Properties []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Signature string `json:"signature"`
		} `json:"properties"`
	} `json:"selectedProfile"`
}
