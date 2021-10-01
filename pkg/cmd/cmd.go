// Copyright 2021 Sylabs Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/sylabs/release-tools/pkg/git"
)

type Command struct {
	env  map[string]string
	args []string
}

// Env returns the command environment.
func (c *Command) Env() map[string]string { return c.env }

// Args returns the command arguments.
func (c *Command) Args() []string { return c.args }

// buildOpts accumulates build/install options.
type buildOpts struct {
	// If true, disable CGO.
	disableCGO bool

	// Remove all file system paths from the resulting executable. Sets "-trimpath" build flag.
	trimPath bool

	// Omit the symbol table and debug information. Sets "-s" linker flag.
	omitSymbolsAndDebug bool

	// Omit the DWARF symbol table. Sets "-w" linker flag.
	omitDWARF bool

	// Entity that created the build. Sets main.builtBy value.
	builtBy string

	// Hash of current commit. Sets main.commit value.
	commit string

	// Date of current commit. Sets main.date value (in time.RFC3339 format).
	date time.Time

	// State of current working directory (ex clean, dirty). Sets main.state value.
	state string

	// Version associated with the current commit. Sets main.version value.
	version semver.Version

	// List of packages to build/install.
	packages []string
}

// linkFlags return build flags that correspond to bo.
func (bo buildOpts) linkFlags() []string {
	var flags []string

	if bo.omitSymbolsAndDebug {
		flags = append(flags, "-s")
	}

	if bo.omitDWARF {
		flags = append(flags, "-w")
	}

	if bo.builtBy != "" {
		flags = append(flags, "-X", fmt.Sprintf("main.builtBy=%v", bo.builtBy))
	}

	if bo.commit != "" {
		flags = append(flags, "-X", fmt.Sprintf("main.commit=%v", bo.commit))
	}

	if !bo.date.IsZero() {
		flags = append(flags, "-X", fmt.Sprintf("main.date=%v", bo.date.UTC().Format(time.RFC3339)))
	}

	if bo.state != "" {
		flags = append(flags, "-X", fmt.Sprintf("main.state=%v", bo.state))
	}

	if !bo.version.Equals(semver.Version{}) {
		flags = append(flags, "-X", fmt.Sprintf("main.version=%v", bo.version))
	}

	return flags
}

// env return a build environment that correspond to bo.
func (bo buildOpts) env() map[string]string {
	env := make(map[string]string)

	if bo.disableCGO {
		env["CGO_ENABLED"] = "0"
	}

	return env
}

// buildFlags return build flags that correspond to bo.
func (bo buildOpts) buildFlags() []string {
	var flags []string

	if bo.trimPath {
		flags = append(flags, "-trimpath")
	}

	ldFlags := bo.linkFlags()
	if len(ldFlags) > 0 {
		flags = append(flags, "-ldflags", strings.Join(ldFlags, " "))
	}

	return flags
}

// BuildOption are used to specify build options.
type BuildOption func(*buildOpts) error

// OptBuildPackages adds package(s) to the build.
func OptBuildPackages(packages ...string) BuildOption {
	return func(bo *buildOpts) error {
		bo.packages = append(bo.packages, packages...)
		return nil
	}
}

// OptBuildWithBuiltBy sets the building entity (ex. mage, goreleaser).
//
// When supplied, the linker sets main.builtBy to the supplied value.
func OptBuildWithBuiltBy(s string) BuildOption {
	return func(bo *buildOpts) error {
		bo.builtBy = s
		return nil
	}
}

// OptBuildWithGitDescription sets values in the build according to the supplied git description.
//
// When supplied, main.commit and main.state values are set to reflect d. If the working directory
// described by d is clean, main.date is set to reflect the commit time. Otherwise, main.date is
// set to the current time. If d contains a version, the linker sets main.version to that value.
func OptBuildWithGitDescription(d *git.Description) BuildOption {
	return func(bo *buildOpts) error {
		bo.commit = d.CommitHash()

		if d.IsClean() {
			bo.date = d.CommitTime()
			bo.state = "clean"
		} else {
			bo.date = time.Now()
			bo.state = "dirty"
		}

		if v, err := d.Version(); err == nil {
			bo.version = v
		}

		return nil
	}
}

// NewBuildCommand returns a Command that builds packages, configured by opts. Flags are set to
// omit the symbol table, debug information, and DWARF symbol table. A flag is set to ensure file
// system paths are removed, which assists in reproducible builds.
//
// By default, all packages are built. To change this, use OptBuildPackages.
//
// To set a building entity, consider using OptBuildWithBuiltBy. To set variables according to a
// git description, consider using OptBuildWithGitDescription.
func NewBuildCommand(opts ...BuildOption) (*Command, error) {
	bo := buildOpts{
		disableCGO:          true,
		omitSymbolsAndDebug: true,
		omitDWARF:           true,
		trimPath:            true,
	}

	for _, opt := range opts {
		if err := opt(&bo); err != nil {
			return nil, err
		}
	}

	c := Command{
		env:  bo.env(),
		args: []string{"build"},
	}

	c.args = append(c.args, bo.buildFlags()...)

	if pkgs := bo.packages; len(pkgs) == 0 {
		c.args = append(c.args, "./...")
	} else {
		c.args = append(c.args, pkgs...)
	}

	return &c, nil
}

// NewInstallCommand returns a Command that installs packages, configured by opts. Flags are set to
// omit the symbol table, debug information, and DWARF symbol table. A flag is set to ensure file
// system paths are removed, which assists in reproducible builds.
//
// By default, packages in ./cmd/... are installed. To change this, use OptBuildPackages.
//
// To set a building entity, consider using OptBuildSetBuiltBy. To set variables according to a git
// description, consider using OptBuildWithGitDescription.
func NewInstallCommand(opts ...BuildOption) (*Command, error) {
	bo := buildOpts{
		disableCGO:          true,
		omitSymbolsAndDebug: true,
		omitDWARF:           true,
		trimPath:            true,
	}

	for _, opt := range opts {
		if err := opt(&bo); err != nil {
			return nil, err
		}
	}

	c := Command{
		env:  bo.env(),
		args: []string{"install"},
	}

	c.args = append(c.args, bo.buildFlags()...)

	if pkgs := bo.packages; len(pkgs) == 0 {
		c.args = append(c.args, "./cmd/...")
	} else {
		c.args = append(c.args, pkgs...)
	}

	return &c, nil
}

// testOpts accumulates test options.
type testOpts struct {
	// If true, disable CGO.
	disableCGO bool

	// If true, enable data race detection.
	race bool

	// If not empty, write test coverage to the supplied path.
	coverPath string

	// List of packages to test.
	packages []string
}

// env return a test environment that correspond to to.
func (to testOpts) env() map[string]string {
	env := make(map[string]string)

	if to.disableCGO {
		env["CGO_ENABLED"] = "0"
	}

	return env
}

// testFlags returns test flags that correspond to bo.
func (to testOpts) testFlags() []string {
	var flags []string

	if to.race {
		flags = append(flags, "-race")
	}

	if path := to.coverPath; path == "" {
		flags = append(flags, "-cover")
	} else {
		flags = append(flags, "-coverprofile", path)
	}

	return flags
}

// TestOption are used to specify testing options.
type TestOption func(*testOpts) error

// OptTestPackages adds package(s) to test.
func OptTestPackages(packages ...string) TestOption {
	return func(to *testOpts) error {
		to.packages = append(to.packages, packages...)
		return nil
	}
}

// OptTestWithCoverPath sets path as the location to write a test coverage profile.
func OptTestWithCoverPath(path string) TestOption {
	return func(to *testOpts) error {
		to.coverPath = path
		return nil
	}
}

// NewTestCommand returns a Command that tests all packages, configured by opts. Data race
// detection is always enabled.
//
// By default, all packages are tested. To change this, use OptTestPackages.
//
// By default, test coverage is written to standard output only. To write a test coverage profile,
// consider using OptTestWithCoverPath.
func NewTestCommand(opts ...TestOption) (*Command, error) {
	to := testOpts{
		disableCGO: false, // CGO required for data race detection.
		race:       true,
	}

	for _, opt := range opts {
		if err := opt(&to); err != nil {
			return nil, err
		}
	}

	c := Command{
		env:  to.env(),
		args: []string{"test"},
	}

	c.args = append(c.args, to.testFlags()...)

	if pkgs := to.packages; len(pkgs) == 0 {
		c.args = append(c.args, "./...")
	} else {
		c.args = append(c.args, pkgs...)
	}

	return &c, nil
}
