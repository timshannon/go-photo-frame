// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"time"

	"github.com/spf13/viper"
	bh "github.com/timshannon/bolthold"
	"go.etcd.io/bbolt"
)

var store *bh.Store

type image struct {
	Key         string `boltholdKey:"Key"`
	Date        time.Time
	Provider    string `boltholdIndex:"Provider"`
	Data        []byte
	ContentType string
}

func openStore(file string) error {
	s, err := bh.Open(file, 0666, nil)
	if err != nil {
		return err
	}
	store = s
	return nil
}

func closeStore() error {
	return store.Close()
}

func getLastImageDate(provider string) (time.Time, error) {
	img := &image{}
	err := store.FindOne(img, bh.Where("Provider").Eq(provider).SortBy("Date").Reverse().Index("Provider"))
	if err == bh.ErrNotFound {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return img.Date, nil
}

func getImage(key string) (*image, error) {
	i := &image{}
	err := store.Get(key, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func getImages(query *bh.Query) ([]*image, error) {
	var images []*image
	err := store.Find(&images, query)
	if err != nil {
		return nil, err
	}
	return images, nil
}

func addImages(images []*image) error {
	return store.Bolt().Update(func(tx *bbolt.Tx) error {
		for i := range images {
			err := store.TxInsert(tx, images[i].Key, images[i])
			if err != nil {
				return err
			}
		}

		all := &bh.Query{}

		count, err := store.TxCount(tx, &image{}, all)
		if err != nil {
			return err
		}
		if count <= viper.GetInt("maxImageCount") {
			return nil
		}

		return store.TxDeleteMatching(tx, &image{},
			all.SortBy("Date").Limit(count-viper.GetInt("maxImageCount")))
	})
}
