// Copyright 2026. All rights reserved.
// Use of this source code is governed by the GNU AGLPv3
// license that can be found in the LICENSE.md file.

/*
Package classifier traverses [metadata.Entry] tree and classifies (i.e assigns [metadata.EntryRole]) each entry. 

# Basics

Classifier expects structure found in metadata 'Recognized Structures' documentation.

To determine role of each entry, and if the overall tree matches a recognized structure the classifier uses 3 things:
	- Recursive result of children classifications (i.e are all children classified with an expected role)
	- Patterns found in file name (stored in [metadata.MediaInfo] field)
	- Height of entry from furthest descendent

# Classification Roles

Entries are classified into one of the following roles:

	Files:        SubtitleFile, BonusFile, EpisodeFile, MovieFile
	Directories:  SubtitleDir, BonusDir, SeasonDir, SeriesDir, MovieDir

# Pipeline

Package relies on parser package for valid tree and produces output used by enricher package.  
*/
package classifier
