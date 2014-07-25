// Copyright 2014 Jaewe Heo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"time"
)

type MetaData struct {
	LastModified string `json:"last_modified"`
}

func NewMetaData() *MetaData {
  const layout = "Jan 2, 2006 at 3:04pm (MST)"
	return &MetaData{
		LastModified: time.Now().UTC().Format(layout),
	}
}

func (self *MetaData) Dump() {
	outDir := "data/"
	os.MkdirAll(outDir, 0755)
	DumpToJson(outDir+"metadata.json", self)
}
