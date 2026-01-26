package config

import (
	"flag"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		envVariables map[string]string
		args         []string
		expected     *Config
	}{
		{
			name:         "no environment variables or flags set",
			envVariables: map[string]string{},
			args:         []string{},
			expected: &Config{
				TorrentPath: "/opt/qbit/downloads",
				MoviePath:   "/opt/jellyfin/media/movies",
				ShowPath:    "/opt/jellyfin/media/shows",
				ManagerPath: "/opt/media_manager",
				DryRun:      true,
			},
		},
		{
			name: "all environment variables set",
			envVariables: map[string]string{
				"ENIACORE_TORRENT_PATH": "/custom/downloads",
				"ENIACORE_MOVIE_PATH":   "/custom/movies",
				"ENIACORE_SHOW_PATH":    "/custom/shows",
				"ENIACORE_MANAGER_PATH": "/custom/manager",
				"ENIACORE_DRY_RUN":      "true",
			},
			args: []string{},
			expected: &Config{
				TorrentPath: "/custom/downloads",
				MoviePath:   "/custom/movies",
				ShowPath:    "/custom/shows",
				ManagerPath: "/custom/manager",
				DryRun:      true,
			},
		},
		{
			name: "flags override environment variables",
			envVariables: map[string]string{
				"ENIACORE_TORRENT_PATH": "/env/downloads",
				"ENIACORE_DRY_RUN":      "false",
			},
			args: []string{
				"-torrent-path", "/flag/downloads",
				"-dry-run",
			},
			expected: &Config{
				TorrentPath: "/flag/downloads",
				MoviePath:   "/opt/jellyfin/media/movies",
				ShowPath:    "/opt/jellyfin/media/shows",
				ManagerPath: "/opt/media_manager",
				DryRun:      true,
			},
		},
		{
			name:         "partial environment variables set",
			envVariables: map[string]string{
				"ENIACORE_TORRENT_PATH": "/custom/downloads",
				"ENIACORE_DRY_RUN":      "true",
			},
			args: []string{},
			expected: &Config{
				TorrentPath: "/custom/downloads",
				MoviePath:   "/opt/jellyfin/media/movies",
				ShowPath:    "/opt/jellyfin/media/shows",
				ManagerPath: "/opt/media_manager",
				DryRun:      true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clear environment and reset flags before each test
			clearEnv()
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set test environment variables
			for key, value := range test.envVariables {
				os.Setenv(key, value)
			}
			defer clearEnv()

			// Set os.Args for flag parsing
			oldArgs := os.Args
			os.Args = append([]string{"cmd"}, test.args...)
			defer func() { os.Args = oldArgs }()

			cfg := New()

			if cfg.TorrentPath != test.expected.TorrentPath {
				t.Errorf("TorrentPath = %v, want %v", cfg.TorrentPath, test.expected.TorrentPath)
			}
			if cfg.MoviePath != test.expected.MoviePath {
				t.Errorf("MoviePath = %v, want %v", cfg.MoviePath, test.expected.MoviePath)
			}
			if cfg.ShowPath != test.expected.ShowPath {
				t.Errorf("ShowPath = %v, want %v", cfg.ShowPath, test.expected.ShowPath)
			}
			if cfg.ManagerPath != test.expected.ManagerPath {
				t.Errorf("ManagerPath = %v, want %v", cfg.ManagerPath, test.expected.ManagerPath)
			}
			if cfg.DryRun != test.expected.DryRun {
				t.Errorf("DryRun = %v, want %v", cfg.DryRun, test.expected.DryRun)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		envValue      string
		expectedValue string
		setEnv        bool
	}{
		{
			name:          "env variable set to custom",
			key:           "TEST_KEY",
			defaultValue:  "default",
			envValue:      "custom",
			expectedValue: "custom",
			setEnv:        true,
		},
		{
			name:          "env variable not set uses default",
			key:           "TEST_KEY",
			defaultValue:  "default",
			envValue:      "",
			expectedValue: "default",
			setEnv:        false,
		},
		{
			name:          "env variable set to empty string uses default",
			key:           "TEST_KEY",
			defaultValue:  "default",
			envValue:      "",
			expectedValue: "default",
			setEnv:        true,
		},
		{
			name:          "env variable not set with empty default",
			key:           "TEST_KEY",
			defaultValue:  "",
			envValue:      "",
			expectedValue: "",
			setEnv:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Unsetenv(test.key)

			if test.setEnv {
				os.Setenv(test.key, test.envValue)
			}
			defer os.Unsetenv(test.key)

			result := getEnv(test.key, test.defaultValue)

			if result != test.expectedValue {
				t.Errorf("getEnv() = %v, want %v", result, test.expectedValue)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  bool
		envValue      string
		expectedValue bool
		setEnv        bool
	}{
		{
			name:          "env variable set to true",
			key:           "TEST_BOOL",
			defaultValue:  false,
			envValue:      "true",
			expectedValue: true,
			setEnv:        true,
		},
		{
			name:          "env variable set to false",
			key:           "TEST_BOOL",
			defaultValue:  true,
			envValue:      "false",
			expectedValue: false,
			setEnv:        true,
		},
		{
			name:          "env variable not set uses default true",
			key:           "TEST_BOOL",
			defaultValue:  true,
			envValue:      "",
			expectedValue: true,
			setEnv:        false,
		},
		{
			name:          "env variable not set uses default false",
			key:           "TEST_BOOL",
			defaultValue:  false,
			envValue:      "",
			expectedValue: false,
			setEnv:        false,
		},
		{
			name:          "env variable set to invalid value uses default",
			key:           "TEST_BOOL",
			defaultValue:  false,
			envValue:      "invalid",
			expectedValue: false,
			setEnv:        true,
		},
		{
			name:          "env variable set to 1 (true)",
			key:           "TEST_BOOL",
			defaultValue:  false,
			envValue:      "1",
			expectedValue: true,
			setEnv:        true,
		},
		{
			name:          "env variable set to 0 (false)",
			key:           "TEST_BOOL",
			defaultValue:  true,
			envValue:      "0",
			expectedValue: false,
			setEnv:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Unsetenv(test.key)

			if test.setEnv {
				os.Setenv(test.key, test.envValue)
			}
			defer os.Unsetenv(test.key)

			result := getEnvBool(test.key, test.defaultValue)
			if result != test.expectedValue {
				t.Errorf("getEnvBool() = %v, want %v", result, test.expectedValue)
			}
		})
	}
}

func clearEnv() {
	os.Unsetenv("ENIACORE_TORRENT_PATH")
	os.Unsetenv("ENIACORE_MOVIE_PATH")
	os.Unsetenv("ENIACORE_SHOW_PATH")
	os.Unsetenv("ENIACORE_MANAGER_PATH")
	os.Unsetenv("ENIACORE_DRY_RUN")
}
