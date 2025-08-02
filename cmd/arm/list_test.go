package main

import (
	"testing"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			"https URL",
			"https://registry.example.com/path",
			"registry.example.com",
		},
		{
			"http URL",
			"http://registry.example.com/path",
			"registry.example.com",
		},
		{
			"no protocol",
			"registry.example.com/path",
			"registry.example.com",
		},
		{
			"no path",
			"https://registry.example.com",
			"registry.example.com",
		},
		{
			"empty string",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDomain(tt.url)
			if got != tt.want {
				t.Errorf("extractDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatSource(t *testing.T) {
	tests := []struct {
		name      string
		sourceURL string
		want      string
	}{
		{
			"unknown source",
			"https://unknown-registry.com",
			"unknown-registry.com",
		},
		{
			"empty source",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSource(tt.sourceURL)
			if got != tt.want {
				t.Errorf("formatSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
