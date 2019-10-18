// +build mage

// Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
// Copyright (C) 2017-present Sven Greb <development@svengreb.de>
//
// Project:    snowsaw
// Repository: https://github.com/arcticicestudio/snowsaw
// License:    MIT

// Author: Arctic Ice Studio <development@arcticicestudio.com>
// Author: Sven Greb <development@svengreb.de>
// Since: 0.4.0

// The main build and development toolchain of the snowsaw project powered by "Mage".
// See the official documentations for more details:
//   https://magefile.org
//   https://github.com/magefile/mage

package main

import (
	"bytes"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	"github.com/arcticicestudio/snowsaw/pkg/config"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
)

// appVersion stores information and metadata about the application version.
type appVersion struct {
	// Version is the application SemVer version.
	*semver.Version
	// GitCommitsAhead is the count of commits ahead to the latest Git version tag in the current branch.
	GitCommitsAhead int
	// GitCommitHash is the hash of the latest commit in the current branch.
	GitCommitHash plumbing.Hash
	// GitLatestVersionTag is the latest Git version tag in the current branch.
	GitLatestVersionTag *plumbing.Reference
}

// buildDependency represents a build dependency like a tool used to build or develop the project.
type buildDependency struct {
	// BinaryExecPath is the path of the binary executable.
	BinaryExecPath string
	// BinaryName is the name of the binary.
	BinaryName string
	// ModuleName is the name of the module.
	ModuleName string
	// ModuleVersion is the version of the module including prefixes like "v" if any.
	ModuleVersion string
}

const (
	// The output directory for all builds.
	buildDir = "build"

	// The name of the environment variable to define a space-separated custom list of build tags.
	snowsawEnvBuildTags = "SNOWSAW_BUILD_TAGS"

	// The name of the environment variable to define a space-separated custom list of platform targets.
	snowsawEnvCrossPlatformTargets = "SNOWSAW_CROSS_PLATFORM_TARGETS"

	// The name of the environment variable to define a custom path to the Go executable.
	snowsawEnvGoExec = "SNOWSAW_GOEXEC"

	// The file name for the test coverage profile report.
	testCoverageOutputFileName = "coverage.out"
)

var (
	// Arguments for the `-asmflags` flag to pass on each `go tool asm` invocation.
	asmFlags = "all=-trimpath=$PROJECT_ROOT"

	// The name template for cross-compiled binaries.
	crossCompileNameTemplate = config.ProjectName + "-{{.OS}}-{{.Arch}}"

	// The names of cross-compile platform targets.
	// See the official Go tool `dist` to get a list of supported platforms: `go tool dist list`
	// Also see the source of the `build` command: https://github.com/golang/go/blob/master/src/cmd/dist/build.go
	crossCompileTargetPlatforms = []string{
		"darwin/amd64",
		"linux/amd64",
		"windows/amd64",
	}

	// The tool used to cross-compile the project asynchronously.
	// See https://github.com/mitchellh/gox for more details.
	crossCompileTool = &buildDependency{
		BinaryName:    "gox",
		ModuleName:    "github.com/mitchellh/gox",
		ModuleVersion: "v1.0.1",
	}

	// devToolManager is the tool to install and run all used project tools and applications with Go's module mode.
	// This is necessary because the Go toolchain currently doesn't support the handling of local or global project tool
	// dependencies in module mode without "polluting" the project's Go module file (go.mod).
	//
	// See the FAQ/documentations of "gobin" as well as issue references for more details about the tool and its purpose:
	// https://github.com/myitcv/gobin/wiki/FAQ
	//
	// For more details about the status of proposed official Go toolchain solutions and workarounds see the following
	// references:
	//   - https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
	//   - https://github.com/golang/go/issues/27653
	//   - https://github.com/golang/go/issues/25922
	devToolManager = &buildDependency{
		BinaryName:    "gobin",
		ModuleName:    "github.com/myitcv/gobin",
		ModuleVersion: "v0.0.13",
	}

	// The tool used to format all Go source files.
	// See https://godoc.org/golang.org/x/tools/cmd/goimports for more details.
	formatTool = &buildDependency{
		BinaryName: "goimports",
		ModuleName: "golang.org/x/tools/cmd/goimports",
	}

	// Arguments for the `-gcflags` flag to pass on each `go tool compile` invocation.
	gcFlags = "all=-trimpath=$PROJECT_ROOT"

	// The path of the Go executable either set to a custom value or Go's default path.
	goExec string

	// The GOPATH either set to a custom value or Go's default path.
	goPath string

	// Arguments for the `-ldflags` flag to pass on each `go tool link` invocation.
	ldFlags = "-X $PACKAGE_NAME/pkg/config.BuildDateTime=$BUILD_DATE_TIME" +
		" -X $PACKAGE_NAME/pkg/config.Version=$VERSION"

	// The tool used to lint all Go source files.
	// This is the same tool used by the https://golangci.com service that is also integrated in snowsaw's CI/CD pipeline.
	// See https://github.com/golangci/golangci-lint for more details.
	lintTool = &buildDependency{
		BinaryName:    "golangci-lint",
		ModuleName:    "github.com/golangci/golangci-lint/cmd/golangci-lint",
		ModuleVersion: "v1.19.1",
	}

	// The output directory for reports like test coverage.
	reportsDir = filepath.Join(buildDir, "reports")

	// The build tags that can be specified through the `SNOWSAW_BUILD_TAGS` environment variable.
	tags []string

	// The flag for the coverage profile from the `go test` command.
	testCoverageProfileFlag string
)

func init() {
	customExecPath, isCustomExecPathSet := os.LookupEnv(snowsawEnvGoExec)
	if isCustomExecPathSet {
		prt.Infof("Running with user-defined Go executable: %s", color.CyanString(customExecPath))
		goExec = customExecPath
	} else {
		goExecPath, err := exec.LookPath("go")
		if err != nil {
			prt.Errorf("Couldn't determine path to Go executable, make sure it is available on PATH!")
			os.Exit(1)
		}
		goExec = goExecPath
	}

	value, isGoPathSet := os.LookupEnv("GOPATH")
	if !isGoPathSet {
		prt.Warnf(
			"%s environment variable not set, falling back to Go's default path: %s",
			color.CyanString("GOPATH"),
			color.CyanString(build.Default.GOPATH))
		goPath = build.Default.GOPATH
	}
	goPath = value
}

// Bootstrap bootstraps the local development environment by installing the required tools and build dependencies.
func Bootstrap() {
	mg.SerialDeps(bootstrap)
}

// Build compiles the project in development mode for the current OS and architecture type.
func Build() {
	mg.SerialDeps(clean, compile)
}

// Clean removes previous development and distribution builds from the project root.
func Clean() {
	mg.SerialDeps(clean)
}

// Dist builds the project in production mode for the current platform.
// It trims paths of the current working directory and injects interpolated application metadata like build
// version information via LDFLAGS.
// Run `strings <PATH_TO_BINARY> | grep "$PWD"` to verify that all paths have been successfully stripped.
func Dist() {
	mg.SerialDeps(validateDevTools, clean, compileProd)
}

// DistCrossPlatform builds the project in production mode for cross-platform distribution.
// This includes all steps from the current platform distribution/production task `Dist`,
// but instead builds for all configured OS/architecture types.
func DistCrossPlatform() {
	mg.SerialDeps(validateDevTools, clean, compileProdCross)
}

// DistCrossPlatformOpt builds the project in production mode for cross-platform distribution with optimizations.
// This includes all steps from the cross-platform distribution task `DistCrossPlatform` and additionally removes all
// debug metadata to shrink the memory overhead and file size as well as reducing the chance for possible security
// related problems due to enabled development features and leaked debug information.
func DistCrossPlatformOpt() {
	mg.SerialDeps(validateDevTools, clean, compileProdCrossOpt)
}

// DistOpt builds the project in production mode with optimizations like minification and debug symbol stripping.
// This includes all steps from the production build task `Dist` and additionally removes all debug metadata to shrink
// the memory overhead and file size as well as reducing the chance for possible security related problems due to
// enabled development features and leaked debug information.
func DistOpt() {
	mg.SerialDeps(validateDevTools, clean, compileProdOpt)
}

// Format searches all project Go source files and formats them according to the Go code styleguide.
func Format() {
	mg.SerialDeps(validateDevTools, runGoImports)
}

// Lint runs all linters configured and executed through `golangci-lint`.
// See the `.golangci.yml` configuration file and official GolangCI documentations at https://golangci.com
// and https://github.com/golangci/golangci-lint for more details.
func Lint() {
	mg.SerialDeps(validateDevTools, runGolangCILint)
}

// Test runs all unit tests with enabled race detection.
func Test() {
	mg.SerialDeps(unitTests)
}

// TestCover runs all unit tests with with coverage reports and enabled race detection.
func TestCover() {
	mg.SerialDeps(clean)
	// Ensure the required directory structure exists, `go test` doesn't create it automatically.
	createDirectoryStructure(reportsDir)
	testCoverageProfileFlag = fmt.Sprintf("-coverprofile=%s", filepath.Join(reportsDir, testCoverageOutputFileName))
	mg.SerialDeps(unitTests)
}

// TestIntegration runs all integration tests with enabled race detection.
func TestIntegration() {
	mg.SerialDeps(integrationTests)
}

func bootstrap() {
	prt.Infof("Bootstrapping development tool/dependency manager %s",
		color.CyanString("%s@%s", devToolManager.ModuleName, devToolManager.ModuleVersion))
	cmdInstallGobin := exec.Command(goExec, "get", "-u",
		fmt.Sprintf("%s@%s", devToolManager.ModuleName, devToolManager.ModuleVersion))
	// Run the installation outside of the project root directory to prevent the pollution of the project's Go module
	// file.
	// This is a necessary workaround until the Go toolchain is able to install packages globally without
	// updating the module file when the "go get" command is run from within the project root directory.
	// See https://github.com/golang/go/issues/30515 for more details or more details and proposed solutions
	// that might be added to Go's build tools in future versions.
	cmdInstallGobin.Dir = os.TempDir()
	cmdInstallGobin.Env = os.Environ()
	// Explicitly enable "module" mode when installing the dev tool manager to allow to use pinned module version.
	cmdInstallGobin.Env = append(cmdInstallGobin.Env, "GO111MODULE=on")
	if gobinInstallErr := cmdInstallGobin.Run(); gobinInstallErr != nil {
		prt.Errorf("Failed to install required development tool/dependency manager %s:\n  %s",
			color.CyanString("%s@%s", devToolManager.ModuleName, devToolManager.ModuleVersion),
			color.RedString("%s", gobinInstallErr))
		os.Exit(1)
	}

	prt.Infof("Bootstrapping required development tools/dependencies:")
	for _, bd := range []*buildDependency{crossCompileTool, formatTool, lintTool} {
		modulePath := bd.ModuleName
		// If the non-module dependency is not installed yet, install it normally into the $GOBIN path,...
		if bd.ModuleVersion == "" {
			fmt.Println(color.CyanString("  %s", modulePath))
			if installErr := sh.Run(devToolManager.BinaryName, "-u", modulePath); installErr != nil {
				prt.Errorf("Failed to install required development tool/dependency %s:\n  %s",
					color.CyanString(modulePath), color.RedString("%s", installErr))
				os.Exit(1)
			}
			continue
		}

		// ...otherwise install into "gobin" binary cache.
		modulePath = fmt.Sprintf("%s@%s", bd.ModuleName, bd.ModuleVersion)
		fmt.Println(color.CyanString("  %s", modulePath))
		if installErr := sh.Run(devToolManager.BinaryName, "-u", modulePath); installErr != nil {
			prt.Errorf("Failed to install required development tool/dependency %s:\n  %s",
				color.CyanString(modulePath), color.RedString("%s", installErr))
			os.Exit(1)
		}
	}

	prt.Successf("Successfully bootstrapped required development tools/dependencies")
}

func clean() {
	if err := os.RemoveAll(buildDir); err != nil {
		prt.Errorf("Failed to clean up project directory: %v", err)
		os.Exit(1)
	}
	prt.Infof("Removed previous build directory: %s", color.CyanString(buildDir))
}

func compile() {
	prt.Infof("Compiling package in development mode: %s", color.GreenString(config.PackageName))
	prepareBuildTags()
	buildFlags := []string{
		"-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")),
		"-ldflags", ldFlags,
		"-asmflags", asmFlags,
		"-gcflags", gcFlags}
	runGoBuild(getEnvFlags(), buildFlags...)
}

func compileProd() {
	prt.Infof("Compiling package in production mode: %s", color.GreenString(config.PackageName))
	prepareBuildTags()
	buildFlags := []string{
		"-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")),
		"-ldflags", ldFlags,
		"-asmflags", asmFlags,
		"-gcflags", gcFlags}
	runGoBuild(getEnvFlags(), buildFlags...)
}

func compileProdCross() {
	prt.Infof("Cross compiling package %s in production mode", color.GreenString(config.PackageName))
	prepareBuildTags()
	buildFlags := []string{
		"-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")),
		"-ldflags", ldFlags,
		"-asmflags", asmFlags,
		"-gcflags", gcFlags}
	runGox(getEnvFlags(), buildFlags...)
}

func compileProdCrossOpt() {
	prt.Infof("Cross compiling package %s in production mode with optimizations", color.GreenString(config.PackageName))
	prepareBuildTags()
	ldFlags += "-s -w"
	buildFlags := []string{
		"-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")),
		"-ldflags", ldFlags,
		"-asmflags", asmFlags,
		"-gcflags", gcFlags}
	runGox(getEnvFlags(), buildFlags...)
}

func compileProdOpt() {
	prt.Infof("Compiling module in production mode with optimizations: %s",
		color.GreenString(config.PackageName))
	ldFlags += "-s -w"
	buildFlags := []string{"-ldflags", ldFlags, "-asmflags", asmFlags, "-gcflags", gcFlags}
	runGoBuild(getEnvFlags(), buildFlags...)
}

func createDirectoryStructure(paths ...string) {
	prt.Infof("Creating required directory structure: %s", color.CyanString("%v", paths))
	for _, path := range paths {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			prt.Warnf("Failed to create required directory structure %s: %v", color.CyanString(path), err)
		}
	}
}

// getAppVersionFromGit assembles the version of the application from the metadata of the Git repository.
// It searches for the latest SemVer (https://semver.org) compatible version tag in the current branch and falls back to
// the default version from the application configuration if none is found.
// If at least one tag is found but it is not the latest commit of the current branch, the build metadata will be
// appended, consisting of the amount of commits ahead and the shortened reference hash (8 digits) of the latest commit
// from the current branch.
// This function is a early implementation of the Git `describe` command because support in `go-git` has not been
// implemented yet. See the full compatibility comparision documentation with Git at
// https://github.com/src-d/go-git/blob/master/COMPATIBILITY.md as well as the proposed Git `describe` command
// implementation at https://github.com/src-d/go-git/pull/816 for more details.
func getAppVersionFromGit() (*appVersion, error) {
	// Open the repository in the current working directory.
	repo, repoOpenErr := git.PlainOpen(".")
	if repoOpenErr != nil {
		return nil, repoOpenErr
	}

	// Find the latest commit reference of the current branch.
	branchRefs, repoBranchErr := repo.Branches()
	if repoBranchErr != nil {
		return nil, repoBranchErr
	}
	headRef, repoHeadErr := repo.Head()
	if repoHeadErr != nil {
		return nil, repoHeadErr
	}
	var currentBranchRef plumbing.Reference
	branchRefIterErr := branchRefs.ForEach(func(branchRef *plumbing.Reference) error {
		if branchRef.Hash() == headRef.Hash() {
			currentBranchRef = *branchRef
			return nil
		}
		return nil
	})
	if branchRefIterErr != nil {
		return nil, branchRefIterErr
	}

	// Find all commits in the repository starting from the HEAD of the current branch.
	commitIterator, commitIterErr := repo.Log(&git.LogOptions{
		From:  currentBranchRef.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if commitIterErr != nil {
		return nil, commitIterErr
	}

	// Query all tags and store them in a temporary map.
	tagIterator, repoTagsErr := repo.Tags()
	if repoTagsErr != nil {
		return nil, repoTagsErr
	}
	repoTags := make(map[plumbing.Hash]*plumbing.Reference)
	tagIterErr := tagIterator.ForEach(func(tag *plumbing.Reference) error {
		if tagObject, tagObjectErr := repo.TagObject(tag.Hash()); tagObjectErr == nil {
			// Only include tags that have a valid SemVer version format.
			if _, semVerParseErr := semver.NewVersion(tag.Name().Short()); semVerParseErr == nil {
				repoTags[tagObject.Target] = tag
			}
		} else {
			repoTags[tag.Hash()] = tag
		}
		return nil
	})
	tagIterator.Close()
	if tagIterErr != nil {
		return nil, tagIterErr
	}

	type describeCandidate struct {
		ref       *plumbing.Reference
		annotated bool
		distance  int
	}
	var tagCandidates []*describeCandidate
	var tagCandidatesFound int
	var tagCount = -1
	var lastCommit *object.Commit

	// Search for maximal 10 (Git default) suitable tag candidates in all commits of the current branch.
	for {
		var candidate = &describeCandidate{annotated: false}
		tagCommitIterErr := commitIterator.ForEach(func(commit *object.Commit) error {
			lastCommit = commit
			tagCount++
			if tagReference, ok := repoTags[commit.Hash]; ok {
				delete(repoTags, commit.Hash)
				candidate.ref = tagReference
				hash := tagReference.Hash()
				if !bytes.Equal(commit.Hash[:], hash[:]) {
					candidate.annotated = true
				}
				return storer.ErrStop
			}
			return nil
		})
		if tagCommitIterErr != nil {
			return nil, tagCommitIterErr
		}

		if candidate.annotated {
			if tagCandidatesFound < 10 {
				candidate.distance = tagCount
				tagCandidates = append(tagCandidates, candidate)
			}
			tagCandidatesFound++
		}

		if tagCandidatesFound > 10 || len(tags) == 0 {
			break
		}
	}

	// Use the version from the application configuration by default or...
	semVersion, semVerErr := semver.NewVersion(config.Version)
	version := &appVersion{Version: semVersion}
	if semVerErr != nil {
		return nil, fmt.Errorf("failed to parse default version from application configuration: %s", semVerErr)
	}
	if len(tagCandidates) == 0 {
		prt.Infof("No Git tag found, using defined version %s as fallback", color.CyanString(config.Version))
		// ...the latest Git tag from the current branch if at least one has been found.
	} else {
		semVersion, semVerErr = semver.NewVersion(tagCandidates[0].ref.Name().Short())
		version = &appVersion{Version: semVersion}
		if semVerErr != nil {
			return nil, fmt.Errorf("failed to parse version from Git tag %s: %s",
				tagCandidates[0].ref.Name().Short(), semVerErr)
		}
	}
	// Add additional version information if the latest commit of the current branch is not the found tag.
	if len(tagCandidates) != 0 && tagCandidates[0].distance > 0 {
		// If not included in the tag already, append metadata consisting of the amount of commit(s) ahead and the shortened
		// commit hash (8 digits) of the latest commit.
		buildMetaData := fmt.Sprintf("%s.%s", strconv.Itoa(tagCandidates[0].distance), currentBranchRef.Hash().String()[:8])
		if version.Metadata() != "" {
			metadataVersion, err := version.SetMetadata(fmt.Sprintf("%s-%s", version.Metadata(), buildMetaData))
			if err != nil {
				return nil, err
			}
			version.Version = &metadataVersion
		} else {
			metadataVersion, err := version.SetMetadata(buildMetaData)
			if err != nil {
				return nil, err
			}
			version.Version = &metadataVersion
		}

		version.GitCommitsAhead = tagCandidates[0].distance
		version.GitCommitHash = currentBranchRef.Hash()
		version.GitLatestVersionTag = tagCandidates[0].ref
		prt.Infof("Using latest Git commit %s, %s commit(s) ahead of %s",
			color.CyanString(version.GitCommitHash.String()[:8]),
			color.CyanString(strconv.Itoa(version.GitCommitsAhead)),
			color.CyanString("%s", version.GitLatestVersionTag.Name().Short()))
	} else {
		prt.Infof("Using Git tag %s as application version", color.CyanString(version.Original()))
	}

	return version, nil
}

// getEnvFlags returns environment variables storing metadata the build time, Git version tags and commit checksum.
func getEnvFlags() map[string]string {
	buildDate := time.Now().Format(time.RFC3339)

	pwd, err := os.Getwd()
	if err != nil {
		prt.Errorf("Could not determine project root path: %v", err)
		os.Exit(1)
	}

	version, versionErr := getAppVersionFromGit()
	if versionErr != nil {
		prt.Errorf("Failed to assemble application version: %s", versionErr)
		os.Exit(1)
	}

	prt.Infof(
		"Injecting %s:\n"+
			"  Build Date: %s\n"+
			"  Version: %s",
		color.BlueString("LDFLAGS"), color.CyanString(buildDate), color.CyanString(version.String()))

	prt.Infof(
		"Injecting %s:\n"+
			"  -trimpath: %s",
		color.BlueString("ASMFLAGS"), color.CyanString(pwd))

	prt.Infof(
		"Injecting %s:\n"+
			"  -trimpath: %s",
		color.BlueString("GCFLAGS"), color.CyanString(pwd))

	return map[string]string{
		"BUILD_DATE_TIME": buildDate,
		"PACKAGE_NAME":    config.PackageName,
		"PROJECT_ROOT":    pwd,
		"VERSION":         version.String()}
}

// getExecutablePath returns the path to the executable for the given package/module.
// When the "resolveWithGobin" parameter is set to true, the path will be resolved from the "gobin" binary cache.
func getExecutablePath(name string, resolveWithGobin bool) (string, error) {
	if resolveWithGobin {
		return sh.Output(devToolManager.BinaryName, "-p", "-nonet", name)
	}
	return exec.LookPath(name)
}

// prepareBuildTags reads custom build tags defined by the user through the `SNOWSAW_BUILD_TAGS` environment
// variable and appends them together with all additionally passed tags to the global `tags` slice.
// Returns `true` if custom build tags have been loaded, `false` otherwise.
func prepareBuildTags(additionalTags ...string) bool {
	customTags, hasTags := os.LookupEnv(snowsawEnvBuildTags)
	if hasTags {
		prt.Infof("Picking up build tags from %s environment variable: %v",
			color.CyanString(snowsawEnvBuildTags), color.New(color.FgCyan, color.Bold).Sprint(customTags))
		tags = append(tags, strings.Split(customTags, " ")...)
	}
	tags = append(tags, additionalTags...)
	return hasTags
}

func unitTests() {
	prt.Infof("Running unit tests with enabled race detection")
	prepareBuildTags()
	testFlags := []string{
		"-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")),
		"-v",
		"-race",
		testCoverageProfileFlag,
		"./..."}
	runGoTest(testFlags...)
}

func integrationTests() {
	prt.Infof("Running integration tests with enabled race detection")
	prepareBuildTags("integration")
	testFlags := []string{"-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")), "-v", "-race", "./..."}
	runGoTest(testFlags...)
}

// runGoBuild runs the Go `build` command using the given build flags.
// By default this adds the `-tags` flag to include build tags passed through the`SNOWSAW_BUILD_TAGS` environment
// variable.
func runGoBuild(envFlags map[string]string, buildFlags ...string) {
	buildFlags = append(buildFlags, "-o", fmt.Sprintf("%s/%s", buildDir, config.ProjectName), config.PackageName)
	// Prepend the Go `build` command.
	buildFlags = append([]string{"build"}, buildFlags...)

	if err := sh.RunWith(envFlags, goExec, buildFlags...); err != nil {
		prt.Errorf("Failed to build package %s: %v", color.GreenString(config.PackageName), err)
		os.Exit(1)
	}

	prt.Successf("Build completed successfully: %s", color.GreenString(fmt.Sprintf("%s/%s", buildDir,
		config.ProjectName)))
}

func runGoImports() {
	prt.Infof("Formatting Go source files")
	formatFlags := []string{
		// A comma-separated list of prefixes for local package imports to be put after 3rd-party packages.
		"-local", fmt.Sprintf("'%s'", config.PackageName),
		// Report all errors and not just the first 10 on different lines.
		"-e",
		// List files whose formatting are not conform to the styleguide.
		"-l",
		// Write result to source files instead of stdout.
		"-w",
		// Search all folders for Go source files recursively starting from the current working directory.
		"."}
	if err := sh.RunV(formatTool.BinaryExecPath, formatFlags...); err != nil {
		prt.Errorf("Failed to format Go source files with import optimizations: %v", err)
		prt.Warnf("Please run manually: %s",
			color.CyanString("%s %s", formatTool.BinaryExecPath, strings.Join(formatFlags, " ")))
		os.Exit(1)
	}

	prt.Successf("All Go source files formatted successfully")
}

func runGolangCILint() {
	golangCIFlags := []string{"run"}
	if err := sh.RunV(lintTool.BinaryExecPath, golangCIFlags...); err != nil {
		prt.Errorf("Linters finished with non-zero exit code")
		os.Exit(1)
	}
	prt.Successf("Linters finished successfully with zero exit code")
}

func runGoTest(testFlags ...string) {
	testFlags = append([]string{"test"}, testFlags...)
	if err := sh.RunV(goExec, testFlags...); err != nil {
		prt.Errorf("Failed to run tests: %v", err)
		prt.Warnf("Please run manually: %s", color.CyanString("%s %s", goExec, strings.Join(testFlags, " ")))
		os.Exit(1)
	}

	prt.Successf("All tests completed successfully")
}

// runGox runs the cross-compile tool command using the given build flags.
// By default this includes the `-tags` flag to include custom build tags passed through the `SNOWSAW_BUILD_TAGS`
// environment variable.
func runGox(envFlags map[string]string, buildFlags ...string) {
	if customTargetPlatforms, hasCustomPlatforms := os.LookupEnv(snowsawEnvCrossPlatformTargets); hasCustomPlatforms {
		prt.Infof("Using custom cross-compile platform targets: %v", strings.Split(customTargetPlatforms, " "))
		crossCompileTargetPlatforms = strings.Split(customTargetPlatforms, " ")
	}
	buildFlags = append(
		buildFlags,
		fmt.Sprintf("-osarch=%s", strings.Join(crossCompileTargetPlatforms, " ")),
		fmt.Sprintf("--output=%s/%s", buildDir, crossCompileNameTemplate))

	// Use sh.Exec helper function to pass the `gox` command output through to stdout/stderr,
	// otherwise the output will be absorbed by the child process.
	if _, err := sh.Exec(envFlags, os.Stdout, os.Stderr, crossCompileTool.BinaryExecPath, buildFlags...); err != nil {
		prt.Errorf("Failed to cross-compile package %s: %v", color.GreenString(config.PackageName), err)
		os.Exit(1)
	}

	prt.Successf("Cross compilation completed successfully with output to %s directory", color.GreenString(buildDir))
}

// validateDevTools validates that all required development tool/dependency executables are bootstrapped and
// available in PATH or "gobin" binary cache.
func validateDevTools() {
	prt.Infof("Verifying development tools/dependencies")
	handleError := func(name string, err error) {
		prt.Errorf("Failed do determine development tool/dependency %s:\n%s",
			color.CyanString(name), color.RedString("  %s", err))
		prt.Warnf("Run the %s task to install all required tools/dependencies!", color.YellowString("bootstrap"))
		os.Exit(1)
	}

	gobinPath, checkGobinPathErr := getExecutablePath(devToolManager.BinaryName, false)
	if checkGobinPathErr != nil {
		handleError(fmt.Sprintf("%s@%s", devToolManager.ModuleName, devToolManager.ModuleVersion), checkGobinPathErr)
	}
	devToolManager.BinaryExecPath = gobinPath

	for _, bd := range []*buildDependency{crossCompileTool, formatTool, lintTool} {
		if bd.ModuleVersion == "" {
			p, e := getExecutablePath(bd.BinaryName, false)
			if e != nil {
				handleError(bd.ModuleName, e)
			}
			bd.BinaryExecPath = p
			continue
		}

		p, e := getExecutablePath(fmt.Sprintf("%s@%s", bd.ModuleName, bd.ModuleVersion), true)
		if e != nil {
			handleError(fmt.Sprintf("%s@%s", bd.ModuleName, bd.ModuleVersion), e)
		}
		bd.BinaryExecPath = p
	}
}
