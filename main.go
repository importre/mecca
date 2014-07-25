// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	stars    *int
	repoType *string
)

const (
	DEFAULT_STARS     = 200
	DEFAULT_REPO_TYPE = "android"
)

func init() {
	helpStars := fmt.Sprintf("Searching repositories which have %v stars or more than", DEFAULT_STARS)
	helpRepoType := "Supported 'polymer' and 'android'"

	stars = flag.Int("stars", DEFAULT_STARS, helpStars)
	repoType = flag.String("type", DEFAULT_REPO_TYPE, helpRepoType)
}

func main() {
	flag.Parse()
	var crawler Crawler

	switch *repoType {
	case "android":
		crawler = NewAndroidCrawler(GITHUB_TOKEN)
	case "polymer":
		crawler = NewPolymerCrawler(GITHUB_TOKEN)
	default:
		log.Fatal("Unknown repo type: ", *repoType)
	}

	crawler.FindRepos(*stars)
	crawler.Dump()
	NewMetaData().Dump()
}
