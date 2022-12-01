//go:build windows

package utils

import (
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func FindJavaPath() (j []string, e error) {
	// 从注册表寻找 java
	var javas = make(map[string]struct{})
	// JDK
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\JavaSoft\JDK`, registry.READ)
	if err != nil {
		log.Debugln(err)
		return nil, err
	}
	jdks, err := key.ReadSubKeyNames(0)
	if err != nil {
		log.Debugln(err)
		return nil, err
	}
	for _, v := range jdks {
		jdk, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\JavaSoft\JDK\`+v, registry.READ)
		if err != nil {
			log.Debugln(err)
			continue
		}
		path, _, err := jdk.GetStringValue("JavaHome")
		if err != nil {
			log.Debugln(err)
			continue
		}
		javas[path] = struct{}{}
	}

	key, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\JavaSoft\Java Runtime Environment`, registry.READ)
	if err != nil {
		log.Debugln(err)
		return nil, err
	}
	jres, err := key.ReadSubKeyNames(0)
	if err != nil {
		log.Debugln(err)
		return nil, err
	}
	for _, v := range jres {
		jre, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\JavaSoft\Java Runtime Environment\`+v, registry.READ)
		if err != nil {
			log.Debugln(err)
			continue
		}
		path, _, err := jre.GetStringValue("JavaHome")
		if err != nil {
			log.Debugln(err)
			continue
		}
		javas[path] = struct{}{}
	}
	for k := range javas {
		if p := filepath.Join(k, "bin", "javaw.exe"); PathExists(p) {
			j = append(j, p)
		}
	}
	return j, nil
}
