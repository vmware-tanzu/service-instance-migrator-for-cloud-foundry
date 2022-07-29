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

package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type MapDecoder struct {
	Config  *interface{}
	Decoder *mapstructure.Decoder
}

func NewMapDecoder(cfg interface{}) MapDecoder {
	decoderCfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &cfg,
		TagName:  "yaml",
	}

	decoder, err := mapstructure.NewDecoder(decoderCfg)
	if err != nil {
		log.Fatalf("failed to decode config, err: %s", err)
	}

	return MapDecoder{
		Config:  &cfg,
		Decoder: decoder,
	}
}

func (d MapDecoder) Decode(migration Migration, key string) interface{} {
	m := LookupMigrator(migration, key)
	if m != nil {
		if err := d.Decoder.Decode(m.Value); err != nil {
			log.Fatalf("failed to decode config, err: %s", err)
		}
	}

	return *d.Config
}

func LookupMigrator(migration Migration, key string) *Migrator {
	for _, v := range migration.Migrators {
		if v.Name == key {
			return &v
		}
	}
	return nil
}
