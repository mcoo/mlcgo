package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	mlclog "mlcgo/log"
	"mlcgo/model"
	"mlcgo/utils"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

const (
	mlcgo_client_id    = "f10650e0-a065-4638-8721-74a7f137cf3b"
	mlcgo_redirect_uri = "http://localhost:8809/oauth20_desktop.srf"
)

var log = mlclog.Log

/*
MicrosoftAuth
1. Get MicrosoftAuth Code
2. Get MicrosoftAuth Token
3. Get XBL Token
4. Get XSTS Token
5. Auth Minecraft
6. Verify Game Ownership
7. Get Profile
*/
type MicrosoftAuth struct {
	microsoft_info *MicrosoftOauthResponse
	minecraft_info *MinecraftAuthResponse
}

type SaveFile struct {
	Microsoft MicrosoftOauthResponse `json:"microsoft"`
	Minecraft MinecraftAuthResponse  `json:"minecraft"`
}

func (ms *MicrosoftAuth) Auth(loginInfo map[string]string) (*model.UserInfo, error) {
	var configPath string
	var Password []byte
	if v, ok := loginInfo["configPath"]; ok {
		configPath = v
	} else {
		return nil, errors.New("need configPath")
	}
	if v, ok := loginInfo["configPassword"]; ok {
		Password = []byte(v)
	} else {
		Password = PwdKey
	}
	err := ms.readConfig(configPath, Password)
	if err == nil {
		if ms.minecraft_info.ExpiresIn+ms.minecraft_info.TimeNow.Unix() > time.Now().Unix() {
			// 还没过期
			info, err := ms.loginWithMinecraftAccessToken()
			if err == nil {
				return info, nil
			}
		}
		if ms.microsoft_info != nil {
			info, err := ms.loginWithMicrosoft(configPath, Password)
			if err == nil {
				return info, nil
			}
		}
	}
	return ms.firstLogin(configPath, Password)
}

func (ms *MicrosoftAuth) firstLogin(configPath string, Password []byte) (*model.UserInfo, error) {
	loginUrl := getMicrosoftAuthUrl("", "")
	log.Infoln("loginUrl:", loginUrl)
	utils.OpenUrl(loginUrl)
	code, err := getMicrosoftCode("")
	if err != nil {
		return nil, err
	}
	m_info, err := getMicrosoftToken(code, "", "", "")
	if err != nil {
		return nil, err
	}
	m_info.TimeNow = time.Now()
	xbl_info, err := getXBLToken(m_info.AccessToken)
	if err != nil {
		return nil, err
	}
	xsts_info, err := getXSTSToken(xbl_info.Token)
	if err != nil {
		return nil, err
	}
	if len(xsts_info.DisplayClaims.Xui) == 0 {
		return nil, errors.New("you don not have user in xbox")
	}
	mc_info, err := authenticateMinecraft(xsts_info.DisplayClaims.Xui[0].Uhs, xsts_info.Token)
	if err != nil {
		return nil, err
	}
	mc_info.TimeNow = time.Now()
	isOwnGame, err := verifyMinecraftOwnership(mc_info.AccessToken)
	if err != nil {
		return nil, err
	}
	if !isOwnGame {
		return nil, errors.New("your account don not own minecraft")
	}
	profile, err := getProfile(mc_info.AccessToken)
	if err != nil {
		return nil, err
	}
	ms.microsoft_info = m_info
	ms.minecraft_info = mc_info
	ms.saveConfig(configPath, Password)
	return &model.UserInfo{
		Name:        profile.Name,
		UUID:        profile.ID,
		Skins:       profile.Skins,
		AccessToken: mc_info.AccessToken,
	}, nil
}

func (ms *MicrosoftAuth) loginWithMicrosoft(configPath string, Password []byte) (*model.UserInfo, error) {
	m_info := ms.microsoft_info
	if m_info.ExpiresIn+m_info.TimeNow.Unix() < time.Now().Unix() {
		// access_token 无效了 使用refresh_token尝试
		m_info_new, err := getMicrosoftToken("", m_info.RefreshToken, "", "")
		if err != nil {
			return nil, err
		}
		m_info = m_info_new
	}

	xbl_info, err := getXBLToken(m_info.AccessToken)
	if err != nil {
		return nil, err
	}
	xsts_info, err := getXSTSToken(xbl_info.Token)
	if err != nil {
		return nil, err
	}
	if len(xsts_info.DisplayClaims.Xui) == 0 {
		return nil, errors.New("you don not have user in xbox")
	}
	mc_info, err := authenticateMinecraft(xsts_info.DisplayClaims.Xui[0].Uhs, xsts_info.Token)
	if err != nil {
		return nil, err
	}
	mc_info.TimeNow = time.Now()
	isOwnGame, err := verifyMinecraftOwnership(mc_info.AccessToken)
	if err != nil {
		return nil, err
	}
	if !isOwnGame {
		return nil, errors.New("your account don not own minecraft")
	}
	profile, err := getProfile(mc_info.AccessToken)
	if err != nil {
		return nil, err
	}
	ms.microsoft_info = m_info
	ms.minecraft_info = mc_info
	ms.saveConfig(configPath, Password)
	return &model.UserInfo{
		Name:        profile.Name,
		UUID:        profile.ID,
		Skins:       profile.Skins,
		AccessToken: mc_info.AccessToken,
	}, nil
}

func (ms *MicrosoftAuth) loginWithMinecraftAccessToken() (*model.UserInfo, error) {
	isOwnGame, err := verifyMinecraftOwnership(ms.minecraft_info.AccessToken)
	if err != nil {
		return nil, err
	}
	if !isOwnGame {
		return nil, errors.New("your account don not own minecraft")
	}
	profile, err := getProfile(ms.minecraft_info.AccessToken)
	if err != nil {
		return nil, err
	}
	return &model.UserInfo{
		Name:        profile.Name,
		UUID:        profile.ID,
		Skins:       profile.Skins,
		AccessToken: ms.minecraft_info.AccessToken,
	}, nil
}

func (ms *MicrosoftAuth) Logout() error {
	return nil
}
func (ms *MicrosoftAuth) saveConfig(path string, password []byte) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}
	config := SaveFile{
		Minecraft: *ms.minecraft_info,
		Microsoft: *ms.microsoft_info,
	}
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
func (ms *MicrosoftAuth) readConfig(path string, password []byte) error {
	if utils.PathExists(path) {
		configBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		configBytesDecrypt, err := utils.AesDecrypt(configBytes, password)
		if err != nil {
			return err
		}
		var config SaveFile
		err = json.Unmarshal(configBytesDecrypt, &config)
		if err != nil {
			return err
		}
		ms.microsoft_info = &config.Microsoft
		ms.minecraft_info = &config.Minecraft
		return nil
	}
	return errors.New("config file not found")
}

func getMicrosoftAuthUrl(client_id string, redirect_uri string) string {
	if client_id == "" {
		client_id = mlcgo_client_id
	}
	if redirect_uri == "" {
		redirect_uri = mlcgo_redirect_uri
	}
	q := url.Values{}
	q.Add("client_id", client_id)
	q.Add("response_type", "code")
	q.Add("scope", "XboxLive.signin offline_access")
	q.Add("redirect_uri", redirect_uri)
	return "https://login.live.com/oauth20_authorize.srf?" + q.Encode()
}

func getMicrosoftCode(host string) (code string, err error) {
	log.Debugln("get Microsoft code")
	if host == "" {
		host = "localhost:8809"
	}
	codeCh := make(chan string, 1)
	http.HandleFunc("/oauth20_desktop.srf", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		c := q.Get("code")
		if c != "" {
			codeCh <- c
			io.WriteString(w, "Successful!")
			return
		} else {
			io.WriteString(w, "Failed!")
			return
		}
	})
	srv := http.Server{
		Addr:    host,
		Handler: http.DefaultServeMux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Errorln(err)
			}
		}
	}()
	if err != nil {
		log.Errorln(err)
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		err = errors.New("context timeout")
	case code = <-codeCh:
	}
	err = srv.Shutdown(context.Background())
	return
}
func getMicrosoftToken(code, refresh_token, client_id, redirect_uri string) (m *MicrosoftOauthResponse, err error) {
	log.Debugln("get Microsoft token")
	if client_id == "" {
		client_id = mlcgo_client_id
	}
	if redirect_uri == "" {
		redirect_uri = mlcgo_redirect_uri
	}
	if code == "" && refresh_token != "" {
		// 第二次登录
		var oauthToken MicrosoftOauthResponse
		resp, err := req.R().SetFormData(map[string]string{
			"client_id":     client_id,
			"refresh_token": refresh_token,
			"grant_type":    "refresh_token",
			"redirect_uri":  redirect_uri,
		}).Post("https://login.live.com/oauth20_token.srf")
		if err != nil {
			return nil, err
		}
		err = resp.UnmarshalJson(&oauthToken)
		if oauthToken.AccessToken == "" {
			log.Debugln(resp.String())
			return nil, fmt.Errorf("get AccessToken error")
		}

		return &oauthToken, err
	} else if code != "" {
		// 第一次登录
		var oauthToken MicrosoftOauthResponse
		resp, err := req.R().SetFormData(map[string]string{
			"client_id":    client_id,
			"code":         code,
			"grant_type":   "authorization_code",
			"redirect_uri": redirect_uri,
		}).Post("https://login.live.com/oauth20_token.srf")
		if err != nil {
			return nil, err
		}
		err = resp.UnmarshalJson(&oauthToken)
		if oauthToken.AccessToken == "" {
			log.Debugln(resp.String())
			return nil, fmt.Errorf("get AccessToken error")
		}
		return &oauthToken, err

	} else {
		// 你想干啥？
		return nil, errors.New("参数错误")
	}
}

func getXBLToken(access_token string) (m *XBLResponse, err error) {
	log.Debugln("get xbl token")
	// 第一次登录
	var oauthToken XBLResponse
	resp, err := req.R().SetBodyJsonString(fmt.Sprintf(`
	{"Properties": {"AuthMethod": "RPS","SiteName": "user.auth.xboxlive.com","RpsTicket": "d=%s"},"RelyingParty": "http://auth.xboxlive.com","TokenType": "JWT"}`, access_token)).
		Post("https://user.auth.xboxlive.com/user/authenticate")
	if err != nil {
		return nil, err
	}
	err = resp.UnmarshalJson(&oauthToken)
	if err != nil {
		log.Debugln(fmt.Sprintf(`
		{"Properties": {"AuthMethod": "RPS","SiteName": "user.auth.xboxlive.com","RpsTicket": "d=%s"},"RelyingParty": "http://auth.xboxlive.com","TokenType": "JWT"}`, access_token))
		log.Debugln(resp.String())
	}
	return &oauthToken, err
}

func getXSTSToken(xbl_token string) (*XSTSResponse, error) {
	log.Debugln("get xsts token")
	var token XSTSResponse
	resp, err := req.R().SetBodyJsonString(fmt.Sprintf(`
	{
		"Properties": {
			"SandboxId": "RETAIL",
			"UserTokens": [
				"%s"
			]
		},
		"RelyingParty": "rp://api.minecraftservices.com/",
		"TokenType": "JWT"
	}
	`, xbl_token)).
		Post("https://xsts.auth.xboxlive.com/xsts/authorize")
	if err != nil {
		return nil, err
	}
	err = resp.UnmarshalJson(&token)
	return &token, err
}

// Authenticate with Minecraft
func authenticateMinecraft(userhash, xsts_token string) (*MinecraftAuthResponse, error) {
	log.Debugln("authenticate Minecraft")
	var token MinecraftAuthResponse
	resp, err := req.R().SetBodyJsonString(fmt.Sprintf(`
	{
		"identityToken": "XBL3.0 x=%s;%s"
	}
	`, userhash, xsts_token)).
		Post("https://api.minecraftservices.com/authentication/login_with_xbox")
	if err != nil {
		return nil, err
	}
	err = resp.UnmarshalJson(&token)
	return &token, err
}

func verifyMinecraftOwnership(access_token string) (bool, error) {
	log.Debugln("verify Minecraft Ownership")
	resp, err := req.R().SetHeader("Authorization", "Bearer "+access_token).Get("https://api.minecraftservices.com/entitlements/mcstore")
	if err != nil {
		return false, err
	}
	g := gjson.GetBytes(resp.Bytes(), "items")
	if g.Exists() && len(g.Array()) > 0 {
		return true, nil
	}
	return false, nil
}

func getProfile(access_token string) (*ProfileResponse, error) {
	log.Debugln("get Profile")
	profile := ProfileResponse{}
	resp, err := req.R().SetHeader("Authorization", "Bearer "+access_token).Get("https://api.minecraftservices.com/minecraft/profile")
	if err != nil {
		return nil, err
	}
	err = resp.UnmarshalJson(&profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

type ProfileResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Skins []struct {
		ID      string `json:"id"`
		State   string `json:"state"`
		URL     string `json:"url"`
		Variant string `json:"variant"`
		Alias   string `json:"alias"`
	} `json:"skins"`
	Capes []interface{} `json:"capes"`
}

type MicrosoftOauthResponse struct {
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	Scope        string    `json:"scope"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UserID       string    `json:"user_id"`
	TimeNow      time.Time `json:"time_now"`
}
type XBLResponse struct {
	IssueInstant  time.Time `json:"IssueInstant"`
	NotAfter      time.Time `json:"NotAfter"`
	Token         string    `json:"Token"`
	DisplayClaims struct {
		Xui []struct {
			Uhs string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}
type XSTSResponse struct {
	IssueInstant  time.Time `json:"IssueInstant"`
	NotAfter      time.Time `json:"NotAfter"`
	Token         string    `json:"Token"`
	DisplayClaims struct {
		Xui []struct {
			Uhs string `json:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}
type MinecraftAuthResponse struct {
	Username string        `json:"username"`
	Roles    []interface{} `json:"roles"`
	Metadata struct {
	} `json:"metadata"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int64     `json:"expires_in"`
	TokenType   string    `json:"token_type"`
	TimeNow     time.Time `json:"time_now"`
}
