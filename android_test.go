// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "testing"

func TestNewAndroidCrawler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	android := NewAndroidCrawler()
	if android == nil {
		t.Fail()
		return
	}
}
