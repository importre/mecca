// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Crawler interface {
	FindRepos(stars int)
  Dump()
}
