// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type instagram struct {
	username string
}

func (i *instagram) initialize(config map[string]interface{}) {
	if username, ok := config["username"]; ok {
		i.username = username.(string)
	}
}

func (i *instagram) name() string { return "instagram" }
func (i *instagram) getImages(since time.Time) ([]*image, error) {
	res, err := http.Get("https://instagram.com/" + i.username)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(doc.Text())
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, ok := s.Attr("src")
		if ok {
			fmt.Printf("%d, src: %s\n", i, src)
		}
	})
	return nil, nil
}
