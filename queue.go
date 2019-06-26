// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math/rand"
	"sync"

	bh "github.com/timshannon/bolthold"
)

const (
	queueOrderDefault = "default"
	queueOrderRandom  = "random"
	queueOrderNewest  = "newest"
	queueOrderOldest  = "oldest"
)

type queue struct {
	sync.Mutex
	queue []string
	order collator
}

type collator interface {
	query() *bh.Query
	next(total int) int
}

func newQueue(size int, order string) *queue {
	var col collator

	switch order {
	case queueOrderRandom:
		col = &randomCollator{}
	case queueOrderNewest:
		col = &sequentialCollator{descending: true}
	case queueOrderOldest:
		col = &sequentialCollator{}
	default:
		col = &defaultCollator{}
	}
	return &queue{
		queue: make([]string, 0, size),
		order: col,
	}
}

func (q *queue) repopulate() error {
	images, err := getImages(q.order.query())
	if err != nil {
		return err
	}

	if len(images) == 0 {
		return fmt.Errorf("no images found")
	}

	q.queue = q.queue[:0]
	for i := range images {
		q.queue = append(q.queue, images[i].Key)
	}

	return nil
}

func (q *queue) next() (*image, error) {
	q.Lock()
	defer q.Unlock()

	if len(q.queue) == 0 {
		err := q.repopulate()
		if err != nil {
			return nil, err
		}
	}

	i := q.order.next(len(q.queue))
	if i == -1 {
		err := q.repopulate()
		if err != nil {
			return nil, err
		}
	}

	img, err := getImage(q.queue[i])
	if err != nil {
		return nil, err
	}

	q.queue = append(q.queue[:i], q.queue[i+1:]...)
	return img, nil
}

// collators

// defaultCollator returns images randomly weighted towards newer images
type defaultCollator struct {
	queueSize int
}

func (d *defaultCollator) query() *bh.Query {
	d.queueSize = 0
	all := bh.Query{}

	return all.SortBy("Date").Reverse()
}

func (d *defaultCollator) next(total int) int {
	d.queueSize++
	if d.queueSize > 50 {
		// every 50 images repopulate image queue
		return -1
	}
	weight := 3 // 1 in 3 chance of return an image in the newest 3rd of the image queue
	if rand.Intn(weight) == 0 {
		return rand.Intn(total)
	}
	return rand.Intn(total / 3)
}

type randomCollator struct{}

func (r *randomCollator) query() *bh.Query {
	// return all in any order
	return nil
}

func (r *randomCollator) next(total int) int {
	return rand.Intn(total)
}

type sequentialCollator struct {
	descending bool
}

func (s *sequentialCollator) query() *bh.Query {
	all := bh.Query{}
	if s.descending {
		return all.SortBy("Date").Reverse()
	}
	return all.SortBy("Date")
}

func (s *sequentialCollator) next(total int) int {
	return 0
}
