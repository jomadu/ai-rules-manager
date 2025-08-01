package parser

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			"basic version",
			"1.2.3",
			&Version{Major: 1, Minor: 2, Patch: 3},
			false,
		},
		{
			"version with prerelease",
			"1.2.3-alpha.1",
			&Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1"},
			false,
		},
		{
			"version with build",
			"1.2.3+build.1",
			&Version{Major: 1, Minor: 2, Patch: 3, Build: "build.1"},
			false,
		},
		{
			"version with prerelease and build",
			"1.2.3-alpha.1+build.1",
			&Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "alpha.1", Build: "build.1"},
			false,
		},
		{
			"invalid version",
			"invalid",
			nil,
			true,
		},
		{
			"missing patch",
			"1.2",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.Major != tt.want.Major || got.Minor != tt.want.Minor ||
				got.Patch != tt.want.Patch || got.Prerelease != tt.want.Prerelease ||
				got.Build != tt.want.Build) {
				t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"equal versions", "1.2.3", "1.2.3", 0},
		{"v1 greater major", "2.0.0", "1.9.9", 1},
		{"v1 less major", "1.0.0", "2.0.0", -1},
		{"v1 greater minor", "1.2.0", "1.1.9", 1},
		{"v1 less minor", "1.1.0", "1.2.0", -1},
		{"v1 greater patch", "1.2.3", "1.2.2", 1},
		{"v1 less patch", "1.2.2", "1.2.3", -1},
		{"prerelease vs release", "1.2.3-alpha", "1.2.3", -1},
		{"release vs prerelease", "1.2.3", "1.2.3-alpha", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			if err != nil {
				t.Fatalf("Failed to parse v1: %v", err)
			}
			v2, err := ParseVersion(tt.v2)
			if err != nil {
				t.Fatalf("Failed to parse v2: %v", err)
			}

			got := v1.Compare(v2)
			if got != tt.want {
				t.Errorf("Version.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseVersionRange(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		wantOp     string
		wantVer    string
		wantErr    bool
	}{
		{"caret range", "^1.2.3", "^", "1.2.3", false},
		{"tilde range", "~1.2.3", "~", "1.2.3", false},
		{"greater equal", ">=1.2.3", ">=", "1.2.3", false},
		{"less equal", "<=1.2.3", "<=", "1.2.3", false},
		{"greater than", ">1.2.3", ">", "1.2.3", false},
		{"less than", "<1.2.3", "<", "1.2.3", false},
		{"exact version", "1.2.3", "=", "1.2.3", false},
		{"invalid version", "^invalid", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersionRange(tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersionRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Operator != tt.wantOp {
					t.Errorf("ParseVersionRange() operator = %v, want %v", got.Operator, tt.wantOp)
				}
				if got.Version.String() != tt.wantVer {
					t.Errorf("ParseVersionRange() version = %v, want %v", got.Version.String(), tt.wantVer)
				}
			}
		})
	}
}

func TestVersionRangeSatisfies(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		want       bool
	}{
		{"caret allows minor updates", "^1.2.3", "1.3.0", true},
		{"caret blocks major updates", "^1.2.3", "2.0.0", false},
		{"caret allows patch updates", "^1.2.3", "1.2.4", true},
		{"tilde allows patch updates", "~1.2.3", "1.2.4", true},
		{"tilde blocks minor updates", "~1.2.3", "1.3.0", false},
		{"exact match", "1.2.3", "1.2.3", true},
		{"exact no match", "1.2.3", "1.2.4", false},
		{"greater equal true", ">=1.2.3", "1.2.3", true},
		{"greater equal false", ">=1.2.3", "1.2.2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr, err := ParseVersionRange(tt.constraint)
			if err != nil {
				t.Fatalf("Failed to parse constraint: %v", err)
			}

			version, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to parse version: %v", err)
			}

			got := vr.Satisfies(version)
			if got != tt.want {
				t.Errorf("VersionRange.Satisfies() = %v, want %v", got, tt.want)
			}
		})
	}
}
