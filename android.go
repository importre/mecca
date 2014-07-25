// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

// AndroidCrawler gathers android repositories that exist on Github and
// dumps to json file.
type AndroidCrawler struct {
	client    *github.Client
	appRepos  []AndroidRepository
	libRepos  []AndroidRepository
	lRepos    []AndroidRepository
	wearRepos []AndroidRepository
}

// AndroidRepository represents an android repository.
// It embeds github.Repository basically.
type AndroidRepository struct {
	Library  *bool `json:"library,omitempty"`
	AndroidL *bool `json:"android_l,omitempty"`
	Wearable *bool `json:"wearable,omitempty"`
	github.Repository
}

// NewAndroidCrawler returns a new AndroidCrawler.
func NewAndroidCrawler(accessToken string) *AndroidCrawler {
	if accessToken == "" {
		log.Println("Github AccessToken is empty.")
		return &AndroidCrawler{
			client: github.NewClient(nil),
		}
	}

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: GITHUB_TOKEN},
	}

	return &AndroidCrawler{
		client: github.NewClient(t.Client()),
	}
}

func (self *AndroidCrawler) Dump() {
	outDir := "data/android/"
	os.MkdirAll(outDir, 0755)
	DumpToJson(outDir+"app_repos.json", self.appRepos)
	DumpToJson(outDir+"lib_repos.json", self.libRepos)
	DumpToJson(outDir+"l_repos.json", self.lRepos)
	DumpToJson(outDir+"wear_repos.json", self.wearRepos)
}

// FindRepos finds repositories.
func (self *AndroidCrawler) FindRepos(stars int) {
	capacity := 100
	self.appRepos = make([]AndroidRepository, 0, capacity)  // app repos
	self.libRepos = make([]AndroidRepository, 0, capacity)  // library repos
	self.lRepos = make([]AndroidRepository, 0, capacity)    // android-l repos
	self.wearRepos = make([]AndroidRepository, 0, capacity) // wearable repos

	page := 1
	maxPage := math.MaxInt32

	query := fmt.Sprintf("android fork:true stars:>=%v", stars)
	//query := "android fork:true stars:840..850"
	opts := &github.SearchOptions{
		Sort: "stars",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for page <= maxPage {
		opts.Page = page
		result, response, err := self.client.Search.Repositories(query, opts)
		Wait(response)

		if err != nil {
			log.Fatal("FindRepos:", err)
		}

		maxPage = response.LastPage

		msg := fmt.Sprintf("page: %v/%v, size: %v, total: %v",
			page, maxPage, len(result.Repositories), *result.Total)
		log.Println(msg)

		for _, repo := range result.Repositories {
			if !self.IsAndroid(&repo) {
				continue
			}

			// check repo if it has library or not
			library := self.IsLibrary(&repo)

			// check repo if it has android-l or not
			androidL := self.IsAndroidL(&repo)

			// check wearable
			wearable := self.IsWearable(&repo)

			RemoveUnnecessaryFeatures(&repo)

			androidRepo := AndroidRepository{
				&library,
				&androidL,
				&wearable,
				repo,
			}

			if library {
				self.libRepos = append(self.libRepos, androidRepo)
			} else {
				self.appRepos = append(self.appRepos, androidRepo)
			}

			if androidL {
				self.lRepos = append(self.lRepos, androidRepo)
			}

			if wearable {
				self.wearRepos = append(self.wearRepos, androidRepo)
			}

			self.downloadAvatarImage(&repo)
			log.Println("android:", *repo.FullName)
			time.Sleep(time.Millisecond * 500)
		}

		page++
	}
}

func (self *AndroidCrawler) downloadAvatarImage(repo *github.Repository) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
		}
	}()

	outDir := "images/"
	os.Mkdir(outDir, 0755)

	outPath := outDir + *repo.Owner.Login
	out, err := os.Create(outPath)
	defer out.Close()
	if err != nil {
		return
	}

	resp, err := http.Get(*repo.Owner.AvatarURL)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	repo.Owner.AvatarURL = &outPath
}

// IsAndroid returns true if the repo has android project otherwise false.
func (self *AndroidCrawler) IsAndroid(repo *github.Repository) bool {
	opts := &github.SearchOptions{}

	// I'm not sure exactly the rule is good.
	query := fmt.Sprintf("androidmanifest.xml in:path repo:%v", *repo.FullName)
	result, response, err := self.client.Search.Code(query, opts)
	Wait(response)

	if err != nil {
		log.Fatal("IsAndroid:", err)
	}

	return *result.Total > 0
}

// IsLibrary returns true if the repo has library project otherwise false.
func (self *AndroidCrawler) IsLibrary(repo *github.Repository) bool {
	opts := &github.SearchOptions{}

	// I'm not sure exactly the rule is good.
	queries := [...]string{
		fmt.Sprintf("\"android-library\" extension:gradle language:groovy repo:%v", *repo.FullName),
		fmt.Sprintf("\"android.library=true\" language:ini repo:%v", *repo.FullName),
	}

	for _, query := range queries {
		time.Sleep(time.Millisecond * 100)
		result, response, err := self.client.Search.Code(query, opts)
		Wait(response)

		if err != nil {
			log.Fatal("IsLibrary:", err)
		}

		if *result.Total > 0 {
			for _, codeResult := range result.CodeResults {
				data := strings.Split(*codeResult.Path, "/")
				if len(data) <= 2 {
					return true
				}
			}
		}
	}

	if self.HasLibraryLiteral(repo.Description) {
		return true
	}

	return false
}

func (self *AndroidCrawler) HasLibraryLiteral(str *string) bool {
	if str == nil {
		return false
	}

	re, err := regexp.Compile(`(?i)\blibrary\b`)
	if err != nil {
		return false
	}

	res := re.FindAllString(*str, 1)
	if len(res) > 0 {
		return true
	}

	return false
}

// IsAndroidL returns true if the repo has android-l project otherwise false.
func (self *AndroidCrawler) IsAndroidL(repo *github.Repository) bool {
	opts := &github.SearchOptions{}

	// I'm not sure exactly the rule is good.
	format := "\"compileSdkVersion android-l\" extension:gradle language:groovy repo:%v"
	query := fmt.Sprintf(format, *repo.FullName)

	result, response, err := self.client.Search.Code(query, opts)
	Wait(response)

	if err != nil {
		log.Fatal("IsLibrary:", err)
	}

	return *result.Total > 0
}

// IsWearable returns true if the repo has wearable project otherwise false.
func (self *AndroidCrawler) IsWearable(repo *github.Repository) bool {
	opts := &github.SearchOptions{}

	// I'm not sure exactly the rule is good.
	format := "\"com.google.android.support:wearable\" extension:gradle language:groovy repo:%v"
	query := fmt.Sprintf(format, *repo.FullName)

	result, response, err := self.client.Search.Code(query, opts)
	Wait(response)

	if err != nil {
		log.Fatal("IsWearable:", err)
	}

	return *result.Total > 0
}
