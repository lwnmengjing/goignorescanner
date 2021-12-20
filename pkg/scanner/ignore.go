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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
)

const (
	// invertPrefix is used to prefix the Patterns to symbolize inversions
	invertPrefix = "!"
)

var (
	_         DirectoryScanner = &defaultIgnorer{}
	directory string
	includes  []string
	// patterns holds all the possible Ignorable patterns
	ignorePatterns []FileIgnorePattern
)

// dockerIgnorer processes .dockerignore and use the extracted Patterns for ignoring files
type dockerIgnorer struct {
}

// NewOrDefault builds and returns the new or default DirectoryScanner interface
// In all file not found cases the this wil return the default ignorer
func NewOrDefault(dir, ignore string) (DirectoryScanner, error) {

	// every new gets clean includes
	includes = make([]string, 0)
	ignorePatterns = make([]FileIgnorePattern, 0)

	directory = dir

	ignoreFile := filepath.Join(dir, ignore)

	p, err := scanAndBuildPatternsList(ignoreFile)
	if err != nil {
		return nil, err
	}

	rawPatterns = append(rawPatterns, p...)

	switch {
	case os.IsNotExist(err):
		return &defaultIgnorer{}, nil
	case err != nil:
		return nil, err
	}

	for _, ip := range p {
		x, err := toFileIgnorePattern(ip)
		if err != nil {
			return nil, err
		}
		ignorePatterns = append(ignorePatterns, *x)
	}

	return &dockerIgnorer{}, nil
}

// scanAndBuildPatternsList takes the file typically the .dockerignore file
// splits the file by new line (\n) and normalize them with following rules
// - removes the UTF8 Byte Order Mark (BOM) characters
// - scanner comments line starting with "#"
// - trim spaces in the pattern
// - make the pattern as clean filenames using golang filepath utils
func scanAndBuildPatternsList(ignoreFile string) ([]string, error) {

	var patterns []string
	patterns = append(patterns, defaultPatterns...)

	fr, err := os.Open(ignoreFile)

	switch {
	case os.IsNotExist(err):
		return patterns, nil
	case err != nil:
		return nil, err
	}

	scanner := bufio.NewScanner(fr)

	// UTF8 byte order mark (BOM) which are typically first three bytes of the file with
	// the hexadecimal characters: EF,BB,BF
	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	lineNo := 0

	for scanner.Scan() {
		scannedBytes := scanner.Bytes()

		if lineNo == 0 {
			scannedBytes = bytes.TrimPrefix(scannedBytes, utf8bom)
		}

		pattern := string(scannedBytes)
		// removes the all unwanted spaces from the beginning and end of the string
		// e.g. "\n\t\nfoo bar\n\t" will be trimmed to "foo bar"
		pattern = strings.TrimSpace(pattern)
		lineNo++

		// When the line starts with comments like #, scanner the line.
		// e.g. dockerignore
		//   # comment to scanner target
		//   target
		if strings.HasPrefix("#", pattern) {
			continue
		}

		// Skip Empty lines that might be present in the file
		if pattern == "" {
			continue
		}

		// The Patterns can start with ! symbolizing inversion of the pattern
		// When ! is seen remove the ! to clean up the pattern for Path separators
		// e.g. Patterns line !README.md
		invert := strings.HasPrefix(pattern, invertPrefix)
		if invert {
			pattern = strings.TrimPrefix(pattern, invertPrefix)
		}

		if len(pattern) > 0 {
			pattern = filepath.Clean(pattern)
			pattern = filepath.ToSlash(pattern)
			if len(pattern) > 1 && strings.HasPrefix(pattern, string(os.PathSeparator)) {
				pattern = strings.TrimPrefix(pattern, string(os.PathSeparator))
			}
		}
		// prefix the clean pattern with invertPrefix
		if invert {
			pattern = invertPrefix + pattern
		}

		patterns = append(patterns, pattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s : %w", ignoreFile, err)
	}
	return patterns, nil
}

// Ignore implements DirectoryScanner, for  dockerignore cases
func (id *dockerIgnorer) Scan() ([]string, error) {
	err := godirwalk.Walk(directory, dirOpts)
	return includes, err
}
