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
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"strings"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Unmarshal(out interface{}, fd FileDescriptor) error {
	file := path.Join(fd.BaseDir, fd.Org, fd.Space, strings.Join([]string{fd.Name, fd.Extension}, "."))
	data, err := os.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot open file: %s", file))
	}

	err = yaml.Unmarshal(data, out)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot unmarshal data: %v", err))
	}

	return nil
}

func (p *Parser) Marshal(in interface{}, fd FileDescriptor) error {
	b, err := yaml.Marshal(in)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot marshal data: %v", in))
	}

	d := NewFileSystemHelper()
	dir := path.Join(fd.BaseDir, fd.Org, fd.Space)
	err = d.Mkdir(dir)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot create directory: %s", dir))
	}

	file := path.Join(dir, strings.Join([]string{fd.Name, fd.Extension}, "."))
	err = os.WriteFile(file, b, 0755)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot save to file: %s", file))
	}

	return nil
}
