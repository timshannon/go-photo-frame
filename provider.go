// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"log"
	"time"
)

const maxImagesPerPoll = 50

type providerConfig map[string]interface{}

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
		if k == "instagram" {
			p = &instagram{}
		}
		err = p.initialize(v.(map[string]interface{}))
		if err != nil {
			log.Printf("Error initializing provider %s", p.name())
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
