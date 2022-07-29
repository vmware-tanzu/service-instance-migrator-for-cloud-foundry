/* 
 *  Copyright 2022 VMware, Inc.
 *  
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package io

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . FileSystemOperations

type FileSystemOperations interface {
	Mkdir(dir string) error
	IsEmpty(name string) (bool, error)
	Exists(path string) (bool, error)
	Open(name string) (*os.File, error)
	Create(name string) (*os.File, error)
	Tar(src string, writers ...io.Writer) error
	Untar(dst string, r io.Reader) error
	Rename(re *regexp.Regexp, dir string, newName string) error
}

type FileSystemHelper struct{}

func NewFileSystemHelper() FileSystemHelper {
	return FileSystemHelper{}
}

func (h FileSystemHelper) Mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h FileSystemHelper) IsEmpty(name string) (bool, error) {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return true, nil
	}
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}

func (h FileSystemHelper) Exists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func (h FileSystemHelper) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (h FileSystemHelper) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (h FileSystemHelper) Tar(src string, writers ...io.Writer) error {
	return Tar(src, writers...)
}

func (h FileSystemHelper) Untar(dst string, r io.Reader) error {
	return Untar(dst, r)
}

func (h FileSystemHelper) Rename(re *regexp.Regexp, dir string, newName string) error {
	list, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, dirEntry := range list {
		if re.MatchString(dirEntry.Name()) {
			return os.Rename(filepath.Join(dir, dirEntry.Name()), filepath.Join(dir, newName))
		}
	}

	return nil
}
