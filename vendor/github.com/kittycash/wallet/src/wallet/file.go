package wallet

import (
	"fmt"
	"gopkg.in/sirupsen/logrus.v1"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// This holds the root directory.
var (
	rootDir string
	log     = logrus.New()
)

// SetRootDir sets the root directory.
func SetRootDir(r string) error {
	var e error
	if rootDir, e = filepath.Abs(r); e != nil {
		return e
	}
	if e = os.MkdirAll(rootDir, os.FileMode(0700)); e != nil {
		return e
	}
	return nil
}

func ExtractLabel(filePath string) string {
	base := path.Base(filePath)
	return strings.TrimSuffix(base, string(FileExt))
}

func LabelPath(label string) string {
	return filepath.Join(rootDir, fmt.Sprintf("%s%s", label, FileExt))
}

func ListLabels() ([]string, error) {
	list, e := ioutil.ReadDir(rootDir)
	if e != nil {
		return nil, e
	}
	var out []string
	for _, info := range list {
		if info.IsDir() {
			continue
		}
		name := info.Name()
		if strings.HasSuffix(name, string(FileExt)) == false {
			continue
		}
		out = append(out, strings.TrimSuffix(name, string(FileExt)))
	}
	return out, nil
}

type LabelAction func(f io.Reader, label, fPath string, prefix Prefix)

func RangeLabels(action LabelAction) error {
	list, e := ioutil.ReadDir(rootDir)
	if e != nil {
		return e
	}
	for _, info := range list {
		if info.IsDir() {
			continue
		}
		name := info.Name()
		if strings.HasSuffix(name, string(FileExt)) == false {
			continue
		}
		label := strings.TrimSuffix(name, string(FileExt))
		fPath := LabelPath(label)

		f, e := os.Open(fPath)
		if e != nil {
			return e
		}

		var prefix Prefix
		f.Read(prefix[:])
		action(f, label, fPath, prefix)
		f.Close()
	}
	return nil
}
