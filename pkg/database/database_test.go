package database

import "testing"

func TestVersionAtLeast(t *testing.T) {
	tests := []struct {
		version string
		minimum string
		want    bool
		wantErr bool
	}{
		{"1.8.11", "1.8.4", true, false},
		{"1.8.10", "1.8.4", true, false},
		{"1.8.4", "1.8.4", true, false},
		{"1.8.3", "1.8.4", false, false},
		{"1.10.0", "1.8.4", true, false},
		{"2.0.0", "1.8.4", true, false},
		{"1.7.99", "1.8.4", false, false},
		{"1.8", "1.8.4", false, false},
		{"1.8", "1.8.0", true, false},
		{"1", "1.0.0", true, false},
		{"2", "1.8.4", true, false},
		{"1.0.0", "1", true, false},
		{"1.8.0.1", "1.8.4", false, true},
		{"", "1.8.4", false, true},
		{"1.8.", "1.8.4", false, true},
		{"1.8.11-rc1", "1.8.4", false, true},
		{"1..4", "1.8.4", false, true},
		{"1.8.18446744073709551616", "1.8.4", false, true},
		{"1.8.11", "invalid", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.version+"_vs_"+tt.minimum, func(t *testing.T) {
			got, err := versionAtLeast(tt.version, tt.minimum)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("versionAtLeast(%q, %q) = %v, %v; want %v, error=%v", tt.version, tt.minimum, got, err, tt.want, tt.wantErr)
			}
		})
	}
}
