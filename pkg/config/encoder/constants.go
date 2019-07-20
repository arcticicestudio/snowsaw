// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

package encoder

import (
	"github.com/arcticicestudio/snowsaw/pkg/config/encoder/json"
	"github.com/arcticicestudio/snowsaw/pkg/config/encoder/yaml"
)

var (
	// ExtensionMapping maps supported file extensions to their compatible encoders.
	ExtensionMapping = map[string]Encoder{
		ExtensionsJSON: json.NewJSONEncoder(),
		ExtensionsYaml: yaml.NewYamlEncoder(),
	}
	// ExtensionsJSON is the supported extension for files containing JSON data.
	ExtensionsJSON = "json"
	// ExtensionsYaml is the supported extension for files containing YAML data.
	ExtensionsYaml = "yml"
)
