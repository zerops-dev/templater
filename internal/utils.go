package internal

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

func getDirectoryPkgByModules(dir string) (dataPkg string, projectPkg string, packagePath string, projectPath string, err error) {
	dir = strings.TrimSuffix(dir, string(filepath.Separator))
	modFile := path.Join(dir, "go.mod")
	if _, exists := os.Stat(modFile); exists != nil {
		subDir := path.Base(dir)
		newDir := strings.TrimSuffix(strings.TrimSuffix(dir, subDir), string(filepath.Separator))
		if newDir == "" {
			err = fmt.Errorf("not found")
			return
		}
		dataPkg, projectPkg, packagePath, projectPath, err = getDirectoryPkgByModules(newDir)
		if err != nil {
			return
		}
		dataPkg = path.Join(dataPkg, subDir)
		packagePath = path.Join(packagePath, subDir)
		return
	}

	mod, err := os.Open(modFile)
	if err != nil {
		err = errors.Wrapf(err, "open file: %s err: %s", modFile)
		return
	}
	defer mod.Close()

	scanner := bufio.NewScanner(mod)
	scanner.Split(bufio.ScanLines)

	modRer := regexp.MustCompile("^module (.*)$")

	for scanner.Scan() {
		if modMatch := modRer.FindStringSubmatch(scanner.Text()); modMatch != nil {
			return modMatch[1], modMatch[1], dir, dir, nil
		}
	}
	err = fmt.Errorf("go.mod file not found")
	return
}

func FirstLower(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}
