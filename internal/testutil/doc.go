// Package testutil provides testing utilities for creating test metadata entries and filesystem structures.
//
// # Test Entry Builders
//
// The package provides builder functions for creating pre-populated [metadata.Entry] instances:
//
//	Files:        CreateTestSubFile, CreateTestBonusFile, CreateTestEpFile, CreateTestMovieFile
//	Directories:  CreateTestSubDir, CreateTestBonusDir, CreateTestSeasonDir, CreateTestSeriesDir, CreateTestMovieDir
//
// Each builder creates an entry with realistic metadata that matches expected patterns
// for its role, suitable for testing classification, processing, and transfer logic.
//
// # Test Environment Setup
//
// Helper functions for test infrastructure:
//
//	CreateTestDir()        - Creates temporary directory with source/movies/shows/manager subdirs
//	CreateTestCfg()        - Creates test configuration pointing to test directories
//	CreateTestFiles()      - Recursively creates actual files on filesystem for integration tests
//
// # Usage Example
//
//	func TestProcessor(t *testing.T) {
//	    testDir := testutil.CreateTestDir(t)
//	    cfg := testutil.CreateTestCfg(testDir)
//
//	    movieFile := testutil.CreateTestMovieFile(nil)
//	    testutil.CreateTestFiles(movieFile, testDir)
//
//	    // Test processing logic...
//	}
//
// # Implementation
//
// All test entries use consistent test values:
//
//	Title: ["TEST", "TITLE"]
//	Year: 2025
//	Resolution: "1080p"
//	Codec: "x264"
//	Source: "Remux"
//	Audio: "Atmos"
package testutil
