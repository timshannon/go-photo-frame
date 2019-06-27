// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"time"

	bh "github.com/timshannon/bolthold"
)

func startServer(port string, imageDuration time.Duration, q *queue) error {
	var mainTemplate = template.Must(template.New("").Parse(html))
	var loadingTemplate = template.Must(template.New("").Parse(loading))

	templateData := struct {
		Duration int64
	}{
		Duration: int64(imageDuration / time.Millisecond),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		count, err := store.Count(&image{}, &bh.Query{})
		if err != nil {
			log.Printf("Error getting image count: %s", err)
		}
		if count == 0 {
			loadingTemplate.Execute(w, templateData)
			return
		}
		mainTemplate.Execute(w, templateData)
	})

	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		img, err := q.next()
		if err != nil {
			log.Printf("Error getting image: %s\n", err)
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", img.ContentType)

		http.ServeContent(w, r, img.Key, time.Time{}, bytes.NewReader(img.Data))
	})

	log.Printf("Go Photo Frame is running on port %s\n", port)
	return http.ListenAndServe(":"+port, nil)
}
