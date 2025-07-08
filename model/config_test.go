// Package model - configuration tests
package model

import "testing"

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config - default",
			config: Config{
				MoveDuplicates:        false,
				Verbose:               false,
				Silent:                false,
				DetectTransparentPNGs: false,
			},
			wantErr: false,
		},
		{
			name: "valid config - verbose only",
			config: Config{
				MoveDuplicates:        true,
				Verbose:               true,
				Silent:                false,
				DetectTransparentPNGs: false,
			},
			wantErr: false,
		},
		{
			name: "valid config - silent only",
			config: Config{
				MoveDuplicates:        true,
				Verbose:               false,
				Silent:                true,
				DetectTransparentPNGs: false,
			},
			wantErr: false,
		},
		{
			name: "valid config - with transparent PNG detection",
			config: Config{
				MoveDuplicates:        false,
				Verbose:               true,
				Silent:                false,
				DetectTransparentPNGs: true,
			},
			wantErr: false,
		},
		{
			name: "valid config - all features enabled",
			config: Config{
				MoveDuplicates:        true,
				Verbose:               true,
				Silent:                false,
				DetectTransparentPNGs: true,
				LogFilePath:           "/tmp/test.log",
			},
			wantErr: false,
		},
		{
			name: "valid config - silent with PNG detection",
			config: Config{
				MoveDuplicates:        false,
				Verbose:               false,
				Silent:                true,
				DetectTransparentPNGs: true,
			},
			wantErr: false,
		},
		{
			name: "invalid config - both verbose and silent",
			config: Config{
				MoveDuplicates:        false,
				Verbose:               true,
				Silent:                true,
				DetectTransparentPNGs: false,
			},
			wantErr: true,
		},
		{
			name: "invalid config - verbose and silent with all features",
			config: Config{
				MoveDuplicates:        true,
				Verbose:               true,
				Silent:                true,
				DetectTransparentPNGs: true,
				LogFilePath:           "/tmp/test.log",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStats_ThreadSafety(t *testing.T) {
	stats := &Stats{}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				stats.IncrementTotalFiles()
				stats.IncrementFilesMoved()
				stats.IncrementErrors()
				stats.IncrementDuplicatesMoved()
				stats.IncrementTransparentPNGsMoved()
				stats.IncrementUnknownExtensions(".unknown")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if stats.GetTotalFiles() != 1000 {
		t.Errorf("Expected 1000 total files, got %d", stats.GetTotalFiles())
	}
	if stats.GetFilesMoved() != 1000 {
		t.Errorf("Expected 1000 files moved, got %d", stats.GetFilesMoved())
	}
	if stats.GetErrorsCount() != 1000 {
		t.Errorf("Expected 1000 errors, got %d", stats.GetErrorsCount())
	}
	if stats.GetDuplicatesMoved() != 1000 {
		t.Errorf("Expected 1000 duplicates moved, got %d", stats.GetDuplicatesMoved())
	}
	if stats.GetTransparentPNGsMoved() != 1000 {
		t.Errorf("Expected 1000 transparent PNGs moved, got %d", stats.GetTransparentPNGsMoved())
	}
	if stats.GetUnknownExtensions() != 1000 {
		t.Errorf("Expected 1000 unknown extensions, got %d", stats.GetUnknownExtensions())
	}
}

func TestStats_InitialValues(t *testing.T) {
	stats := &Stats{}

	if stats.GetTotalFiles() != 0 {
		t.Errorf("Expected 0 initial total files, got %d", stats.GetTotalFiles())
	}
	if stats.GetFilesMoved() != 0 {
		t.Errorf("Expected 0 initial files moved, got %d", stats.GetFilesMoved())
	}
	if stats.GetErrorsCount() != 0 {
		t.Errorf("Expected 0 initial errors, got %d", stats.GetErrorsCount())
	}
	if stats.GetDuplicatesMoved() != 0 {
		t.Errorf("Expected 0 initial duplicates moved, got %d", stats.GetDuplicatesMoved())
	}
	if stats.GetTransparentPNGsMoved() != 0 {
		t.Errorf("Expected 0 initial transparent PNGs moved, got %d", stats.GetTransparentPNGsMoved())
	}
	if stats.GetUnknownExtensions() != 0 {
		t.Errorf("Expected 0 initial unknown extensions, got %d", stats.GetUnknownExtensions())
	}
}

func TestStats_UnknownExtensions(t *testing.T) {
	stats := &Stats{}

	// initial state
	if stats.GetUnknownExtensions() != 0 {
		t.Errorf("Expected 0 initial unknown extensions, got %d", stats.GetUnknownExtensions())
	}

	// unknown extensions
	stats.IncrementUnknownExtensions(".unknown1")
	stats.IncrementUnknownExtensions(".unknown2")
	stats.IncrementUnknownExtensions(".unknown1") // duplicate

	// total count
	if stats.GetUnknownExtensions() != 3 {
		t.Errorf("Expected 3 unknown extensions total, got %d", stats.GetUnknownExtensions())
	}

	// extension mapping
	unknownMap := stats.GetUnknownExtMap()
	if unknownMap[".unknown1"] != 2 {
		t.Errorf("Expected 2 .unknown1 files, got %d", unknownMap[".unknown1"])
	}
	if unknownMap[".unknown2"] != 1 {
		t.Errorf("Expected 1 .unknown2 file, got %d", unknownMap[".unknown2"])
	}
}
