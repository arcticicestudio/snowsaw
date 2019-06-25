// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package yaml provides an encoder to de/encode YAML data.
package yaml

import (
	"github.com/ghodss/yaml"
)

// Encoder represents a YAML configuration file encoder.
// See the underlying YAML (un)marshaller library for more details: https://github.com/ghodss/yaml
// Also see the official YAML specification and documentations: https://yaml.org
type Encoder struct{}

// Encode encodes the given YAML data.
func (y Encoder) Encode(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// Decode decodes the given YAML data.
func (y Encoder) Decode(d []byte, v interface{}) error {
	return yaml.Unmarshal(d, v)
}

// NewYamlEncoder returns a new YAML Encoder.
func NewYamlEncoder() Encoder {
	return Encoder{}
}
