/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scanner

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/karrick/godirwalk"
	"k8s.io/apimachinery/pkg/util/sets"
)

// IgnorePattern holds the ignorable patterns
type FileIgnorePattern struct {
	Pattern      string
	Paths        []string
	RegexPattern *regexp.Regexp
	Invert       bool
}

// DirectoryScanner  helps identifying if a BundleFile needs to be ignored
type DirectoryScanner interface {
	// Scan checks file has to be ignored or not, returns true if it needs to be ignored
	Scan() ([]string, error)
}

var (
	_               DirectoryScanner = (*defaultIgnorer)(nil)
	defaultPatterns                  = []string{".git", "vendor", "node_modules"}
	rawPatterns     []string
	excludedDirs    sets.String = sets.NewString()
)

// scanDir scans the directory and checks whether the file or directory is ignorable
func scanDir(osPathname string, dirEntry *godirwalk.Dirent) error {

	// convenience to flag the current path as not root directory
	rootDir := osPathname == directory

	// dont append the rootdir as its always included
	if rootDir {
		return nil
	}

	// if the dir is any of default patterns then skip scanning the dir
	skipPaths := sets.NewString(defaultPatterns...)
	if skipPaths.Has(dirEntry.Name()) {
		return godirwalk.SkipThis
	}

	// flag to keep track if file or directory is excluded
	isExcluded := false

	// flag to keep track transitive  directories
	// Transitive directories are are the ones that are excluded
	// but might have files under them with inversions
	// e.g. with a ignore file like
	// foo
	// !foo/bar/one.txt
	// In above case directory foo is transitive
	isTransitive := false

	for _, igp := range ignorePatterns {

		re := igp.RegexPattern

		// if pattern is not a inversion and the excluded directories has
		// the current path parent the continue to check next pattern
		if !igp.Invert && excludedDirs.Has(filepath.Dir(osPathname)) {
			isExcluded = true
			continue
		}

		regxMatches := re.FindAllStringSubmatch(osPathname, -1)

		if len(regxMatches) > 0 {
			isExcluded = true

			// when a file or directory matches the pattern but has inversion
			// then add the file to include list
			if igp.Invert {
				for _, tuple := range regxMatches {
					rel, err := filepath.Rel(directory, tuple[0])
					if err != nil {
						return err
					}
					appendIfNotExist(rel)
				}
			}
		}

		if dirEntry.IsDir() && !rootDir && igp.Invert && igp.Paths != nil {
			pPath := strings.Join(igp.Paths, string(os.PathSeparator))
			pPath = filepath.Join(directory, pPath)
			if osPathname == pPath {
				isTransitive = true
			}
		}

	}

	if isExcluded {
		// if the matched directory then add the directory to excludedDirs list
		if dirEntry.IsDir() {
			excludedDirs.Insert(osPathname)
		}
	} else {
		// If there are no matches, then the walked directory or file need
		// to be included as part of the include file list

		// build relative path to root directory
		rel, err := filepath.Rel(directory, osPathname)

		if err != nil {
			return err
		}

		appendIfNotExist(rel)
	}

	// the directory is not transitive and not root directory check is being done
	// at last to avoid skipping directories that are not listed in the ignore file
	// with "!" i.e. implicit includes
	if dirEntry.IsDir() && !rootDir && isExcluded && !isTransitive {
		log.Printf("Directory %s will be skipped from further iteration", osPathname)
		return godirwalk.SkipThis
	}

	return nil
}

var dirOpts = &godirwalk.Options{
	Callback: scanDir,
}

// defaultIgnorer is the default DirectoryScanner which is returned when no .dockerignore file is present
// or error processing .dockerignore
type defaultIgnorer struct{}

// Ignore implements DirectoryScanner, for no dockerignore cases, where only .git is ignored
func (i *defaultIgnorer) Scan() ([]string, error) {
	err := godirwalk.Walk(directory, dirOpts)
	return includes, err
}

func appendIfNotExist(item string) {
	incl := sets.NewString(includes...)
	if !incl.Has(item) {
		includes = append(includes, item)
	}
}
