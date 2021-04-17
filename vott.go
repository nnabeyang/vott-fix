package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type vottEntry interface {
	accept(cxt *vottContext) error
	fix(cxt *vottContext) error
	fileName() string
	save(src, dst string) error
}

type fileConnectionOptions struct {
	FolderPath string `json:"folderPath"`
}

type encryptedOptions struct {
	Encrypted string `json:"encrypted"`
}

type fileConnection struct {
	Name            string           `json:"name"`
	ProviderType    string           `json:"providerType"`
	ProviderOptions encryptedOptions `json:"providerOptions"`
	ID              string           `json:"id"`
}

type asset struct {
	Format string `json:"format"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"size"`
	State int `json:"state"`
	Type  int `json:"type"`
}

type vottFile struct {
	Name             string         `json:"name"`
	SecurityToken    string         `json:"securityToken"`
	SourceConnection fileConnection `json:"sourceConnection"`
	TargetConnection fileConnection `json:"targetConnection"`
	VideoSettings    struct {
		FrameExtractionRate int `json:"frameExtractionRate"`
	} `json:"videoSettings"`
	Tags []struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"tags"`
	ID                     string `json:"id"`
	ActiveLearningSettings struct {
		AutoDetect    bool   `json:"autoDetect"`
		PredictTag    bool   `json:"predictTag"`
		ModelPathType string `json:"modelPathType"`
	} `json:"activeLearningSettings"`
	ExportFormat struct {
		ProviderType    string           `json:"providerType"`
		ProviderOptions encryptedOptions `json:"providerOptions"`
	} `json:"exportFormat"`
	Version            string           `json:"version"`
	LastVisitedAssetId string           `json:"lastVisitedAssetId"`
	Assets             map[string]asset `json:"assets"`
}

func (f *vottFile) fix(cxt *vottContext) error {
	f.Assets = cxt.Assets
	if err := changeConnect(cxt.SecurityKey, cxt.SrcTargetBase, &f.TargetConnection.ProviderOptions); err != nil {
		return err
	}
	if err := changeConnect(cxt.SecurityKey, cxt.SrcImgBase, &f.SourceConnection.ProviderOptions); err != nil {
		return err
	}
	cxt.register("vott", f)
	return nil
}
func (f *vottFile) fileName() string {
	return fmt.Sprintf("%s.vott", f.Name)
}

func (f *vottFile) accept(cxt *vottContext) error {
	securityKey, err := readSecurityKey(cxt.SecurityPath, f.SecurityToken)
	if err != nil {
		return err
	}
	cxt.SecurityKey = securityKey
	srcTargetBase, err := cxt.getFolderPath(&f.TargetConnection)
	if err != nil {
		return err
	}

	srcImgBase, err := cxt.getFolderPath(&f.SourceConnection)
	if err != nil {
		return err
	}
	err = fixPath(cxt.TargetPath, &srcImgBase, &srcTargetBase)
	cxt.SrcImgBase = srcImgBase
	cxt.SrcTargetBase = srcTargetBase

	return err
}

func (f *vottFile) save(src, dst string) error {
	w := new(bytes.Buffer)
	decoder := json.NewEncoder(w)
	decoder.SetIndent("", "  ")
	if err := decoder.Encode(f); err != nil {
		return err
	}
	b, err := io.ReadAll(w)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0)
}

type assetFile struct {
	Asset   asset `json:"asset"`
	Regions []struct {
		ID          string   `json:"id"`
		Type        string   `json:"type"`
		Tags        []string `json:"tags"`
		BoundingBox struct {
			Height float64 `json:"height"`
			Width  float64 `json:"width"`
			Left   float64 `json:"left"`
			Top    float64 `json:"top"`
		} `json:"boundingBox"`
		Points []struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"points"`
	} `json:"regions"`
	Version string `json:"version"`
}

func (f *assetFile) fix(cxt *vottContext) error {
	f.Asset.Path = fmt.Sprintf("file:%s", path.Join(cxt.SrcImgBase, f.Asset.Name))
	oldID := f.Asset.ID
	f.Asset.ID = fmt.Sprintf("%x", md5.Sum([]byte(f.Asset.Path)))
	cxt.registerAsset(f.Asset)
	cxt.register(oldID, f)
	return nil
}

func (f *assetFile) fileName() string {
	return fmt.Sprintf("%s-asset.json", f.Asset.ID)
}

func (f *assetFile) accept(cxt *vottContext) error {
	return nil
}

func (f *assetFile) save(src, dst string) error {
	w := new(bytes.Buffer)
	decoder := json.NewEncoder(w)
	decoder.SetIndent("", "  ")
	if err := decoder.Encode(f); err != nil {
		return err
	}
	b, err := io.ReadAll(w)
	if err != nil {
		return err
	}
	if err := os.WriteFile(src, b, 0); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func changeConnect(securityKey string, newPath string, options *encryptedOptions) error {
	enc, err := encryptFolderPath(securityKey, newPath)
	if err != nil {
		return err
	}
	options.Encrypted = enc
	return nil
}

func getFolderPath(securityKey, data string) (string, error) {
	srcPath, err := decrypt(securityKey, data)
	if err != nil {
		if _, ok := err.(base64.CorruptInputError); ok {
			return "", errors.New("セキュリティトークンの形式が正しくありません")
		}

		return "", err
	}
	buf := bytes.NewBuffer([]byte(srcPath))
	var srcParams fileConnectionOptions

	if err := json.NewDecoder(buf).Decode(&srcParams); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			return "", errors.New("セキュリティトークンが正しくありません")
		}
		return "", err
	}
	return srcParams.FolderPath, nil
}

func encryptFolderPath(securityKey, newPath string) (string, error) {
	iv, err := generateRandomKey(24)
	if err != nil {
		return "", err
	}
	input, err := json.Marshal(fileConnectionOptions{
		FolderPath: newPath,
	})
	if err != nil {
		return "", err
	}
	enc, err := encrypt(securityKey, string(input), hex.EncodeToString(iv))
	if err != nil {
		return "", err
	}
	return enc, nil
}

type vottContext struct {
	SecurityPath  string
	SecurityKey   string
	TargetPath    string
	SrcImgBase    string
	SrcTargetBase string
	Assets        map[string]asset
	Entries       map[string]vottEntry
}

func (c *vottContext) registerAsset(a asset) {
	c.Assets[a.ID] = a
}

func (c *vottContext) register(id string, e vottEntry) {
	c.Entries[id] = e
}

func (c *vottContext) getFolderPath(fc *fileConnection) (string, error) {
	return getFolderPath(c.SecurityKey, fc.ProviderOptions.Encrypted)
}

func (c *vottContext) fix(srcTargetBase string) error {
	c.Assets = make(map[string]asset)
	c.Entries = make(map[string]vottEntry)
	if err := filepath.Walk(srcTargetBase, func(path string, f os.FileInfo, err error) error {
		return visitFile(c, path, f, err)
	}); err != nil {
		return err
	}
	return filepath.Walk(srcTargetBase, func(path string, f os.FileInfo, err error) error {
		return fixFile(c, path, f, err)
	})
}

func processFile(c *vottContext, assetPath string) error {
	entry, err := loadEntry(assetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: skipped\n", assetPath)
		return nil
	}
	return entry.accept(c)
}

func processFixFile(c *vottContext, assetPath string) error {
	entry, err := loadEntry(assetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: skipped\n", assetPath)
		return nil
	}
	return entry.fix(c)
}

func fixFile(c *vottContext, path string, f os.FileInfo, err error) error {
	if err == nil && isEntryFile(f) {
		err = processFixFile(c, path)
	}
	return err
}

func visitFile(c *vottContext, path string, f os.FileInfo, err error) error {
	if err == nil && isEntryFile(f) {
		err = processFile(c, path)
	}
	return err
}

func isEntryFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && (strings.HasSuffix(name, "-asset.json") || strings.HasSuffix(name, ".vott"))
}

func (c *vottContext) save() error {
	for id, entry := range c.Entries {
		dst := filepath.Join(c.SrcTargetBase, entry.fileName())
		src := filepath.Join(c.SrcTargetBase, fmt.Sprintf("%s-asset.json", id))
		if err := entry.save(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func createVottContext(securityPath, targetPath string) (*vottContext, error) {
	targetBase, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, err
	}

	return &vottContext{
		SecurityPath: securityPath,
		TargetPath:   targetBase,
	}, nil
}
