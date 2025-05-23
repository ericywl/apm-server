// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ecclient

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type StackVersionInfo struct {
	Version      StackVersion
	UpgradableTo []StackVersion
}

// CanUpgradeTo checks if the current stack version can upgrade to the provided `version`.
func (info StackVersionInfo) CanUpgradeTo(version StackVersion) bool {
	for _, upgrade := range info.UpgradableTo {
		if upgrade == version {
			return true
		}
	}
	return false
}

type StackVersionInfos []StackVersionInfo

// Sort sorts the stack versions in ascending order based on
// major, minor, patch, suffix in order of importance.
func (infos StackVersionInfos) Sort() {
	cmpFn := func(a, b StackVersionInfo) int {
		return a.Version.Compare(b.Version)
	}

	if slices.IsSortedFunc(infos, cmpFn) {
		return
	}

	slices.SortFunc(infos, cmpFn)
}

// Last returns the last version in the list.
func (infos StackVersionInfos) Last() (StackVersionInfo, bool) {
	if len(infos) == 0 {
		return StackVersionInfo{}, false
	}
	return infos[len(infos)-1], true
}

// LatestFor retrieves the latest stack version for that prefix.
// The prefix must loosely follow semantic versioning in the form of:
//   - X.Y.Z
//   - X.Y
//   - X
//
// Invalid prefix will cause this function to panic.
//
// Note: This assumes that StackVersionInfos is already sorted in ascending order.
func (infos StackVersionInfos) LatestFor(prefix string) (StackVersionInfo, bool) {
	lv, err := parseVersionPrefix(prefix)
	if err != nil {
		panic(err)
	}

	for i := len(infos) - 1; i >= 0; i-- {
		if ok := infos[i].Version.looseMatch(lv); ok {
			return infos[i], true
		}
	}
	return StackVersionInfo{}, false
}

// LatestForMajor retrieves the latest stack version for that major.
//
// Note: This assumes that StackVersionInfos is already sorted in ascending order.
func (infos StackVersionInfos) LatestForMajor(major uint64) (StackVersionInfo, bool) {
	for i := len(infos) - 1; i >= 0; i-- {
		if infos[i].Version.IsMajor(major) {
			return infos[i], true
		}
	}
	return StackVersionInfo{}, false
}

// LatestForMinor retrieves the latest stack version for that minor.
//
// Note: This assumes that StackVersionInfos is already sorted in ascending order.
func (infos StackVersionInfos) LatestForMinor(major, minor uint64) (StackVersionInfo, bool) {
	for i := len(infos) - 1; i >= 0; i-- {
		if infos[i].Version.IsMinor(major, minor) {
			return infos[i], true
		}
	}
	return StackVersionInfo{}, false
}

// PreviousMinorLatest retrieves the latest stack version from the previous
// minor of the provided `version`.
// If the minor of `version` is 0, the latest version for previous major is
// returned instead.
//
// Note: This assumes that StackVersionInfos is already sorted in ascending order.
func (infos StackVersionInfos) PreviousMinorLatest(version StackVersion) (StackVersionInfo, bool) {
	if version.Minor == 0 {
		// When the minor is 0, we want the latest of the previous major
		return infos.LatestForMajor(version.Major - 1)
	}
	return infos.LatestForMinor(version.Major, version.Minor-1)
}

// PreviousPatch retrieves the previous patch version info from the provided `version`.
//
// Note: This assumes that StackVersionInfos is already sorted in ascending order.
func (infos StackVersionInfos) PreviousPatch(version StackVersion) (StackVersionInfo, bool) {
	if version.Patch == 0 {
		// When the patch is 0, we want the latest of the previous minor
		return infos.LatestForMinor(version.Major, version.Minor-1)
	}
	prevPatch := version
	prevPatch.Patch = version.Patch - 1
	return infos.GetByVersion(prevPatch)
}

// GetByVersion returns the version info if the provided `version` exists in the list.
func (infos StackVersionInfos) GetByVersion(version StackVersion) (StackVersionInfo, bool) {
	for i := len(infos) - 1; i >= 0; i-- {
		if infos[i].Version == version {
			return infos[i], true
		}
	}
	return StackVersionInfo{}, false
}

type StackVersion struct {
	Major  uint64
	Minor  uint64
	Patch  uint64
	Suffix string // Optional
}

func NewStackVersion(major, minor, patch uint64, suffix string) StackVersion {
	return StackVersion{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		Suffix: suffix,
	}
}

func NewStackVersionFromStr(versionStr string) (StackVersion, error) {
	splits := strings.SplitN(versionStr, ".", 3)
	if len(splits) != 3 {
		return StackVersion{}, errors.New("invalid format")
	}

	major, err := strconv.ParseUint(splits[0], 10, 64)
	if err != nil {
		return StackVersion{}, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.ParseUint(splits[1], 10, 64)
	if err != nil {
		return StackVersion{}, fmt.Errorf("invalid minor version: %w", err)
	}

	splits = strings.SplitN(splits[2], "-", 2)
	patch, err := strconv.ParseUint(splits[0], 10, 64)
	if err != nil {
		return StackVersion{}, fmt.Errorf("invalid patch version: %w", err)
	}

	suffix := ""
	if len(splits) > 1 {
		suffix = splits[1]
	}
	return NewStackVersion(major, minor, patch, suffix), nil
}

func (v StackVersion) String() string {
	var suffix string
	if v.Suffix != "" {
		suffix = "-" + v.Suffix
	}
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, suffix)
}

func (v StackVersion) MajorMinor() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func (v StackVersion) IsMajor(major uint64) bool {
	return v.Major == major
}

func (v StackVersion) IsMinor(major, minor uint64) bool {
	return v.Major == major && v.Minor == minor
}

func (v StackVersion) IsPatch(major, minor, patch uint64) bool {
	return v.Major == major && v.Minor == minor && v.Patch == patch
}

func (v StackVersion) Compare(other StackVersion) int {
	res := cmp.Compare(v.Major, other.Major)
	if res != 0 {
		return res
	}
	res = cmp.Compare(v.Minor, other.Minor)
	if res != 0 {
		return res
	}
	res = cmp.Compare(v.Patch, other.Patch)
	if res != 0 {
		return res
	}
	return cmp.Compare(v.Suffix, other.Suffix)
}

// HasPrefix checks if the stack version contains the prefix.
// The prefix must loosely follow semantic versioning in the form of:
//   - X.Y.Z
//   - X.Y
//   - X
func (v StackVersion) HasPrefix(prefix string) (bool, error) {
	lv, err := parseVersionPrefix(prefix)
	if err != nil {
		return false, err
	}
	return v.looseMatch(lv), nil
}

type looseVersion struct {
	major, minor, patch *uint64
}

func (v StackVersion) looseMatch(lv looseVersion) bool {
	// Only major
	if lv.minor == nil {
		return v.IsMajor(*lv.major)
	}
	// Only major minor
	if lv.patch == nil {
		return v.IsMinor(*lv.major, *lv.minor)
	}
	// Major, minor, patch
	return v.IsPatch(*lv.major, *lv.minor, *lv.patch)
}

var looseVersionRegex = regexp.MustCompile(`^(\d*)(?:\.(\d*))?(?:\.(\d*))?(?:-(\w*))?$`)

func parseVersionPrefix(prefix string) (looseVersion, error) {
	lv := looseVersion{}
	// First match is the whole string, last match is the suffix, total 5
	matches := looseVersionRegex.FindStringSubmatch(prefix)
	if len(matches) == 0 || len(matches) > 5 {
		return looseVersion{}, errors.New("invalid prefix format")
	}

	major, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return looseVersion{}, fmt.Errorf("invalid major version: %w", err)
	}

	// Only major
	lv.major = &major
	if matches[2] == "" {
		return lv, nil
	}

	minor, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return looseVersion{}, fmt.Errorf("invalid minor version: %w", err)
	}

	// Only major minor
	lv.minor = &minor
	if matches[3] == "" {
		return lv, nil
	}

	// Major, minor, patch
	patch, err := strconv.ParseUint(matches[3], 10, 64)
	if err != nil {
		return looseVersion{}, fmt.Errorf("invalid patch version: %w", err)
	}

	lv.patch = &patch
	return lv, nil
}
