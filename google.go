// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	bh "github.com/timshannon/bolthold"
)

type google struct {
	urls []string
}

func (g *google) initialize(config providerConfig) error {
	if urls, ok := config.getStringSlice("urls"); ok {
		g.urls = urls
	}

	return nil
}

func (g *google) name() string { return "google-photos" }
func (g *google) getImages(lastImage *image) ([]*image, error) {
	if len(g.urls) == 0 {
		return nil, nil
	}

	var images []*image

	for _, album := range g.urls {
		imgCount := 0

		baseCtx := context.Background()
		// baseCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://0.0.0.0:9222/devtools/browser/1c166044-6769-4285-84a3-6827acc42af8")
		ctx, cancel := chromedp.NewContext(baseCtx)

		defer cancel()

		// create a timeout
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		var imgNodes []*cdp.Node
		err := chromedp.Run(ctx,
			chromedp.Navigate(album),
			chromedp.WaitVisible(`a[aria-label*="Photo"]`),
			chromedp.Nodes(`a[aria-label*="Photo"] > div[style*="background-image"]`, &imgNodes),
		)
		if err != nil {
			return nil, err
		}
		for _, n := range imgNodes {
			img, err := g.getImage(n)
			if err != nil {
				return nil, err
			}
			if img == nil {
				continue
			}
			images = append(images, img)
			imgCount++
			if imgCount >= maxImagesPerPoll {
				break
			}
		}

	}

	return images, nil
}

func (g *google) imageDate(node *cdp.Node) time.Time {
	if node.Parent == nil {
		return time.Now()
	}

	split := strings.Split(node.Parent.AttributeValue("aria-label"), " - ")
	if len(split) != 3 {
		return time.Now()
	}

	t, err := time.Parse("Jan 02, 2006, 3:05:05 PM", split[2])
	if err != nil {
		return time.Now()
	}

	return t
}

func (g *google) imageURL(node *cdp.Node) string {
	attr := node.AttributeValue("style")
	index := `background-image: url("`
	part := attr[strings.Index(attr, index)+len(index):]
	u := part[:strings.Index(part, `"`)]
	// remove size part of url
	s := strings.Split(u, "=")
	if len(s) < 1 {
		return u
	}
	return s[0]
}

func (g *google) getImage(imageNode *cdp.Node) (*image, error) {
	var img *image

	imgURL := g.imageURL(imageNode)

	_, err := getImage(imgURL)
	if err == nil {
		// image already added
		return nil, nil
	}
	if err != bh.ErrNotFound {
		return nil, err
	}

	imageDate := g.imageDate(imageNode)

	// add arbitrarily large image size, google returns the largest it has
	resp, err := http.Get(imgURL + "=w4048-h4048-no")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	img = &image{
		Key:         imgURL,
		Date:        imageDate,
		Data:        body,
		Provider:    g.name(),
		ContentType: resp.Header.Get("Content-Type"),
	}

	return img, nil
}
