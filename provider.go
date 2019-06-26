// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"
)

type provider interface {
	name() string
	initialize(config map[string]interface{})
	getImages(since time.Time) ([]*image, error)
}

var providers []provider

func initializeProviders(pollDuration string, config map[string]interface{}) {
	poll, err := time.ParseDuration(pollDuration)
	if err != nil {
		poll = 1 * time.Hour
	}

	for k, v := range config {
		var p provider
		if k == "instagram" {
			p = &instagram{}
		}
		p.initialize(v.(map[string]interface{}))

		providers = append(providers, p)
	}

	go pollProviders(poll)
}

func pollProviders(poll time.Duration) {
	for _, p := range providers {
		dt, err := getLastImageDate(p.name())
		if err != nil {
			fmt.Printf("%s\tError getting last image from %s: %s\n", time.Now().Format(time.RFC3339),
				p.name(), err)
			continue
		}
		images, err := p.getImages(dt)
		if err != nil {
			fmt.Printf("%s\tError getting images from %s: %s\n", time.Now().Format(time.RFC3339),
				p.name(), err)
			continue
		}

		err = addImages(images)
		if err != nil {
			fmt.Printf("%s\tError inserting images from %s: %s\n", time.Now().Format(time.RFC3339),
				p.name(), err)
		}
	}

	time.Sleep(poll)
	pollProviders(poll)
}
