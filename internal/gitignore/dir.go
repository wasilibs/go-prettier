package gitignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const (
	commentPrefix   = "#"
	gitDir          = ".git"
	gitignoreFile   = ".gitignore"
	infoExcludeFile = gitDir + "/info/exclude"
)

// ReadIgnoreFile reads a specific git ignore file.
func ReadIgnoreFile(path string, ignoreFile string) (ps []Pattern, err error) {
	f, err := os.Open(filepath.Join(path, ignoreFile))
	if err == nil {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			s := scanner.Text()
			if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
				ps = append(ps, ParsePattern(s, strings.Split(path, string(filepath.Separator))))
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}

// ReadPatterns reads the .git/info/exclude and then the gitignore patterns
// recursively traversing through the directory structure. The result is in
// the ascending order of priority (last higher).
func ReadPatterns(path string) (ps []Pattern, err error) {
	ps, _ = ReadIgnoreFile(path, infoExcludeFile)

	subps, _ := ReadIgnoreFile(path, gitignoreFile)
	ps = append(ps, subps...)

	var fis []os.DirEntry
	fis, err = os.ReadDir(path)
	if err != nil {
		return
	}

	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != gitDir {
			var subps []Pattern
			subps, err = ReadPatterns(filepath.Join(path, fi.Name()))
			if err != nil {
				return
			}

			if len(subps) > 0 {
				ps = append(ps, subps...)
			}
		}
	}

	return
}
