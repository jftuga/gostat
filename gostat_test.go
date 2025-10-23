/*
gostat_test.go - Comprehensive unit and integration tests for gostat utility.

Tests core functionality including number formatting, date parsing, glob expansion,
and the actual file timestamp modification operations (-a, -m, -b flags).
Includes both unit tests for individual functions and integration tests that verify
end-to-end behavior with real file operations.
*/

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFormatWithCommas verifies that integers are formatted with proper thousand separators.
func TestFormatWithCommas(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0"},
		{"small number", 42, "42"},
		{"hundreds", 999, "999"},
		{"thousands", 1000, "1,000"},
		{"ten thousands", 12345, "12,345"},
		{"millions", 1234567, "1,234,567"},
		{"billions", 1234567890, "1,234,567,890"},
		{"negative", -1234567, "-1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatWithCommas(tt.input)
			if result != tt.expected {
				t.Errorf("FormatWithCommas(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCreateDate verifies that valid date strings are correctly parsed into time.Time objects.
func TestCreateDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		year      int
		month     time.Month
		day       int
		hour      int
		minute    int
		second    int
	}{
		{
			name:      "valid date",
			input:     "20210329.143025",
			shouldErr: false,
			year:      2021,
			month:     time.March,
			day:       29,
			hour:      14,
			minute:    30,
			second:    25,
		},
		{
			name:      "new year",
			input:     "20250101.000000",
			shouldErr: false,
			year:      2025,
			month:     time.January,
			day:       1,
			hour:      0,
			minute:    0,
			second:    0,
		},
		{
			name:      "leap year date",
			input:     "20240229.120000",
			shouldErr: false,
			year:      2024,
			month:     time.February,
			day:       29,
			hour:      12,
			minute:    0,
			second:    0,
		},
		{
			name:      "invalid format - missing period",
			input:     "20210329143025",
			shouldErr: true,
		},
		{
			name:      "invalid format - too short",
			input:     "2021.143025",
			shouldErr: true,
		},
		{
			name:      "invalid date - month 13",
			input:     "20211329.143025",
			shouldErr: true,
		},
		{
			name:      "invalid date - day 32",
			input:     "20210332.143025",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := createDate(tt.input)

			// Check error expectation
			if tt.shouldErr {
				if err == nil {
					t.Errorf("createDate(%s) expected error but got none", tt.input)
				}
				return
			}

			// No error expected - verify the parsed date components
			if err != nil {
				t.Errorf("createDate(%s) unexpected error: %v", tt.input, err)
				return
			}

			if result.Year() != tt.year {
				t.Errorf("createDate(%s) year = %d; want %d", tt.input, result.Year(), tt.year)
			}
			if result.Month() != tt.month {
				t.Errorf("createDate(%s) month = %v; want %v", tt.input, result.Month(), tt.month)
			}
			if result.Day() != tt.day {
				t.Errorf("createDate(%s) day = %d; want %d", tt.input, result.Day(), tt.day)
			}
			if result.Hour() != tt.hour {
				t.Errorf("createDate(%s) hour = %d; want %d", tt.input, result.Hour(), tt.hour)
			}
			if result.Minute() != tt.minute {
				t.Errorf("createDate(%s) minute = %d; want %d", tt.input, result.Minute(), tt.minute)
			}
			if result.Second() != tt.second {
				t.Errorf("createDate(%s) second = %d; want %d", tt.input, result.Second(), tt.second)
			}
		})
	}
}

// TestExpandGlobs verifies that file glob patterns are correctly expanded.
// This test creates temporary files to work with.
func TestExpandGlobs(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gostat_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{"test1.txt", "test2.txt", "data.csv"}
	for _, filename := range testFiles {
		filepath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filepath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
		description   string
	}{
		{
			name:          "all txt files",
			pattern:       filepath.Join(tempDir, "*.txt"),
			expectedCount: 2,
			description:   "should find both .txt files",
		},
		{
			name:          "specific file",
			pattern:       filepath.Join(tempDir, "data.csv"),
			expectedCount: 1,
			description:   "should find the exact file",
		},
		{
			name:          "all files",
			pattern:       filepath.Join(tempDir, "*"),
			expectedCount: 3,
			description:   "should find all files",
		},
		{
			name:          "no matches",
			pattern:       filepath.Join(tempDir, "*.pdf"),
			expectedCount: 0,
			description:   "should return empty slice when no files match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandGlobs([]string{tt.pattern})
			if len(result) != tt.expectedCount {
				t.Errorf("expandGlobs(%s) returned %d files; want %d (%s)",
					tt.pattern, len(result), tt.expectedCount, tt.description)
			}
		})
	}
}

// TestExpandGlobsMultiplePatterns verifies that multiple glob patterns can be processed together.
func TestExpandGlobsMultiplePatterns(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gostat_test_multi_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{"file1.txt", "file2.log", "file3.txt"}
	for _, filename := range testFiles {
		filepath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filepath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test with multiple patterns
	patterns := []string{
		filepath.Join(tempDir, "*.txt"),
		filepath.Join(tempDir, "*.log"),
	}

	result := expandGlobs(patterns)
	expectedCount := 3 // 2 txt files + 1 log file

	if len(result) != expectedCount {
		t.Errorf("expandGlobs with multiple patterns returned %d files; want %d",
			len(result), expectedCount)
	}
}

// TestGetFileTimes verifies that file timestamp metadata can be retrieved correctly.
func TestGetFileTimes(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "gostat_time_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Get file times
	times := getFileTimes(tempFile.Name())

	// Verify that we got at least access and modify times (always available)
	if _, found := times["a"]; !found {
		t.Error("getFileTimes() did not return access time")
	}
	if _, found := times["m"]; !found {
		t.Error("getFileTimes() did not return modify time")
	}

	// Note: birth time and change time availability depends on the file system
	// So we don't fail if they're missing, just verify the function doesn't crash
}

// TestSetFileTimeModify tests the -m flag: modifying only the modification time.
func TestSetFileTimeModify(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "gostat_modify_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Get original times
	originalTimes := getFileTimes(tempFile.Name())
	originalAccess := originalTimes["a"]
	originalModify := originalTimes["m"]

	// Wait a moment to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Set a new modify time (using a past date to avoid future date issues)
	newTimeStr := "20200115.120000"
	setFileTime([]string{tempFile.Name()}, newTimeStr, OpModify)

	// Get updated times
	updatedTimes := getFileTimes(tempFile.Name())
	updatedAccess := updatedTimes["a"]
	updatedModify := updatedTimes["m"]

	// Verify access time was NOT changed (within 1 second tolerance for file system precision)
	accessDiff := updatedAccess.Sub(originalAccess).Abs()
	if accessDiff > time.Second {
		t.Errorf("setFileTime with OpModify changed access time: original=%v, updated=%v, diff=%v",
			originalAccess, updatedAccess, accessDiff)
	}

	// Verify modify time WAS changed
	expectedModify, _ := createDate(newTimeStr)
	modifyDiff := updatedModify.Sub(expectedModify).Abs()
	if modifyDiff > time.Second {
		t.Errorf("setFileTime with OpModify did not set modify time correctly: expected=%v, got=%v, diff=%v",
			expectedModify, updatedModify, modifyDiff)
	}

	// Verify the modify time actually changed from original
	if updatedModify.Equal(originalModify) {
		t.Error("setFileTime with OpModify did not change the modify time")
	}
}

// TestSetFileTimeAccess tests the -a flag: modifying only the access time.
func TestSetFileTimeAccess(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "gostat_access_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Get original times
	originalTimes := getFileTimes(tempFile.Name())
	originalAccess := originalTimes["a"]
	originalModify := originalTimes["m"]

	// Wait a moment to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Set a new access time (using a past date)
	newTimeStr := "20200215.140000"
	setFileTime([]string{tempFile.Name()}, newTimeStr, OpAccess)

	// Get updated times
	updatedTimes := getFileTimes(tempFile.Name())
	updatedAccess := updatedTimes["a"]
	updatedModify := updatedTimes["m"]

	// Verify modify time was NOT changed (within 1 second tolerance)
	modifyDiff := updatedModify.Sub(originalModify).Abs()
	if modifyDiff > time.Second {
		t.Errorf("setFileTime with OpAccess changed modify time: original=%v, updated=%v, diff=%v",
			originalModify, updatedModify, modifyDiff)
	}

	// Verify access time WAS changed
	expectedAccess, _ := createDate(newTimeStr)
	accessDiff := updatedAccess.Sub(expectedAccess).Abs()
	if accessDiff > time.Second {
		t.Errorf("setFileTime with OpAccess did not set access time correctly: expected=%v, got=%v, diff=%v",
			expectedAccess, updatedAccess, accessDiff)
	}

	// Verify the access time actually changed from original
	if updatedAccess.Equal(originalAccess) {
		t.Error("setFileTime with OpAccess did not change the access time")
	}
}

// TestSetFileTimeBoth tests the -b flag: modifying both access and modify times.
func TestSetFileTimeBoth(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "gostat_both_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Get original times
	originalTimes := getFileTimes(tempFile.Name())
	originalAccess := originalTimes["a"]
	originalModify := originalTimes["m"]

	// Wait a moment to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Set both times to the same new value
	newTimeStr := "20200315.160000"
	setFileTime([]string{tempFile.Name()}, newTimeStr, OpBoth)

	// Get updated times
	updatedTimes := getFileTimes(tempFile.Name())
	updatedAccess := updatedTimes["a"]
	updatedModify := updatedTimes["m"]

	expectedTime, _ := createDate(newTimeStr)

	// Verify both times were changed to the new value
	accessDiff := updatedAccess.Sub(expectedTime).Abs()
	if accessDiff > time.Second {
		t.Errorf("setFileTime with OpBoth did not set access time correctly: expected=%v, got=%v, diff=%v",
			expectedTime, updatedAccess, accessDiff)
	}

	modifyDiff := updatedModify.Sub(expectedTime).Abs()
	if modifyDiff > time.Second {
		t.Errorf("setFileTime with OpBoth did not set modify time correctly: expected=%v, got=%v, diff=%v",
			expectedTime, updatedModify, modifyDiff)
	}

	// Verify both times changed from their originals
	if updatedAccess.Equal(originalAccess) {
		t.Error("setFileTime with OpBoth did not change the access time")
	}
	if updatedModify.Equal(originalModify) {
		t.Error("setFileTime with OpBoth did not change the modify time")
	}

	// Verify both times are equal to each other (within tolerance)
	timeDiff := updatedAccess.Sub(updatedModify).Abs()
	if timeDiff > time.Second {
		t.Errorf("setFileTime with OpBoth: access and modify times differ: access=%v, modify=%v, diff=%v",
			updatedAccess, updatedModify, timeDiff)
	}
}

// TestSetFileTimeMultipleFiles tests that setFileTime correctly handles multiple files.
func TestSetFileTimeMultipleFiles(t *testing.T) {
	// Create multiple temporary files
	tempDir, err := os.MkdirTemp("", "gostat_multi_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileNames := []string{"file1.txt", "file2.txt", "file3.txt"}
	var filePaths []string

	for _, name := range fileNames {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
		filePaths = append(filePaths, path)
	}

	// Set modify time for all files
	newTimeStr := "20200420.180000"
	setFileTime(filePaths, newTimeStr, OpModify)

	// Verify all files were updated
	expectedTime, _ := createDate(newTimeStr)
	for _, path := range filePaths {
		times := getFileTimes(path)
		modifyTime := times["m"]

		diff := modifyTime.Sub(expectedTime).Abs()
		if diff > time.Second {
			t.Errorf("File %s: modify time not set correctly: expected=%v, got=%v, diff=%v",
				path, expectedTime, modifyTime, diff)
		}
	}
}

// TestSetFileTimeWithGlob tests that setFileTime works with glob patterns.
func TestSetFileTimeWithGlob(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "gostat_glob_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some .txt files and one .log file
	txtFiles := []string{"test1.txt", "test2.txt", "test3.txt"}
	for _, name := range txtFiles {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	logFile := filepath.Join(tempDir, "test.log")
	if err := os.WriteFile(logFile, []byte("log"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Use glob pattern to set times for only .txt files
	globPattern := filepath.Join(tempDir, "*.txt")
	newTimeStr := "20200525.200000"
	setFileTime([]string{globPattern}, newTimeStr, OpModify)

	expectedTime, _ := createDate(newTimeStr)

	// Verify .txt files were updated
	for _, name := range txtFiles {
		path := filepath.Join(tempDir, name)
		times := getFileTimes(path)
		modifyTime := times["m"]

		diff := modifyTime.Sub(expectedTime).Abs()
		if diff > time.Second {
			t.Errorf("File %s: modify time not set correctly: expected=%v, got=%v, diff=%v",
				name, expectedTime, modifyTime, diff)
		}
	}

	// Verify .log file was NOT updated (should have recent timestamp)
	logTimes := getFileTimes(logFile)
	logModifyTime := logTimes["m"]
	logDiff := logModifyTime.Sub(expectedTime).Abs()

	// Log file should have a timestamp very different from the set time (likely current time)
	if logDiff < 24*time.Hour {
		t.Errorf("Log file was incorrectly modified by glob pattern that should only match .txt files")
	}
}

// TestShowFileTimes verifies that showFileTimes returns the correct count of files processed.
func TestShowFileTimes(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "gostat_show_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	fileCount := 3
	var filePaths []string
	for i := 0; i < fileCount; i++ {
		path := filepath.Join(tempDir, "file"+string(rune('1'+i))+".txt")
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		filePaths = append(filePaths, path)
	}

	// Test showFileTimes returns correct count
	count := showFileTimes(filePaths)
	if count != fileCount {
		t.Errorf("showFileTimes() returned count=%d; want %d", count, fileCount)
	}

	// Test with non-existent file - should return 0
	nonExistentPath := filepath.Join(tempDir, "does_not_exist.txt")
	count = showFileTimes([]string{nonExistentPath})
	if count != 0 {
		t.Errorf("showFileTimes() with non-existent file returned count=%d; want 0", count)
	}
}
