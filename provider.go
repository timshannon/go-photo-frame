// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"
)

const maxImagesPerPoll = 50
const userAgent = "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"

type providerConfig map[string]interface{}

// Provider is an interface for an image provider
type provider interface {
	name() string
	initialize(config providerConfig) error
	getImages(lastImage *image) ([]*image, error)
}

var providers []provider

func initializeProviders(pollDuration string, config providerConfig) {
	poll, err := time.ParseDuration(pollDuration)
	if err != nil {
		poll = 1 * time.Hour
	}

	for k, v := range config {
		var p provider
		switch k {
		case "instagram":
			p = &instagram{}
		case "google-photos":
			p = &google{}
		default:
			log.Printf("Invalid provider name: %s", k)
			continue
		}
		err = p.initialize(v.(map[string]interface{}))
		if err != nil {
			log.Printf("Error initializing provider %s: %s", p.name(), err)
		}

		providers = append(providers, p)
	}

	go pollProviders(poll)
}

func pollProviders(poll time.Duration) {
	for _, p := range providers {
		last, err := getLastImage(p.name())
		if err != nil {
			log.Printf("Error getting last image from %s: %s\n", p.name(), err)
			continue
		}
		images, err := p.getImages(last)
		if err != nil {
			log.Printf("Error getting images from %s: %s\n", p.name(), err)
			continue
		}

		err = addImages(images)
		if err != nil {
			log.Printf("Error inserting images from %s: %s\n", p.name(), err)
		}
	}

	time.Sleep(poll)
	pollProviders(poll)
}

func (c providerConfig) getString(field string) (string, bool) {
	if val, ok := c[field]; ok {
		val, ok := val.(string)
		return val, ok
	}
	return "", false
}

func (c providerConfig) getStringSlice(field string) ([]string, bool) {
	if val, ok := c[field]; ok {
		val, ok := val.([]interface{})
		s := make([]string, len(val))
		for i := range s {
			str, ok := val[i].(string)
			if !ok {
				return nil, false
			}
			s[i] = str
		}

		return s, ok
	}
	return nil, false
}
