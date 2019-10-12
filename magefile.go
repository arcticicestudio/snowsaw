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
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/arcticicestudio/snowsaw/pkg/config"
	"github.com/arcticicestudio/snowsaw/pkg/prt"
)

// buildDependency represents a build dependency like a tool used to build or develop the project.
type buildDependency struct {
	// The path of the binary executable.
	BinaryExecPath string
	// The name of the binary.
	BinaryName string
	// The name of the package.
	PackageName string
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
		BinaryName:  "gox",
		PackageName: "github.com/mitchellh/gox@v1.0.1",
	}

	// The tool used to format all Go source files.
	// See https://godoc.org/golang.org/x/tools/cmd/goimports for more details.
	formatTool = &buildDependency{
		PackageName: "golang.org/x/tools/cmd/goimports",
		BinaryName:  "goimports",
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
		PackageName: "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1",
		BinaryName:  "golangci-lint",
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

// Build compiles the project in development mode for the current OS and architecture type.
func Build() {
	mg.SerialDeps(Clean, compile)
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
	mg.SerialDeps(Clean, validateBuildDependencies, compileProd)
}

// DistCrossPlatform builds the project in production mode for cross-platform distribution.
// This includes all steps from the current platform distribution/production task `Dist`,
// but instead builds for all configured OS/architecture types.
func DistCrossPlatform() {
	mg.SerialDeps(Clean, validateBuildDependencies, compileProdCross)
}

// DistCrossPlatformOpt builds the project in production mode for cross-platform distribution with optimizations.
// This includes all steps from the cross-platform distribution task `DistCrossPlatform` and additionally removes all
// debug metadata to shrink the memory overhead and file size as well as reducing the chance for possible security
// related problems due to enabled development features and leaked debug information.
func DistCrossPlatformOpt() {
	mg.SerialDeps(Clean, validateBuildDependencies, compileProdCrossOpt)
}

// DistOpt builds the project in production mode with optimizations like minification and debug symbol stripping.
// This includes all steps from the production build task `Dist` and additionally removes all debug metadata to shrink
// the memory overhead and file size as well as reducing the chance for possible security related problems due to
// enabled development features and leaked debug information.
func DistOpt() {
	mg.SerialDeps(Clean, validateBuildDependencies, compileProdOpt)
}

// Format searches all project Go source files and formats them according to the Go code styleguide.
func Format() {
	mg.SerialDeps(validateBuildDependencies, runGoImports)
}

// Lint runs all linters configured and executed through `golangci-lint`.
// See the `.golangci.yml` configuration file and official GolangCI documentations at https://golangci.com
// and https://github.com/golangci/golangci-lint for more details.
func Lint() {
	mg.SerialDeps(validateBuildDependencies, runGolangCILint)
}

// Test runs all unit tests with enabled race detection.
func Test() {
	mg.SerialDeps(unitTests)
}

// TestCover runs all unit tests with with coverage reports and enabled race detection.
func TestCover() {
	mg.SerialDeps(Clean)
	// Ensure the required directory structure exists, `go test` doesn't create it automatically.
	createDirectoryStructure(reportsDir)
	testCoverageProfileFlag = fmt.Sprintf("-coverprofile=%s", filepath.Join(reportsDir, testCoverageOutputFileName))
	mg.SerialDeps(unitTests)
}

// TestIntegration runs all integration tests with enabled race detection.
func TestIntegration() {
	mg.SerialDeps(integrationTests)
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

// getEnvFlags returns environment variables storing metadata the build time, Git version tags and commit checksum.
func getEnvFlags() map[string]string {
	buildDate := time.Now().Format(time.RFC3339)

	pwd, err := os.Getwd()
	if err != nil {
		prt.Errorf("Could not determine project root path: %v", err)
		os.Exit(1)
	}

	var version []string
	// Find the latest Git tag in all branches, otherwise use the defined version.
	latestGitTag, _ := sh.Output("git", "rev-list", "--tags", "--max-count=1")
	if tag, gitTagErr := sh.Output("git", "describe", "--tags", latestGitTag); gitTagErr != nil {
		version = append(version, "develop")
		prt.Infof("No Git tag found, using %s as fallback version", color.CyanString("develop"))
	} else {
		version = append(version, tag)
		prt.Infof("Using Git tag %s as version", color.CyanString(tag))
	}

	// If the current branch is ahead of the `master` branch append the commit count and hash of the latest commit.
	commitsAhead, _ := sh.Output("git", "rev-list", "--count", "master..HEAD")
	commitCount, err := strconv.Atoi(commitsAhead)
	if err == nil && commitCount > 0 {
		commitHash, _ := sh.Output("git", "rev-parse", "--short=8", "HEAD")
		version = append(version, commitsAhead, commitHash)
		prt.Infof("Building with latest Git commit %s, %s commits ahead of %s branch",
			color.CyanString(commitHash), color.CyanString(commitsAhead), color.CyanString("master"))
	}

	prt.Infof(
		"Injecting %s:\n"+
			"  Build Date: %s\n"+
			"  Version: %s",
		color.CyanString("LDFLAGS"), color.CyanString(buildDate), color.CyanString(strings.Join(version, "-")))

	prt.Infof(
		"Injecting %s:\n"+
			"  -trimpath: %s",
		color.CyanString("ASMFLAGS"), color.CyanString(pwd))

	prt.Infof(
		"Injecting %s:\n"+
			"  -trimpath: %s",
		color.CyanString("GCFLAGS"), color.CyanString(pwd))

	return map[string]string{
		"BUILD_DATE_TIME": buildDate,
		"PACKAGE_NAME":    config.PackageName,
		"PROJECT_ROOT":    pwd,
		"VERSION":         strings.Join(version, "-")}
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

// validateBuildDependencies checks if all required build dependencies are installed, the binaries are available in
// PATH and will try to install them if not passing the checks.
func validateBuildDependencies() {
	for _, bd := range []*buildDependency{crossCompileTool, formatTool, lintTool} {
		binPath, err := exec.LookPath(bd.BinaryName)
		if err == nil {
			bd.BinaryExecPath = binPath
			prt.Infof("Required build dependency %s already installed: %s",
				color.CyanString(bd.PackageName),
				color.BlueString(bd.BinaryExecPath))
			continue
		}

		prt.Infof("Installing required build dependency: %s", color.CyanString(bd.PackageName))
		c := exec.Command(goExec, "get", "-u", bd.PackageName)
		// Run installations outside of the project root directory to prevent the pollution of the project's Go module
		// file.
		// This is a necessary workaround until the Go toolchain is able to install packages globally without
		// updating the module file when the "go get" command is run from within the project root directory.
		// See https://github.com/golang/go/issues/30515 for more details or more details and proposed solutions
		// that might be added to Go's build tools in future versions.
		c.Dir = os.TempDir()
		c.Env = os.Environ()
		// Explicitly enable "module" mode to install development dependencies to allow to use pinned module versions.
		env := map[string]string{"GO111MODULE": "on"}
		for k, v := range env {
			c.Env = append(c.Env, k+"="+v)
		}
		if err = c.Run(); err != nil {
			prt.Errorf("Failed to install required build dependency %s: %v", color.CyanString(bd.PackageName), err)
			prt.Warnf("Please install manually: %s", color.CyanString("go get -u %s", bd.PackageName))
			os.Exit(1)
		}

		binPath, err = exec.LookPath(bd.BinaryName)
		if err != nil {
			bd.BinaryExecPath = binPath
			prt.Errorf("Failed to find executable path of required build dependency %s after installation: %v",
				color.CyanString(bd.PackageName), err)
			os.Exit(1)
		}
		bd.BinaryExecPath = binPath
		prt.Infof("Using executable %s of installed build dependency %s",
			color.CyanString(bd.BinaryExecPath),
			color.BlueString(bd.PackageName))
	}
}
