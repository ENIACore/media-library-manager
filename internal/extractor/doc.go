// Copyright 2026. All rights reserved.
// Use of this source code is governed by the GNU AGLPv3
// license that can be found in the LICENSE.md file.

/*
Package extractor provides extraction functionality for full paths, enabling the creation of [metadata.Entry] objects
The main three functions are [Filter], [ExtractMedia], and [ExtractPath].

The main purpose of all three functions is to assist parser in creating an [Entry] tree structure that represents a movie or show.

# Pattern Matching Basics

To accurately match patterns file and directory names are sanitized via [sanitizeName]. This ensures special characters are not included in pattern matching, and individual words can be accurately recognized.

The sanitized name is stored in []string format; to support multi-word pattern matching the function [matchSegments] is used in combination with regex patterns that use '\.' as word seperators.

# Main Functions

[ExtractMedia] extracts information about the movie or show a file or directory contains. It does this using advanced pattern recognition of the last element (file or directory) in the path provided. This is used in the creation of the [metadata.MediaInfo] field, of [metadata.Entry] objects.

[ExtractPath] extracts file system information about a file or directory. It does this using golang's os package for accurate information, along with helper functions. This is used in the creation of the [metadata.PathInfo] field, of [metadata.Entry] objects.

[Filter] determines if a file has a recognized file extension. It does this using golang's os package for accurate information. This is used to prevent creation of unwanted [metadata.Entry] objects in [parser.Parse]
*/
package extractor
