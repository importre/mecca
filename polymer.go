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
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

// PolymerCrawler gathers polymer repositories that exist on Github and
// dumps to json file.
type PolymerCrawler struct {
	client *github.Client
	repos  []PolymerRepository
}

// PolymerRepository represents an polymer repository.
// It embeds github.Repository basically.
type PolymerRepository struct {
	github.Repository
}

// NewPolymerCrawler returns a new PolymerCrawler.
func NewPolymerCrawler(accessToken string) *PolymerCrawler {
	if accessToken == "" {
		log.Println("Github AccessToken is empty.")
		return &PolymerCrawler{
			client: github.NewClient(nil),
		}
	}

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: GITHUB_TOKEN},
	}

	return &PolymerCrawler{
		client: github.NewClient(t.Client()),
	}
}

func (self *PolymerCrawler) Dump() {
	outDir := "data/polymer/"
	os.MkdirAll(outDir, 0755)
	DumpToJson(outDir+"repos.json", self.repos)
}

// FindRepos finds repositories.
func (self *PolymerCrawler) FindRepos(stars int) {
	capacity := 100
	self.repos = make([]PolymerRepository, 0, capacity)

	page := 1
	maxPage := math.MaxInt32

	query := fmt.Sprintf("polymer fork:true stars:>=%v", stars)
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
			if !self.IsPolymer(&repo) {
				continue
			}

			RemoveUnnecessaryFeatures(&repo)

			polymerRepo := PolymerRepository{
				repo,
			}

			self.repos = append(self.repos, polymerRepo)

			self.downloadAvatarImage(&repo)
			log.Println("polymer:", *repo.FullName)
			time.Sleep(time.Millisecond * 500)
		}

		page++
	}
}

func (self *PolymerCrawler) downloadAvatarImage(repo *github.Repository) {
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

// IsPolymer returns true if the repo has polymer project otherwise false.
func (self *PolymerCrawler) IsPolymer(repo *github.Repository) bool {
	opts := &github.SearchOptions{}

	query := fmt.Sprintf("\"platform.js\" extension:html language:html repo:%v", *repo.FullName)
	result, response, err := self.client.Search.Code(query, opts)
	Wait(response)

	if err != nil {
		log.Fatal("IsPolymer:", err)
	}

	return *result.Total > 0
}
