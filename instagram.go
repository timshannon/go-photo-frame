// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly"
	bh "github.com/timshannon/bolthold"
)

// from http://go-colly.org/docs/examples/instagram/
// found in https://www.instagram.com/static/bundles/en_US_Commons.js/68e7390c5938.js
// included from profile page
const instagramQueryID = "42323d64886122307be10013ad2dcc45"

// "id": user id, "after": end cursor
const instagramNextPageURL string = `https://www.instagram.com/graphql/query/?query_hash=%s&variables=%s`
const instagramNextPagePayload string = `{"id":"%s","first":12,"after":"%s"}`

// hardcoded query hash id, not sure how long this will work
const instagramRequestID = "f2405b236d85e8296cf30347c9f08c2a"

type instagramPageInfo struct {
	EndCursor string `json:"end_cursor"`
	NextPage  bool   `json:"has_next_page"`
}

type instagramNode struct {
	ID           string `json:"id"`
	ImageURL     string `json:"display_url"`
	ShortCode    string `json:"shortcode"`
	ThumbnailURL string `json:"thumbnail_src"`
	IsVideo      bool   `json:"is_video"`
	Date         int    `json:"date"`
	Dimensions   struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"dimensions"`
}

type instagramPageData struct {
	Rhxgis    string `json:"rhx_gis"`
	EntryData struct {
		ProfilePage []struct {
			Graphql struct {
				User struct {
					ID    string `json:"id"`
					Media struct {
						Edges []struct {
							Node instagramNode `json:"node"`
						} `json:"edges"`
						PageInfo instagramPageInfo `json:"page_info"`
					} `json:"edge_owner_to_timeline_media"`
				} `json:"user"`
			} `json:"graphql"`
		} `json:"ProfilePage"`
		PostPage []struct {
			Graphql struct {
				ShortCodeMedia struct {
					instagramNode
					EdgeSidecarToChildren struct {
						Edges []struct {
							Node instagramNode `json:"node"`
						} `json:"edges"`
					} `json:"edge_sidecar_to_children"`
				} `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
}

type instagramNextPageData struct {
	Data struct {
		User struct {
			Container struct {
				PageInfo instagramPageInfo `json:"page_info"`
				Edges    []struct {
					Node instagramNode
				} `json:"edges"`
			} `json:"edge_owner_to_timeline_media"`
		}
	} `json:"data"`
}

type instagram struct {
	accounts []string
}

func (i *instagram) initialize(config providerConfig) error {
	if accounts, ok := config.getStringSlice("accounts"); ok {
		i.accounts = accounts
	}

	return nil
}

func (i *instagram) name() string { return "instagram" }
func (i *instagram) getImages(lastImage *image) ([]*image, error) {
	if len(i.accounts) == 0 {
		return nil, nil
	}

	var images []*image
	var fatalErr error

	for _, account := range i.accounts {
		imgCount := 0
		done := false
		var actualUserID string

		c := colly.NewCollector(colly.UserAgent(userAgent))

		c.OnRequest(func(r *colly.Request) {
			if done {
				return
			}
			r.Headers.Set("X-Requested-With", "XMLHttpRequest")
			r.Headers.Set("Referrer", "https://www.instagram.com/"+account)
			if r.Ctx.Get("gis") != "" {
				gis := fmt.Sprintf("%s:%s", r.Ctx.Get("gis"), r.Ctx.Get("variables"))
				h := md5.New()
				h.Write([]byte(gis))
				gisHash := fmt.Sprintf("%x", h.Sum(nil))
				r.Headers.Set("X-Instagram-GIS", gisHash)
			}
		})

		c.OnHTML("html", func(e *colly.HTMLElement) {
			if done {
				return
			}
			d := c.Clone()
			// d.OnResponse(func(r *colly.Response) {
			// 	// idStart := bytes.Index(r.Body, []byte(`:n},queryId:"`))
			// 	// requestID = string(r.Body[idStart+13 : idStart+45])
			// })
			requestIDURL := e.Request.AbsoluteURL(e.ChildAttr(`link[as="script"]`, "href"))
			d.Visit(requestIDURL)

			dat := e.ChildText("body > script:first-of-type")
			jsonData := dat[strings.Index(dat, "{") : len(dat)-1]
			data := &instagramPageData{}
			fatalErr = json.Unmarshal([]byte(jsonData), data)
			if fatalErr != nil {
				done = true
				return
			}

			if len(data.EntryData.ProfilePage) > 0 {
				page := data.EntryData.ProfilePage[0]
				actualUserID = page.Graphql.User.ID
				for _, obj := range page.Graphql.User.Media.Edges {
					// skip videos
					if obj.Node.IsVideo {
						continue
					}
					c.Visit(fmt.Sprintf("https://www.instagram.com/p/%s", obj.Node.ShortCode))
				}
				nextPageVars := fmt.Sprintf(instagramNextPagePayload, actualUserID,
					page.Graphql.User.Media.PageInfo.EndCursor)
				e.Request.Ctx.Put("variables", nextPageVars)
				if page.Graphql.User.Media.PageInfo.NextPage {

					u := fmt.Sprintf(
						instagramNextPageURL,
						instagramRequestID,
						url.QueryEscape(nextPageVars),
					)
					e.Request.Ctx.Put("gis", data.Rhxgis)
					e.Request.Visit(u)
				}
			} else if len(data.EntryData.PostPage) > 0 {
				page := data.EntryData.PostPage[0]
				if len(page.Graphql.ShortCodeMedia.EdgeSidecarToChildren.Edges) == 0 {
					c.Visit(page.Graphql.ShortCodeMedia.ImageURL)
				} else {
					for _, obj := range page.Graphql.ShortCodeMedia.EdgeSidecarToChildren.Edges {
						// skip videos
						if obj.Node.IsVideo {
							continue
						}
						c.Visit(obj.Node.ImageURL)
					}
				}
			}
		})

		c.OnError(func(r *colly.Response, e error) {
			log.Printf("error for %s: %s %s %s", account, e, r.Request.URL, string(r.Body))
		})

		c.OnResponse(func(r *colly.Response) {
			if done {
				return
			}
			if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {
				if lastImage != nil && lastImage.Key == r.FileName() {
					done = true
					return
				}

				_, err := getImage(r.FileName())
				if err == nil {
					// image already added
					return
				}
				if err != bh.ErrNotFound {
					fatalErr = err
					done = true
					return
				}

				dt, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", r.Headers.Get("Last-Modified"))
				if err != nil {
					dt = time.Now()
				}
				images = append(images, &image{
					Key:         r.FileName(),
					Date:        dt,
					Data:        r.Body,
					Provider:    i.name(),
					ContentType: r.Headers.Get("Content-Type"),
				})
				imgCount++
				if imgCount >= maxImagesPerPoll {
					done = true
				}
				return
			}

			if strings.Index(r.Headers.Get("Content-Type"), "json") == -1 {
				return
			}

			data := &instagramNextPageData{}
			fatalErr = json.Unmarshal(r.Body, data)
			if fatalErr != nil {
				return
			}

			for _, obj := range data.Data.User.Container.Edges {
				// skip videos
				if obj.Node.IsVideo {
					continue
				}
				// c.Visit(obj.Node.ImageURL)
				c.Visit(fmt.Sprintf("https://www.instagram.com/p/%s", obj.Node.ShortCode))
			}
			if data.Data.User.Container.PageInfo.NextPage {
				if done {
					return
				}
				nextPageVars := fmt.Sprintf(instagramNextPagePayload, actualUserID,
					data.Data.User.Container.PageInfo.EndCursor)
				r.Request.Ctx.Put("variables", nextPageVars)
				u := fmt.Sprintf(
					instagramNextPageURL,
					instagramRequestID,
					url.QueryEscape(nextPageVars),
				)
				r.Request.Visit(u)
			}
		})

		c.Visit("https://instagram.com/" + account)
		if fatalErr != nil {
			return nil, fatalErr
		}
	}

	return images, nil
}
