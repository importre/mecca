// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	helpRepoType := "Supported only 'android' currently"

	stars = flag.Int("stars", DEFAULT_STARS, helpStars)
	repoType = flag.String("type", DEFAULT_REPO_TYPE, helpRepoType)
}

func dumpAndroidRepos() {
	android := NewAndroidCrawler()
	reposList := android.FindRepos(*stars)

	repos := reposList[0]
	libRepos := reposList[1]
	lRepos := reposList[2]
	wearRepos := reposList[3]

	outDir := "data"
	os.Mkdir(outDir, 0755)
	android.DumpToJson(outDir+"/android.json", repos)
	android.DumpToJson(outDir+"/lib_repos.json", libRepos)
	android.DumpToJson(outDir+"/l_repos.json", lRepos)
	android.DumpToJson(outDir+"/wear_repos.json", wearRepos)
}

func main() {
	flag.Parse()

	switch *repoType {
	case "android":
		dumpAndroidRepos()
	default:
		log.Fatal("Unknown repo type: ", *repoType)
	}
}
