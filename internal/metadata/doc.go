// Copyright 2026. All rights reserved.
// Use of this source code is governed by the GNU AGLPv3
// license that can be found in the LICENSE.md file.

/*
Package metadata provides abstraction types to describe downloaded media files and directories.
The main type provided is [Entry] which represents a file or directory and is consumed and processed by other internal packages.

[Entry] objects main purpose is to encapsulate data that enables other internal packages to determine the resulting full destination path from a file or directories full source path.

# Basics

A [Entry] object uses a tree data structure, whose root [Entry] object is a directory or file for the movie, show, or episode to be processed.

A [Entry] object has three main fields, Role, [MediaInfo], and [PathInfo]. [MediaInfo] contains information that describes the movie, show, or episode the file or directory pertains too. [PathInfo] contains information that describes the file or directory. Role describes what the files or directories purpose is (to hold movies, to hold subtitles, ect..).

Each processing step in the program will mutate the [Entry] tree further until the destination (or lack of) for each [Entry] can be determined.

# Recognized Structures

Subtitle and bonus directories can be embedded in their own type, once, to allow recognition of intermediary subtitle or bonus directories (typically found in multi-season or episode bonus or subtitle directories).

	Episode File                        (level >= 0)
	Movie File                          (level >= 0)
	Bonus File							(level >= 1)
	Subtitle File						(level >= 1)

	Subtitle Directory                  (level >= 1)
	├── Subtitle File(s)
	└── Subtitle Directory(s)           (optional)

	Bonus Directory                     (level >= 1)
	├── Bonus File(s)
	├── Subtitle File(s)                (optional)
	└── Bonus Directory(s)           	(optional)

	Season Directory                    (level >= 0)
	├── Episode File(s)
	├── Subtitle File(s)                (optional)
	└── Subtitle Directory              (optional)

	Series Directory                    (level = 0)
	├── Season Directory(s)
	├── Bonus Directory                 (optional)
	└── Subtitle Directory              (optional)

	Movie Directory                     (level = 0)
	├── Movie File
	├── Subtitle File(s)                (optional)
	├── Bonus File(s)                   (optional)
	├── Subtitle Directory              (optional)
	└── Bonus Directory                 (optional)
*/
package metadata

