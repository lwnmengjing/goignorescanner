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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
)

// toFileIgnorePattern is used to perform normalization on the pattern like
// - clean up the path to be good file path
// - make sure paths start with /
// - check if the patterns has inversions i.e !foo kind of things
// - compile the pattern to regular expression
func toFileIgnorePattern(pattern string) (*FileIgnorePattern, error) {

	// clean the pattern to be well formed Go path
	pattern = filepath.Clean(pattern)

	// make sure the path starts with /
	pattern = filepath.FromSlash(pattern)

	ignorePattern := &FileIgnorePattern{}
	ignorePattern.Pattern = pattern

	// check if it has inverts and remove them before creating paths
	if strings.HasPrefix(pattern, "!") {
		pattern = strings.TrimPrefix(pattern, "!")
		ignorePattern.Invert = true
	}

	// add the parent directories if and only if the pattern has a /
	// otherwise the path is deemed to be under root
	if strings.Contains(pattern, string(os.PathSeparator)) {
		ptokens := strings.Split(pattern, string(os.PathSeparator))
		ignorePattern.Paths = ptokens[:len(ptokens)-1]
	}

	re, err := asRegExp(pattern)

	if err != nil {
		return nil, err
	}

	ignorePattern.RegexPattern = re

	return ignorePattern, nil
}

// make each pattern as valid regular expression that can be compared
// with file path
func asRegExp(pattern string) (*regexp.Regexp, error) {
	pathSep := string(os.PathSeparator)
	escPath := pathSep

	// make sure the unix paths are escaped with \\
	if pathSep == `\` {
		escPath += `\`
	}

	//start
	regexPat := ""

	var s scanner.Scanner
	s.Init(strings.NewReader(pattern))

	for s.Peek() != scanner.EOF {
		ch := s.Next()

		//handle *
		if ch == '*' {
			if s.Peek() == '*' {
				//check if next char is also *, typically like **
				s.Next()
				//Treat **/ as **
				if string(s.Peek()) == pathSep {
					s.Next()
				}

				//If pattern EOF
				if s.Peek() == scanner.EOF {
					regexPat += ".*"
				} else {
					//make sure we escape  path separator after **
					regexPat += "(.*" + escPath + ")?"
				}
			} else {
				regexPat += ".*"
			}
		} else if ch == '?' {
			// make sure ? escapes any character than path separator
			regexPat += "[^" + pathSep + "]"
		} else if ch == '.' || ch == '$' {
			regexPat += `\` + string(ch)
		} else if ch == '\\' {
			//handle windows path
			if pathSep == `\` {
				regexPat += escPath
				continue
			}
			if s.Peek() == scanner.EOF {
				regexPat += `\` + string(s.Next())
			} else {
				regexPat += `\`
			}
		} else {
			regexPat += string(ch)
		}
	}

	//end
	regexPat += "$"

	// Since the patterns are relative to the root, compile regex with dir root
	// prepended to it
	regexPat = "^" + directory + "/" + regexPat

	re, err := regexp.Compile(regexPat)

	if err != nil {
		return nil, err
	}

	return re, nil
}
