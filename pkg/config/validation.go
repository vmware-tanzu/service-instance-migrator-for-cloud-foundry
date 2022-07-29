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
	"fmt"
	"strings"
)

type FieldError struct {
	Field string
	Msg   string
}

func (f FieldError) Error() string {
	return fmt.Sprintf("%s %s", f.Field, f.Msg)
}

func NewFieldError(field string, err error) error {
	switch newErr := err.(type) {
	case *FieldError:
		newErr.Field = fmt.Sprintf("%s.%s", field, newErr.Field)
		return newErr
	default:
		return &FieldError{
			Field: field,
			Msg:   err.Error(),
		}
	}
}

func NewFieldsError(fields []string, err error) error {
	switch newErr := err.(type) {
	case *FieldError:
		var newFields []string
		for _, f := range fields {
			field := fmt.Sprintf("%s.%s", f, newErr.Field)
			newFields = append(newFields, field)
		}
		newErr.Field = strings.Join(newFields, ",")
		return newErr
	default:
		return &FieldError{
			Field: strings.Join(fields, ", "),
			Msg:   err.Error(),
		}
	}
}
