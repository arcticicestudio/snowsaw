// +build ignore

// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

package main

import (
	"os"

	"github.com/magefile/mage/mage"
)

// Allows to run the project tasks without installing the mage binary.
// See https://magefile.org/zeroinstall for more details.
func main() { os.Exit(mage.Main()) }
