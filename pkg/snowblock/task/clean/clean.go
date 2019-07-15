// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package clean provides a task runner implementation check for broken symbolic links and automatically remove them if
// they point to the snowblock directory.
package clean

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"

	"github.com/arcticicestudio/snowsaw/pkg/api/snowblock"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
	"github.com/arcticicestudio/snowsaw/pkg/util/filesystem"
)

// Clean is a task runner check for broken symbolic links and automatically remove them if they point to the snowblock
// directory.
type Clean struct {
	absPaths         []string
	targets          []*target
	snowblockAbsPath string
}

type target struct {
	absPath   string
	isSymlink bool
	nodeInfo  os.FileInfo
	path      string
}

// GetTaskName returns the name of the task this runner can process.
func (c Clean) GetTaskName() string {
	return "clean"
}

// Run processes a task using the given task instructions.
// The snowblockAbsPath parameter is the absolute path of the snowblock used as contextual information.
func (c *Clean) Run(configuration snowblock.TaskConfiguration, snowblockAbsPath string) error {
	c.snowblockAbsPath = snowblockAbsPath

	// Try to assert the type of the given task configurations and process the paths if all values are converted
	// successfully.
	switch configType := configuration.(type) {
	// Handle the only support JSON data structure of type `array` that stores `string` values by converting
	// the given values to strings.
	case []interface{}:
		c.absPaths = []string{}
		c.targets = []*target{}
		for _, value := range configType {
			path, converted := value.(string)
			if !converted {
				prt.Debugf("Invalid clean configuration value %s of type %s",
					color.CyanString("%s", value), color.RedString("%T", value))
				return fmt.Errorf("invalid clean configuration value: %s", color.RedString("%s", value))
			}
			// Expand environment variables and special characters in the target paths,...
			expPath, expPathErr := filesystem.ExpandPath(path)
			if expPathErr != nil {
				return fmt.Errorf("could not expand target path %s: %v", color.CyanString(path), expPathErr)
			}
			var absPath string
			// ...ensure relative paths are dissolved from to absolute paths...
			if !filepath.IsAbs(expPath) {
				relToAbsPath, relToAbsPathErr := filepath.Abs(filepath.Join(c.snowblockAbsPath, expPath))
				if relToAbsPathErr != nil {
					return fmt.Errorf("could not dissolve clean target path relative to snowblock path: %v", relToAbsPathErr)
				}
				absPath = relToAbsPath
			} else {
				dissolvedPath, dissolvePathErr := filepath.Abs(expPath)
				if dissolvePathErr != nil {
					return fmt.Errorf("could not dissolve absolute clean target path: %v", dissolvePathErr)
				}
				absPath = dissolvedPath
			}
			c.absPaths = append(c.absPaths, absPath)
		}
		// ...and deduplicate possible duplicates to prevent to process and traverse same paths multiple times.
		prt.Debugf("Filtering possible duplicate clean targets: %s", color.YellowString("%v", c.absPaths))
		c.absPaths = removeDuplicatesTargets(c.absPaths)
		prt.Debugf("Processing deduplicated clean targets: %s", color.CyanString("%v", c.absPaths))
		if execErr := c.execute(); execErr != nil {
			return execErr
		}

	// Reject invalid or unsupported JSON data structures.
	default:
		prt.Debugf("unsupported clean configuration type: %s", color.RedString("%T", configType))
		return fmt.Errorf("unsupported clean configuration")
	}

	return nil
}

func (c *Clean) execute() error {
	for _, targetAbsPath := range c.absPaths {
		// Ignore targets where the directory or file does not exist...
		nodeInfo, nodeInfoErr := os.Lstat(targetAbsPath)
		if os.IsNotExist(nodeInfoErr) {
			prt.Debugf("Ignoring non-existent clean target: %s", color.RedString(targetAbsPath))
			continue
			// ...and fail if any error occurs while trying to describe the node at the given path.
		} else if nodeInfoErr != nil {
			return nodeInfoErr
		}

		t := &target{absPath: targetAbsPath, nodeInfo: nodeInfo, path: targetAbsPath}
		isSymlink, symlinkChkErr := filesystem.IsSymlink(t.absPath)
		if symlinkChkErr != nil {
			return symlinkChkErr
		}
		if isSymlink {
			t.isSymlink = true
		}
		c.targets = append(c.targets, t)
	}

	for _, t := range c.targets {
		// Handle the target when it is a symbolic link...
		if t.isSymlink {
			if brokenSymlinkErr := c.handleBrokenSnowblockSymlink(t.absPath); brokenSymlinkErr != nil {
				return brokenSymlinkErr
			}
			continue
		}

		// ...or traverse all nodes when it is a directory.
		if t.nodeInfo.IsDir() {
			nodes, nodesListErr := ioutil.ReadDir(t.absPath)
			if nodesListErr != nil {
				return fmt.Errorf("could not read clean target directory content: %s", color.RedString("%v", nodesListErr))
			}
			for _, targetNode := range nodes {
				nodeAbsPath := filepath.Join(t.absPath, targetNode.Name())
				isSymlink, symlinkChkErr := filesystem.IsSymlink(nodeAbsPath)
				if symlinkChkErr != nil {
					return symlinkChkErr
				}
				if isSymlink {
					if brokenSymlinkErr := c.handleBrokenSnowblockSymlink(nodeAbsPath); brokenSymlinkErr != nil {
						return brokenSymlinkErr
					}
				}
			}
		}
	}

	return nil
}

// isSnowblockSymlink checks if the symbolic link at the given absolute path is a broken link of a snowblock node.
// Returns any error that might occur during the process, nil otherwise.
func (c *Clean) handleBrokenSnowblockSymlink(absPath string) error {
	// Dissolve the absolute path of the symbolic link and remove it...
	destPath, destPathErr := os.Readlink(absPath)
	if destPathErr != nil {
		return fmt.Errorf("could not read symbolic link: %v", destPathErr)
	}
	if !filepath.IsAbs(destPath) {
		destAbsPath, destAbsPathErr := filepath.Abs(filepath.Join(filepath.Dir(absPath), destPath))
		if destAbsPathErr != nil {
			return fmt.Errorf("could not dissolve absolute path: %v", destAbsPathErr)
		}
		destPath = destAbsPath
	}
	nodeExists, nodeExistsErr := filesystem.NodeExists(destPath)
	if nodeExistsErr != nil {
		return nodeExistsErr
	}
	// ...when the underlying node does not exist...
	if !nodeExists {
		// ...and the path is a subdirectory of the snowblock directory.
		if strings.HasPrefix(destPath, c.snowblockAbsPath) {
			if removeErr := os.Remove(absPath); removeErr != nil {
				return removeErr
			}
			prt.Infof("Removed broken symbolic link: %s → %s",
				color.YellowString(absPath), color.RedString(destPath))
		}
	}

	return nil
}

// removeDuplicatesTargets removes all duplicate target paths.
func removeDuplicatesTargets(targets []string) []string {
	encountered := map[string]bool{}
	// Create a map of all unique targets...
	for t := range targets {
		encountered[targets[t]] = true
	}
	var result []string
	// ... and convert all keys from the map into a slice.
	for key := range encountered {
		result = append(result, key)
	}

	return result
}
