package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)

// Version represents a semantic version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

// ParseVersion parses a semantic version string
func ParseVersion(v string) (*Version, error) {
	matches := semverRegex.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("invalid semantic version: %s", v)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}, nil
}

// String returns the string representation of the version
func (v *Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		version += "-" + v.Prerelease
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}

// Compare compares two versions (-1: less, 0: equal, 1: greater)
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	
	// Handle prerelease comparison
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}
	if v.Prerelease != other.Prerelease {
		return strings.Compare(v.Prerelease, other.Prerelease)
	}
	
	return 0
}

// VersionRange represents a version constraint
type VersionRange struct {
	Operator string
	Version  *Version
}

// ParseVersionRange parses a version range (^1.0.0, ~1.0.0, >=1.0.0, etc.)
func ParseVersionRange(constraint string) (*VersionRange, error) {
	constraint = strings.TrimSpace(constraint)
	
	if strings.HasPrefix(constraint, "^") {
		version, err := ParseVersion(constraint[1:])
		if err != nil {
			return nil, err
		}
		return &VersionRange{Operator: "^", Version: version}, nil
	}
	
	if strings.HasPrefix(constraint, "~") {
		version, err := ParseVersion(constraint[1:])
		if err != nil {
			return nil, err
		}
		return &VersionRange{Operator: "~", Version: version}, nil
	}
	
	if strings.HasPrefix(constraint, ">=") {
		version, err := ParseVersion(constraint[2:])
		if err != nil {
			return nil, err
		}
		return &VersionRange{Operator: ">=", Version: version}, nil
	}
	
	if strings.HasPrefix(constraint, "<=") {
		version, err := ParseVersion(constraint[2:])
		if err != nil {
			return nil, err
		}
		return &VersionRange{Operator: "<=", Version: version}, nil
	}
	
	if strings.HasPrefix(constraint, ">") {
		version, err := ParseVersion(constraint[1:])
		if err != nil {
			return nil, err
		}
		return &VersionRange{Operator: ">", Version: version}, nil
	}
	
	if strings.HasPrefix(constraint, "<") {
		version, err := ParseVersion(constraint[1:])
		if err != nil {
			return nil, err
		}
		return &VersionRange{Operator: "<", Version: version}, nil
	}
	
	// Exact version
	version, err := ParseVersion(constraint)
	if err != nil {
		return nil, err
	}
	return &VersionRange{Operator: "=", Version: version}, nil
}

// Satisfies checks if a version satisfies the range constraint
func (vr *VersionRange) Satisfies(version *Version) bool {
	switch vr.Operator {
	case "=":
		return version.Compare(vr.Version) == 0
	case "^":
		return version.Major == vr.Version.Major && version.Compare(vr.Version) >= 0
	case "~":
		return version.Major == vr.Version.Major && version.Minor == vr.Version.Minor && version.Compare(vr.Version) >= 0
	case ">=":
		return version.Compare(vr.Version) >= 0
	case "<=":
		return version.Compare(vr.Version) <= 0
	case ">":
		return version.Compare(vr.Version) > 0
	case "<":
		return version.Compare(vr.Version) < 0
	default:
		return false
	}
}