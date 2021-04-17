package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func loadEntry(filePath string) (vottEntry, error) {
	switch filepath.Ext(filePath) {
	case ".json":
		var af assetFile
		if err := load(&af, filePath); err != nil {
			return nil, err
		}
		return &af, nil
	case ".vott":
		var vott vottFile
		if err := load(&vott, filePath); err != nil {
			return nil, err
		}
		return &vott, nil
	default:
		return nil, errors.New("skip")
	}
}
func load(x interface{}, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	json.NewDecoder(f).Decode(x)
	return nil
}

func projPath(srcPath, targetPath string) (string, string) {
	i := 0
	for srcPath[i] == targetPath[i] {
		i++
	}
	return srcPath[i:], targetPath[i:]
}

func fixPath(srcPath string, srcImgBase, srcTargetBase *string) error {
	relImgBase, relTargetBase := projPath(*srcImgBase, *srcTargetBase)
	srcProjBase := srcPath[:len(srcPath)-len(relTargetBase)-1]

	*srcImgBase = filepath.Join(srcProjBase, relImgBase)
	*srcTargetBase = filepath.Join(srcProjBase, relTargetBase)
	return nil
}

func readSecurityKey(keyPath, name string) (string, error) {
	if keyPath != "" {
		r, err := os.Open(keyPath)
		if err != nil {
			return "", err
		}
		defer r.Close()
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	} else {
		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, os.Interrupt)
		defer signal.Stop(signalChan)
		currentState, err := terminal.GetState(int(syscall.Stdin))
		if err != nil {
			return "", err
		}

		go func() {
			<-signalChan
			terminal.Restore(int(syscall.Stdin), currentState)
			os.Exit(1)
		}()
		fmt.Printf("Enter Security Key(%s):", name)
		bytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", err
		}
		fmt.Println()
		return string(bytes), nil
	}
}
