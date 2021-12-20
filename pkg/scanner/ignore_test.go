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
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

func TestDirectoryHasDockerIgnore(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "dir2")

	igScanner, err := NewOrDefault(dir, ".dockerignore")

	if err != nil {
		t.Error("hasDockerIgnore() = ", err)
		return
	}

	_, err = igScanner.Scan()

	if err != nil {
		t.Error("igScanner.Scan()= ", err)
		return
	}

	if len(ignorePatterns) == 0 {
		t.Errorf("The directory %s has '.dockerignore', but got it does not", dir)
		return
	}

	if got, want := len(ignorePatterns), int(9); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
		return
	}
}

func TestDirectoryHasNoDockerIgnore(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	dir := filepath.Join(wd, "testdata", "dir1")

	igScanner, err := NewOrDefault(dir, ".dockerignore")

	if err != nil {
		t.Error("hasDockerIgnore() = ", err)
		return
	}

	_, err = igScanner.Scan()

	if err != nil {
		t.Error("igScanner.Scan()= ", err)
		return
	}

	if got, want := len(ignorePatterns), int(3); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
		return
	}

}

func TestDockerIgnoredPatterns(t *testing.T) {
	expected := sets.NewString("lib", "*.md", "!README.md", "temp?", "target", "!target/*-runner.jar")
	expected.Insert(defaultPatterns...)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "dir2")

	igScanner, err := NewOrDefault(dir, ".dockerignore")

	if err != nil {
		t.Error("ignorablePatterns() = ", err)
	}

	_, err = igScanner.Scan()

	if err != nil {
		t.Error("igScanner.Scan()= ", err)
	}

	if got, want := len(ignorePatterns), int(9); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
	}

}

func TestEmptyDockerIgnoredPatterns(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "empty")

	igScanner, err := NewOrDefault(dir, ".dockerignore")

	if err != nil {
		t.Error("ignorablePatterns() = ", err)
	}

	_, err = igScanner.Scan()

	if err != nil {
		t.Error("igScanner.Scan()= ", err)
	}

	if got, want := len(ignorePatterns), int(3); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
	}

}

func TestIgnorables(t *testing.T) {

	eIncludes := sets.NewString([]string{
		".dockerignore",
		"README.md",
		"src",
		"src/main",
		"src/main/java",
		"src/main/java/One.java",
		"src/main/resources",
		"src/main/resources/application.properties",
		"tempABC",
		"pom.xml",
		"target/foo-runner.jar",
	}...)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "dir2")

	ignoreScanner, err := NewOrDefault(dir, ".dockerignore")

	if err != nil {
		t.Error("isIgnorable() = ", err)
	}

	incls, err := ignoreScanner.Scan()

	sIncludes := sets.NewString(incls...)

	if err != nil {
		t.Error("ignoreScanner.Scan() = ", err)
	}

	if got, want := len(incls), int(11); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
	}

	if !eIncludes.Equal(sIncludes) {
		t.Errorf("Expeacted - Actual  : %s", sets.String.Difference(eIncludes, sIncludes))
	}
}

func TestStarAndIgnoreables(t *testing.T) {

	eIncludes := sets.NewString([]string{
		"README.md",
		"target/foo-runner.jar",
		"target/lib/one.jar",
		"target/quarkus-app/one.txt",
	}...)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}

	dir := filepath.Join(wd, "testdata", "starignore")

	ignoreScanner, err := NewOrDefault(dir, ".dockerignore")

	if err != nil {
		t.Error("isIgnorable() = ", err)
		return
	}

	incls, err := ignoreScanner.Scan()

	sIncludes := sets.NewString(incls...)

	if err != nil {
		t.Error("ignoreScanner.Scan() = ", err)
		return
	}

	if got, want := len(incls), int(4); got != want {
		t.Errorf("Patterns() = %d, wanted %d", got, want)
		return
	}

	if !eIncludes.Equal(sIncludes) {
		t.Errorf("Includes() = %s, wanted %s", sIncludes, eIncludes)
		return
	}

}
