// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"time"
)

type provider interface {
	initialize(config map[string]interface{})
	getImages(since time.Time) []*image
}

func initializeProviders() error {
	// TODO
	return nil
}
