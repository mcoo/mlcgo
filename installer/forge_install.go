package installer

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"mlcgo/downloader"
	mlclog "mlcgo/log"
	"mlcgo/model"
	"mlcgo/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

var log = mlclog.Log

/*
Forge 1.13+
1. 下载原版
2. 下载Forge的Installer
3. 解压Installer.jar中的install_profile.json，version.json
4. 下载其中的 libraries
5. 执行 processors
*/

type Forge struct {
	Branch    string    `json:"branch"`
	Build     int       `json:"build"`
	Mcversion string    `json:"mcversion"`
	Modified  time.Time `json:"modified"`
	Version   string    `json:"version"`
	ID        string    `json:"_id"`
	Files     []struct {
		Format   string `json:"format"`
		Category string `json:"category"`
		Hash     string `json:"hash"`
		ID       string `json:"_id"`
	} `json:"files"`
}

func GetForgeList(ctx context.Context, mc_version string) (forges []Forge, err error) {
	resp, err := req.SetContext(ctx).Get("https://bmclapi2.bangbang93.com/forge/minecraft/" + mc_version)
	if err != nil {
		return nil, err
	}
	err = resp.UnmarshalJson(&forges)
	if err != nil {
		return nil, err
	}
	return forges, nil
}

func GetForge(ctx context.Context, forge Forge, gameDir, javaPath string) (*ForgeInstaller, error) {
	log.Debugln("downloading installer Jar")
	resp, err := req.SetContext(ctx).Get("https://bmclapi2.bangbang93.com/forge/download/" + strconv.Itoa(forge.Build))
	if err != nil {
		return nil, err
	}
	log.Debugln("downloaded installer Jar")
	log.Debugln("unzip installer Jar")
	installer_reader, err := zip.NewReader(bytes.NewReader(resp.Bytes()), int64(len(resp.Bytes())))
	if err != nil {
		return nil, err
	}
	// install_profile.json
	log.Debugln("unzip install_profile.json")
	install_profile, err := installer_reader.Open("install_profile.json")
	if err != nil {
		return nil, err
	}
	defer install_profile.Close()
	install_profile_bytes, err := ioutil.ReadAll(install_profile)
	if err != nil {
		return nil, err
	}
	install := gjson.ParseBytes(install_profile_bytes)

	// version.json
	log.Debugln("unzip version.json")
	version, err := installer_reader.Open("version.json")
	if err != nil {
		return nil, err
	}
	version_bytes, err := ioutil.ReadAll(version)
	if err != nil {
		return nil, err
	}
	log.Debugln("unzip data ")
	// client.lzma & server.lzma
	os.MkdirAll("./forgeInstallFiles/data", 0777)

	for _, v := range installer_reader.File {
		if v.FileInfo().IsDir() {
			continue
		}
		m, _ := filepath.Rel("data", v.Name)
		if !strings.Contains(m, "..") {
			path := filepath.Join("./forgeInstallFiles", v.Name)
			dir := filepath.Dir(path)
			os.MkdirAll(dir, 0777)
			rc, err := v.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			w, err := os.Create(path)
			if err != nil {
				return nil, err
			}
			defer w.Close()
			_, err = io.Copy(w, rc)
			if err != nil {
				return nil, err
			}
		}
	}

	log.Debugln("written forgeInstaller.jar")
	err = ioutil.WriteFile("./forgeInstallFiles/forgeInstaller.jar", resp.Bytes(), 0777)
	if err != nil {
		return nil, err
	}

	// 获取 data
	log.Debugln("resolve data")
	data := make(map[string]string)

	install.Get("data").ForEach(func(key, value gjson.Result) bool {
		data["{"+key.String()+"}"], err = GetSpecialData(value.Get("client").String(), filepath.Join(gameDir, "libraries"))
		return true
	})

	data["{SIDE}"] = "client"
	data["{MINECRAFT_JAR}"], err = filepath.Abs(filepath.Join(gameDir, "versions", install.Get("minecraft").String(), install.Get("minecraft").String()+".jar"))

	if err != nil {
		return nil, err
	}
	data["{INSTALLER}"], err = filepath.Abs("./forgeInstallFiles/forgeInstaller.jar")

	if err != nil {
		return nil, err
	}

	data["{LIBRARY_DIR}"], err = filepath.Abs(filepath.Join(gameDir, "libraries"))
	if err != nil {
		return nil, err
	}
	data["{ROOT}"], err = filepath.Abs("./forgeInstallFiles")
	if err != nil {
		return nil, err
	}
	data["{MINECRAFT_VERSION}"] = forge.Mcversion
	return &ForgeInstaller{
		installer:     &install,
		version_bytes: version_bytes,
		data:          data,
		librariesPath: filepath.Join(gameDir, "libraries"),
		javaPath:      javaPath,
	}, nil
}

type ForgeInstaller struct {
	javaPath      string
	installer     *gjson.Result
	librariesPath string
	version_bytes []byte
	data          map[string]string
}

func (i *ForgeInstaller) ExecProcessors(ctx context.Context) error {
	splitString := ";"
	if runtime.GOOS == "windows" {
		splitString = ";"
	} else {
		splitString = ":"
	}
	log.Debugln("resolve processors")
	for _, v := range i.installer.Get("processors").Array() {
		do := false
		if len(v.Get("sides").Array()) == 0 {
			do = true
		}
		for _, s := range v.Get("sides").Array() {
			if s.String() == "client" {
				do = true
				continue
			}
		}
		if !do {
			continue
		}
		log.Debugln("exec processor")
		cmdArgs := []string{}
		// cp
		classpath := []string{}
		log.Debugln("generate classpath")
		for _, j := range v.Get("classpath").Array() {
			f, err := utils.ArtifactFrom(j.String())
			if err != nil {
				return err
			}
			p, err := filepath.Abs(filepath.Join(i.librariesPath, f))
			if err != nil {
				return err
			}
			classpath = append(classpath, p)
		}
		log.Debugln("get mainClass")
		mainClass, err := utils.ArtifactFrom(v.Get("jar").String())
		if err != nil {
			return err
		}
		mainClass, err = filepath.Abs(filepath.Join(i.librariesPath, mainClass))
		if err != nil {
			return err
		}
		classpath = append(classpath, mainClass)
		cmdArgs = append(cmdArgs, "-cp", strings.Join(classpath, splitString))
		z, err := zip.OpenReader(mainClass)
		if err != nil {
			return err
		}
		defer z.Close()
		f, err := z.Open("META-INF/MANIFEST.MF")
		if err != nil {
			return err
		}
		defer f.Close()
		r := bufio.NewReader(f)
		mainClassName := ""
		for {
			line, err := r.ReadString('\n')
			if err != nil || err == io.EOF {
				break
			}
			m := strings.Split(line, ":")
			if len(m) < 2 {
				continue
			}
			if m[0] == "Main-Class" {
				mainClassName = strings.TrimSpace(m[1])
			}
		}
		if mainClassName == "" {
			return errors.New("not found mainclass")
		}
		log.Debugln("mainClass:", mainClassName)
		cmdArgs = append(cmdArgs, mainClassName)
		re := regexp.MustCompile(`\{.+?\}`)
		log.Debugln("get args")
		for _, arg := range v.Get("args").Array() {
			arg, err := GetSpecialData(arg.String(), i.librariesPath)
			if err != nil {
				return err
			}
			arg = re.ReplaceAllStringFunc(arg, func(s string) string {
				if m, ok := i.data[s]; ok {
					return m
				}
				return s
			})
			cmdArgs = append(cmdArgs, arg)
		}
		logFile, err := os.OpenFile("./forgeInstallFiles/install.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			return err
		}
		defer logFile.Close()
		cmd := exec.Command(i.javaPath, cmdArgs...)
		cmd.Stderr = logFile
		cmd.Stdout = logFile
		log.Debugln("exec command")
		err = cmd.Run()
		if err != nil {
			return err
		}

	}
	return nil
}
func (i *ForgeInstaller) DownloadLibraries(ctx context.Context, maxThreads int) error {
	d := downloader.NewDownloader(10, maxThreads)
	go func() {
		defer d.CloseJobCh()
		for _, library := range i.installer.Get("libraries").Array() {
			path := filepath.Join(i.librariesPath, library.Get("downloads.artifact.path").String())
			sha1 := library.Get("downloads.artifact.sha1").String()
			s, err := utils.SHA1File(path)
			if utils.PathExists(path) && strings.EqualFold(sha1, s) && err == nil {
				continue
			}

			err = d.AddJob(ctx, &model.DownloadFile{
				Url:  library.Get("downloads.artifact.url").String(),
				Path: path,
				Sha1: sha1,
			})
			if err != nil {
				log.Debug(err)
			}
			select {
			case <-ctx.Done():
				return
			default:
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
			log.Errorf("Download Error [%s]", job.Url)
		}
	})

	return nil
}

func GetSpecialData(value string, librariesPath string) (string, error) {
	if strings.HasPrefix(value, "[") {
		f, err := utils.ArtifactFrom(value[1 : len(value)-1])
		if err != nil {
			return "", err
		}
		f2, err := filepath.Abs(filepath.Join(librariesPath, f))
		if err != nil {
			return "", err
		}
		return f2, nil
	}
	if strings.HasPrefix(value, "/") {
		return filepath.Abs(filepath.Join("./forgeInstallFiles", value))
	}
	if strings.HasPrefix(value, "'") {
		return value[1 : len(value)-1], nil
	}
	return value, nil
}

func (i *ForgeInstaller) OutputVersionJson(ctx context.Context, gameDir string) error {
	path := filepath.Join(gameDir, "versions", i.installer.Get("version").String(), i.installer.Get("version").String()+".json")
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, i.version_bytes, 0777)
}

type InstallProfile struct {
	Comment       []string    `json:"_comment_"`
	Spec          int         `json:"spec"`
	Profile       string      `json:"profile"`
	Version       string      `json:"version"`
	Path          interface{} `json:"path"`
	Minecraft     string      `json:"minecraft"`
	ServerJarPath string      `json:"serverJarPath"`
	Data          struct {
		Mappings struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MAPPINGS"`
		Mojmaps struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MOJMAPS"`
		MergedMappings struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MERGED_MAPPINGS"`
		Binpatch struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"BINPATCH"`
		McUnpacked struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MC_UNPACKED"`
		McSlim struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MC_SLIM"`
		McSlimSha struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MC_SLIM_SHA"`
		McExtra struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MC_EXTRA"`
		McExtraSha struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MC_EXTRA_SHA"`
		McSrg struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MC_SRG"`
		Patched struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"PATCHED"`
		PatchedSha struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"PATCHED_SHA"`
		McpVersion struct {
			Client string `json:"client"`
			Server string `json:"server"`
		} `json:"MCP_VERSION"`
	} `json:"data"`
	Processors []struct {
		Sides     []string `json:"sides,omitempty"`
		Jar       string   `json:"jar"`
		Classpath []string `json:"classpath"`
		Args      []string `json:"args"`
		Outputs   struct {
			MCSLIM  string `json:"{MC_SLIM}"`
			MCEXTRA string `json:"{MC_EXTRA}"`
		} `json:"outputs,omitempty"`
	} `json:"processors"`
	Libraries []struct {
		Name      string `json:"name"`
		Downloads struct {
			Artifact struct {
				Path string `json:"path"`
				URL  string `json:"url"`
				Sha1 string `json:"sha1"`
				Size int    `json:"size"`
			} `json:"artifact"`
		} `json:"downloads"`
	} `json:"libraries"`
	Icon       string `json:"icon"`
	JSON       string `json:"json"`
	Logo       string `json:"logo"`
	MirrorList string `json:"mirrorList"`
	Welcome    string `json:"welcome"`
}
