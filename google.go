// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

type google struct {
	urls []string
}

func (g *google) initialize(config providerConfig) error {
	if urls, ok := config.getStringSlice("urls"); ok {
		g.urls = urls
	}

	return nil
}

func (g *google) name() string { return "google" }
func (g *google) getImages(lastImage *image) ([]*image, error) {
	if len(g.urls) == 0 {
		return nil, nil
	}

	var images []*image
	return images, nil
}
