// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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

const (
	REMAINING_THRESHOLD = 1
)

// AndroidCrawler gathers android repositories that exist on Github and
// dumps to json file.
type AndroidCrawler struct {
	client *github.Client
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

// DumpToJson writes a json file.
func (self *AndroidCrawler) DumpToJson(path string, repos []AndroidRepository) {
	b, err := json.Marshal(repos)
	if err != nil {
		log.Fatal("DumpToJson:", err)
	}

	var buffer bytes.Buffer
	json.Indent(&buffer, b, "", "  ")
	writeToFile(path, buffer)
}

// FindRepos returns AndroidRepository list.
// 0: app repositories
// 1: library repositories
// 2: android-l repositories
// 3: wearable repositories
func (self *AndroidCrawler) FindRepos(stars int) [4][]AndroidRepository {
	capacity := 100
	appRepos := make([]AndroidRepository, 0, capacity)  // app repos
	libRepos := make([]AndroidRepository, 0, capacity)  // library repos
	lRepos := make([]AndroidRepository, 0, capacity)    // android-l repos
	wearRepos := make([]AndroidRepository, 0, capacity) // wearable repos

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
		self.sleep(response)

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

			androidRepo := AndroidRepository{
				&library,
				&androidL,
				&wearable,
				repo,
			}

			self.removeUnnecessaryFeatures(&androidRepo)

			if library {
				libRepos = append(libRepos, androidRepo)
			} else {
				appRepos = append(appRepos, androidRepo)
			}

			if androidL {
				lRepos = append(lRepos, androidRepo)
			}

			if wearable {
				wearRepos = append(wearRepos, androidRepo)
			}

			self.downloadAvatarImage(&repo)
			log.Println("android:", *repo.FullName)
			time.Sleep(time.Millisecond * 500)
		}

		page++
	}

	return [...][]AndroidRepository{
		appRepos,
		libRepos,
		lRepos,
		wearRepos,
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
	self.sleep(response)

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
		self.sleep(response)

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
	self.sleep(response)

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
	self.sleep(response)

	if err != nil {
		log.Fatal("IsWearable:", err)
	}

	return *result.Total > 0
}

func (self *AndroidCrawler) sleep(response *github.Response) {
	if response != nil && response.Remaining <= REMAINING_THRESHOLD {
		gap := time.Duration(response.Reset.Local().Unix() - time.Now().Unix())
		sleep := gap * time.Second
		if sleep < 0 {
			sleep = -sleep
		}

		time.Sleep(sleep)
	}
}

func (self *AndroidCrawler) removeUnnecessaryFeatures(repo *AndroidRepository) {
	repo.ArchiveURL = nil
	repo.AssigneesURL = nil
	repo.BlobsURL = nil
	repo.BranchesURL = nil
	repo.CloneURL = nil
	repo.CollaboratorsURL = nil
	repo.CommentsURL = nil
	repo.CommitsURL = nil
	repo.CompareURL = nil
	repo.ContentsURL = nil
	repo.ContributorsURL = nil
	repo.DownloadsURL = nil
	repo.EventsURL = nil
	repo.ForksURL = nil
	repo.GitCommitsURL = nil
	repo.GitRefsURL = nil
	repo.GitTagsURL = nil
	repo.GitURL = nil
	repo.HooksURL = nil
	repo.IssueCommentURL = nil
	repo.IssueEventsURL = nil
	repo.IssuesURL = nil
	repo.KeysURL = nil
	repo.LabelsURL = nil
	repo.LanguagesURL = nil
	repo.MergesURL = nil
	repo.MilestonesURL = nil
	repo.NotificationsURL = nil
	repo.Owner.EventsURL = nil
	repo.Owner.FollowersURL = nil
	repo.Owner.FollowingURL = nil
	repo.Owner.GistsURL = nil
	repo.Owner.OrganizationsURL = nil
	repo.Owner.ReceivedEventsURL = nil
	repo.Owner.ReposURL = nil
	repo.Owner.StarredURL = nil
	repo.Owner.SubscriptionsURL = nil
	repo.PullsURL = nil
	repo.ReleasesURL = nil
	repo.SSHURL = nil
	repo.SVNURL = nil
	repo.StargazersURL = nil
	repo.StatusesURL = nil
	repo.SubscribersURL = nil
	repo.SubscriptionURL = nil
	repo.TagsURL = nil
	repo.TeamsURL = nil
	repo.TreesURL = nil
}

func writeToFile(name string, buffer bytes.Buffer) {
	fo, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	defer fo.Close()

	w := bufio.NewWriter(fo)
	buffer.WriteTo(w)
	w.Flush()
}
