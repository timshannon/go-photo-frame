// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
)

func startServer(port string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(html))
	})

	http.HandleFunc("/image/current", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		// TODO: current image
	})

	http.HandleFunc("/image/next", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		// TODO: next image
	})

	fmt.Printf("Go Photo Frame is running on port %s\n", port)
	return http.ListenAndServe(":"+port, nil)
}
