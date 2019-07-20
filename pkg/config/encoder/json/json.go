// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package json provides an encoder to de/encode JSON data.
package json

import (
	"encoding/json"
)

// Encoder represents a JSON configuration file encoder.
// See the official JSON specification and documentations for more details: https://json.org
type Encoder struct{}

// Encode encodes the given JSON data.
func (j Encoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Decode decodes the given JSON data.
func (j Encoder) Decode(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}

// NewJSONEncoder returns a new JSON Encoder.
func NewJSONEncoder() Encoder {
	return Encoder{}
}
