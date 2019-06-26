// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func startServer(port string, q *queue) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(html))
	})

	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		img, err := q.next()
		if err != nil {
			fmt.Printf("%s\tError getting image: %s\n", time.Now(), err)
		}

		w.Header().Set("Content-Type", img.ContentType)

		http.ServeContent(w, r, img.Key, time.Time{}, bytes.NewReader(img.Data))
	})

	fmt.Printf("Go Photo Frame is running on port %s\n", port)
	return http.ListenAndServe(":"+port, nil)
}
