// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"time"

	"github.com/timshannon/bolthold"
)

var store *bolthold.Store

type image struct {
	Key         uint64 `boltholdKey:"Key"`
	UniqueID    string
	Date        time.Time
	Provider    string `boltholdIndex:"Provider"`
	Data        []byte
	ContentType string
}

func openStore(file string) error {
	s, err := bolthold.Open(file, 0666, nil)
	if err != nil {
		return err
	}
	store = s
	return nil
}

func closeStore() error {
	return store.Close()
}

func addImages(images []*image) error {
	for i := range images {
		err := store.Insert(bolthold.NextSequence(), images[i])
		if err != nil {
			return err
		}
	}
	return nil
}
