// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
)

// DumpToJson writes a json file.
func DumpToJson(path string, repos interface{}) {
	b, err := json.Marshal(repos)
	if err != nil {
		log.Fatal("DumpToJson:", err)
	}

	var buffer bytes.Buffer
	json.Indent(&buffer, b, "", "  ")
	WriteToFile(path, buffer)
}

func Wait(response *github.Response) {
	if response != nil && response.Remaining <= REMAINING_THRESHOLD {
		gap := time.Duration(response.Reset.Local().Unix() - time.Now().Unix())
		sleep := gap * time.Second
		if sleep < 0 {
			sleep = -sleep
		}

		time.Sleep(sleep)
	}
}

func RemoveUnnecessaryFeatures(repo *github.Repository) {
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

func WriteToFile(name string, buffer bytes.Buffer) {
	fo, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	defer fo.Close()

	w := bufio.NewWriter(fo)
	buffer.WriteTo(w)
	w.Flush()
}
