package config

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		envVariables  map[string]string
		expected *Config
	}{
		{
			name:    "no environment variables set",
			envVariables: map[string]string{},
			expected: &Config{
				MediaPath:   "/mnt/RAID/qbit-data/downloads",
				ManagerPath: "/mnt/RAID/media-manager",
				LibraryPath: "/mnt/RAID/jelly/media",
				DryRun:      true,
			},
		},
		{
			name: "all environment variables set",
			envVariables: map[string]string{
				"TORRENT_DOWNLOAD_PATH":  "/custom/downloads",
				"TORRENT_MANAGER_PATH":   "/custom/manager",
				"MEDIA_SERVER_PATH":      "/custom/media",
				"TORRENT_MANAGER_DRY_RUN": "false",
			},
			expected: &Config{
				MediaPath:   "/custom/downloads",
				ManagerPath: "/custom/manager",
				LibraryPath: "/custom/media",
				DryRun:      false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clear environment before each test
			clearEnv()

			// Set test environment variables
			for key, value := range test.envVariables {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer clearEnv()

			cfg := New()

			if cfg.MediaPath != test.expected.MediaPath {
				t.Errorf("MediaPath = %v, want %v", cfg.MediaPath, test.expected.MediaPath)
			}
			if cfg.ManagerPath != test.expected.ManagerPath {
				t.Errorf("ManagerPath = %v, want %v", cfg.ManagerPath, test.expected.ManagerPath)
			}
			if cfg.LibraryPath != test.expected.LibraryPath {
				t.Errorf("LibraryPath = %v, want %v", cfg.LibraryPath, test.expected.LibraryPath)
			}
			if cfg.DryRun != test.expected.DryRun {
				t.Errorf("DryRun = %v, want %v", cfg.DryRun, test.expected.DryRun)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name			string
		key 			string
		defaultValue 	string
		envValue   		string
		expectedValue	string
		setEnv     		bool
	}{
		{
			name:			"env variable set to custom",
			key:			"TEST_KEY",
			defaultValue:	"default",
			envValue:		"custom",
			expectedValue:	"custom",
			setEnv:			true,
		},
		{
			name:			"env variable not set and default value is default",
			key:			"TEST_KEY",
			defaultValue:	"default",
			envValue:		"not set",
			expectedValue:	"default",
			setEnv:			false,
		},
		{
			name:			"env variable set to empty string",
			key:			"TEST_KEY",
			defaultValue:	"default",
			envValue:		"",
			expectedValue:	"default",
			setEnv:			true,
		},
		{
			name:			"env variable not set and default value is empty string",
			key:			"TEST_KEY",
			defaultValue:	"",
			envValue:		"not set",
			expectedValue:	"",
			setEnv:			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Unsetenv(test.key)

			if test.setEnv {
				os.Setenv(test.key, test.envValue)
			}

			result := getEnv(test.key, test.defaultValue)

			if result != test.expectedValue {
				t.Errorf("getEnv() = %v, want %v", result, test.expectedValue)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name			string
		key				string
		defaultValue	bool
		envValue   		string
		expectedValue   bool
		setEnv			bool
	}{
		{
			name:			"env variable set to true",
			key: 			"TEST_BOOL",
			defaultValue:	false,
			envValue:		"true",
			expectedValue:	true,
			setEnv:			true,
		},
		{
			name:			"env variable set to false",
			key: 			"TEST_BOOL",
			defaultValue:	true,
			envValue:		"false",
			expectedValue:	false,
			setEnv:			true,
		},
		{
			name:			"env variable not set and default value true",
			key: 			"TEST_BOOL",
			defaultValue:	true,
			envValue:		"false",
			expectedValue:	true,
			setEnv:			false,
		},
		{
			name:			"env variable not set and default value false",
			key: 			"TEST_BOOL",
			defaultValue:	false,
			envValue:		"true",
			expectedValue:	false,
			setEnv:			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Unsetenv(test.key)

			if test.setEnv {
				os.Setenv(test.key, test.envValue)
			}

			result := getEnvBool(test.key, test.defaultValue)
			if result != test.expectedValue {
				t.Errorf("getEnvBool() = %v, want %v", result, test.expectedValue)
			}
		})
	}
}

func clearEnv() {
	os.Unsetenv("TORRENT_DOWNLOAD_PATH")
	os.Unsetenv("TORRENT_MANAGER_PATH")
	os.Unsetenv("MEDIA_SERVER_PATH")
	os.Unsetenv("TORRENT_MANAGER_DRY_RUN")
}
