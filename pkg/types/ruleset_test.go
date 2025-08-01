package types

import (
	"testing"
)

func TestParseRulesetName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOrg  string
		wantPkg  string
	}{
		{"unscoped package", "typescript-rules", "", "typescript-rules"},
		{"scoped package", "company@security-rules", "company", "security-rules"},
		{"empty string", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrg, gotPkg := ParseRulesetName(tt.input)
			if gotOrg != tt.wantOrg || gotPkg != tt.wantPkg {
				t.Errorf("ParseRulesetName(%q) = (%q, %q), want (%q, %q)", 
					tt.input, gotOrg, gotPkg, tt.wantOrg, tt.wantPkg)
			}
		})
	}
}

func TestFormatRulesetName(t *testing.T) {
	tests := []struct {
		name    string
		org     string
		pkg     string
		want    string
	}{
		{"unscoped package", "", "typescript-rules", "typescript-rules"},
		{"scoped package", "company", "security-rules", "company@security-rules"},
		{"empty org", "", "test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRulesetName(tt.org, tt.pkg)
			if got != tt.want {
				t.Errorf("FormatRulesetName(%q, %q) = %q, want %q", 
					tt.org, tt.pkg, got, tt.want)
			}
		})
	}
}

func TestRulesetValidate(t *testing.T) {
	tests := []struct {
		name    string
		ruleset Ruleset
		wantErr bool
	}{
		{
			"valid ruleset",
			Ruleset{
				Name:     "test-rules",
				Version:  "1.0.0",
				Source:   "https://registry.example.com",
				Files:    []string{"rule1.md", "rule2.md"},
				Checksum: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			},
			false,
		},
		{
			"missing name",
			Ruleset{
				Version:  "1.0.0",
				Source:   "https://registry.example.com",
				Files:    []string{"rule1.md"},
				Checksum: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			},
			true,
		},
		{
			"missing files",
			Ruleset{
				Name:     "test-rules",
				Version:  "1.0.0",
				Source:   "https://registry.example.com",
				Files:    []string{},
				Checksum: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			},
			true,
		},
		{
			"invalid checksum",
			Ruleset{
				Name:     "test-rules",
				Version:  "1.0.0",
				Source:   "https://registry.example.com",
				Files:    []string{"rule1.md"},
				Checksum: "invalid",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ruleset.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Ruleset.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	r := &Ruleset{}
	content := []byte("test content")
	checksum := r.CalculateChecksum(content)
	
	if len(checksum) != 64 {
		t.Errorf("CalculateChecksum() returned checksum of length %d, want 64", len(checksum))
	}
	
	// Same content should produce same checksum
	checksum2 := r.CalculateChecksum(content)
	if checksum != checksum2 {
		t.Errorf("CalculateChecksum() not deterministic")
	}
}