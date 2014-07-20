// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "testing"

func TestHasLibraryLiteral(t *testing.T) {
	var description string
	android := NewAndroidCrawler("")

	description = "LIBRARY"
	if !android.HasLibraryLiteral(&description) {
		t.Fail()
	}

	description = "An Asynchronous HTTP Library for Android"
	if !android.HasLibraryLiteral(&description) {
		t.Fail()
	}

	description = "A powerful image downloading and caching library for Android"
	if !android.HasLibraryLiteral(&description) {
		t.Fail()
	}
}
