// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// Package link provides a task runner implementation to create symbolic links for files and directories.
package link

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"

	"github.com/arcticicestudio/snowsaw/pkg/api/snowblock"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
	"github.com/arcticicestudio/snowsaw/pkg/util/filesystem"
)

const (
	// DefaultHostName is the name for host mappings that will apply to all host.
	// To prevent possible collisions with actual host names, it is a single minus character.
	// As defined in the specification this is not a valid hostname since the name should not start or end with a minus.
	// See "RFC 1123" and https://en.wikipedia.org/wiki/Hostname#Restrictions_on_valid_hostnames for more details about
	// restrictions and valid names.
	DefaultHostName = "-"
)

// Link is a task runner to create symbolic links for files and directories.
type Link struct {
	config           *config
	destAbsPath      string
	destPath         string
	snowblockAbsPath string
	srcAbsPath       string
}

type config struct {
	Create   bool              `json:"create" yaml:"create"`
	Force    bool              `json:"force" yaml:"force"`
	Hosts    map[string]string `json:"hosts,flow" yaml:"hosts,flow"`
	Path     string            `json:"path" yaml:"path"`
	Relative bool              `json:"relative" yaml:"relative"`
	Relink   bool              `json:"relink" yaml:"relink"`
}

// GetTaskName returns the name of the task this runner can process.
func (l Link) GetTaskName() string {
	return "link"
}

// Run processes a task using the given task instructions.
// The snowblockAbsPath parameter is the absolute path of the snowblock used as contextual information.
func (l *Link) Run(configuration snowblock.TaskConfiguration, snowblockAbsPath string) error {
	l.snowblockAbsPath = snowblockAbsPath

	// Try to convert given task configurations...
	configMap, ok := configuration.(map[string]interface{})
	if !ok {
		prt.Debugf("invalid link configuration type: %s", color.RedString("%T", configuration))
		return errors.New("malformed link configuration")
	}

	// ...and handle the possible types.
	for destPath, configData := range configMap {
		l.destAbsPath = ""
		l.srcAbsPath = ""

		switch configType := configData.(type) {
		// Handle JSON `null` value configurations used to omit duplicate definitions when the source path equals the
		// destination path.
		// Uses the base name of the destination path and trims a leading dot character if present.
		case nil:
			sourceBaseName := strings.TrimPrefix(filepath.Base(destPath), ".")
			l.config = &config{Path: sourceBaseName}
			l.destPath = destPath
			if execErr := l.execute(); execErr != nil {
				return execErr
			}

		// Handle JSON `object` configurations used to define more link options.
		// Uses the base name of the destination path with leading dot character trimmed if path is not specified.
		case map[string]interface{}:
			c := new(config)
			if err := mapstructure.Decode(configType, &c); err != nil {
				return err
			}
			l.destPath = destPath
			if c.Path == "" {
				c.Path = strings.TrimPrefix(filepath.Base(destPath), ".")
			}
			l.config = c
			if execErr := l.execute(); execErr != nil {
				return execErr
			}

		// Handle JSON `string` configurations used to only specify the source path.
		case string:
			l.config = &config{Path: configType}
			l.destPath = destPath
			if execErr := l.execute(); execErr != nil {
				return execErr
			}

		// Reject invalid or unsupported JSON data structures.
		default:
			prt.Debugf("unsupported destination type: %s", color.RedString("%T", configType))
			return fmt.Errorf("unsupported link configuration: %s", color.CyanString(destPath))
		}
	}

	return nil
}

func (l *Link) execute() error {
	// Check if the current and/or default host is listed in the target mapping, otherwise stop processing.
	isTargetHost, hostCheckErr := l.isTargetHost()
	if hostCheckErr != nil {
		return hostCheckErr
	}
	if !isTargetHost {
		return nil
	}

	// Dissolve the source to an absolute path.
	srcAbsPath, srcToAbsPathErr := filepath.Abs(filepath.Join(l.snowblockAbsPath, l.config.Path))
	if srcToAbsPathErr != nil {
		return srcToAbsPathErr
	}
	l.srcAbsPath = srcAbsPath

	// Fail fast if the source node does not exist.
	if sourceNodeExistsErr := l.checkSourceNode(); sourceNodeExistsErr != nil {
		return sourceNodeExistsErr
	}

	// Expand the destination path to dissolve environment variables and special characters like tilde...
	expDestPath, pathExpandErr := filesystem.ExpandPath(l.destPath)
	if pathExpandErr != nil {
		return pathExpandErr
	}

	if !filepath.IsAbs(expDestPath) {
		l.destAbsPath = filepath.Join(l.snowblockAbsPath, expDestPath)
	} else {
		l.destAbsPath = expDestPath
	}

	destNodeExists, nodeExistErr := filesystem.NodeExists(l.destAbsPath)
	if nodeExistErr != nil {
		return nodeExistErr
	}
	// Check if the destination node already exists,...
	if destNodeExists {
		isSymlink, symlinkCheckErr := filesystem.IsSymlink(l.destAbsPath)
		if symlinkCheckErr != nil {
			return symlinkCheckErr
		}
		// ...evaluate if it is a symbolic link,...
		if isSymlink {
			symlinkDest, symlinkReadErr := os.Readlink(l.destAbsPath)
			if symlinkReadErr != nil {
				return symlinkReadErr
			}
			symlinkDestAbs, symlinkDestAbsErr := filepath.Abs(symlinkDest)
			if symlinkDestAbsErr != nil {
				return symlinkDestAbsErr
			}

			// ...and continue with processing when running in relinking mode,...
			if l.config.Relink {
				prt.Warnf("%s already existing symbolic link: %s",
					color.YellowString("Relinking"), color.CyanString(l.destAbsPath))
				if removeErr := os.Remove(l.destAbsPath); removeErr != nil {
					return removeErr
				}
				if parentDirErr := l.handleParentDirStructure(); parentDirErr != nil {
					return parentDirErr
				}
				if symlinkCreationError := l.createSymbolicLink(); symlinkCreationError != nil {
					return symlinkCreationError
				}
				return nil
			}

			// ...or stop processing when it already links to the correct destination,...
			if symlinkDestAbs == l.srcAbsPath {
				prt.Infof("Skipped already existing link: %s", color.CyanString(l.destAbsPath))
				return nil
			}

			// ...otherwise only if force linking is enabled.
			if l.config.Force {
				prt.Warnf("%s of already existing symbolic link: %s",
					color.YellowString("Forced linking"), color.CyanString(l.destAbsPath))
				if removeErr := os.Remove(l.destAbsPath); removeErr != nil {
					return removeErr
				}
				if parentDirErr := l.handleParentDirStructure(); parentDirErr != nil {
					return parentDirErr
				}
				if symlinkCreationError := l.createSymbolicLink(); symlinkCreationError != nil {
					return symlinkCreationError
				}
				return nil
			}

			return fmt.Errorf("symbolic link already exists: %s ← %s", symlinkDest, l.destAbsPath)
		}

		// Always process the task in force mode when the destination is an already existing file or directory,...
		if l.config.Force {
			prt.Warnf("%s of already existing symbolic link: %s",
				color.YellowString("Forced linking"), color.CyanString(l.destAbsPath))
			if removeErr := os.Remove(l.destAbsPath); removeErr != nil {
				return removeErr
			}
			if parentDirErr := l.handleParentDirStructure(); parentDirErr != nil {
				return parentDirErr
			}
			if symlinkCreationError := l.createSymbolicLink(); symlinkCreationError != nil {
				return symlinkCreationError
			}
			return nil
		}

		return fmt.Errorf("file or directory already exists: %s", l.destAbsPath)
	}

	// ...otherwise only when all previous conditions are not met.
	if parentDirErr := l.handleParentDirStructure(); parentDirErr != nil {
		return parentDirErr
	}
	if symlinkCreateErr := l.createSymbolicLink(); symlinkCreateErr != nil {
		return symlinkCreateErr
	}

	return nil
}

// checkSourceNode checks if the source node at the given path exists, otherwise returns the corresponding error.
func (l *Link) checkSourceNode() error {
	sourceNodeExists, err := filesystem.NodeExists(l.srcAbsPath)
	if err != nil {
		return err
	}
	if !sourceNodeExists {
		return fmt.Errorf("no such file or directory: %s", l.config.Path)
	}

	return nil
}

// createSymbolicLink creates the symbolic link based on the value of the task option that allows to use relative
// instead of absolute paths.
// If any error occurs it will be returned, otherwise returns nil.
func (l *Link) createSymbolicLink() error {
	if l.config.Relative {
		srcRelPath, srcRelPathErr := filepath.Rel(filepath.Dir(l.destAbsPath), l.srcAbsPath)
		if srcRelPathErr != nil {
			return fmt.Errorf("could not dissolve path of source relative to destination directory: %v", srcRelPathErr)
		}

		if relSymlinkErr := os.Symlink(srcRelPath, l.destAbsPath); relSymlinkErr != nil {
			return relSymlinkErr
		}
		prt.Infof("Created relative symbolic link: %s → %s", color.CyanString(l.srcAbsPath), color.BlueString(l.srcAbsPath))
		return nil
	}

	if symlinkErr := os.Symlink(l.srcAbsPath, l.destAbsPath); symlinkErr != nil {
		return symlinkErr
	}
	prt.Infof("Created symbolic link: %s → %s", color.BlueString(l.destAbsPath), color.CyanString(l.srcAbsPath))
	return nil
}

// handleParentDirStructure checks if the required parent directory structure for the symbolic links exists,
// otherwise creates it if the corresponding task option has been specified.
// If any error occurs it will be returned, otherwise returns nil.
func (l *Link) handleParentDirStructure() error {
	destParentDirs := filepath.Dir(l.destAbsPath)
	destParentDirsExist, nodeExistErr := filesystem.DirExists(destParentDirs)
	if nodeExistErr != nil {
		return nodeExistErr
	}
	if !destParentDirsExist {
		if l.config.Create {
			if mkdirErr := os.MkdirAll(destParentDirs, os.ModePerm); mkdirErr != nil {
				return mkdirErr
			}
			prt.Debugf("Created parent directory structure: %s", destParentDirs)
		} else {
			return fmt.Errorf("no such directory: %s", destParentDirs)
		}
	}

	return nil
}

// isTargetHost checks if the current and/or default host is listed in the target mapping.
// It returns the host specific source path, otherwise if an error occurs an empty string along with the error.
func (l *Link) isTargetHost() (bool, error) {
	if len(l.config.Hosts) > 0 {
		hostname, err := os.Hostname()
		if err != nil {
			return false, fmt.Errorf("failed to determine hostname: %v", err)
		}
		sourcePath, isTargetHost := l.config.Hosts[hostname]
		sourcePathDefaultHost, isDefaultTargetHost := l.config.Hosts[DefaultHostName]
		if !isTargetHost && !isDefaultTargetHost {
			prt.Debugf("Skipped host specific link not matching current host %s: %s",
				color.BlueString(hostname), color.CyanString(l.destPath))
			return false, nil
		}

		// Use the default target host if specified...
		if isDefaultTargetHost {
			prt.Debugf("Found host mapping for default target: %s", color.CyanString(sourcePathDefaultHost))
			l.config.Path = sourcePathDefaultHost
		}
		// ...and override when exact host name has also been specified.
		if isTargetHost {
			prt.Debugf("Using source path for exact host name match %s: %s",
				color.BlueString(hostname), color.CyanString(sourcePath))
			l.config.Path = sourcePath
		}

		return true, nil
	}

	if l.config.Path != "" {
		return true, nil
	}

	return false, nil
}
