// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	bh "github.com/timshannon/bolthold"
)

type email struct {
	server   string
	port     string
	username string
	password string
	mailbox  string
	from     []string
	to       string
}

func (e *email) initialize(config providerConfig) error {
	server, serverOk := config.getString("server")
	port, portOk := config.getString("port")
	username, usernameOk := config.getString("username")
	password, passwordOk := config.getString("password")
	mailbox, mailboxOk := config.getString("mailbox")
	if !serverOk || !portOk || !usernameOk || !passwordOk || !mailboxOk {
		return fmt.Errorf("Invalid email config")
	}
	e.server = server
	e.port = port
	e.username = username
	e.password = password
	e.mailbox = mailbox
	e.from, _ = config.getStringSlice("from")
	e.to, _ = config.getString("to")
	return nil
}

func (e *email) name() string { return "email" }
func (e *email) getImages(lastImage *image) ([]*image, error) {
	var images []*image

	c, err := client.DialTLS(fmt.Sprintf("%s:%s", e.server, e.port), nil)
	if err != nil {
		return nil, err
	}

	defer c.Logout()

	if err := c.Login(e.username, e.password); err != nil {
		return nil, err
	}

	mbox, err := c.Select(e.mailbox, false)
	if err != nil {
		return nil, err
	}

	if mbox.Messages == 0 {
		return nil, nil
	}

	for i := mbox.Messages; i != 0; i-- {
		imgs, err := e.getImagesFromEmail(c, i)
		if err != nil {
			return nil, err
		}

		images = append(images, imgs...)
	}

	return images, nil
}

func (e *email) getImagesFromEmail(client *client.Client, sequence uint32) ([]*image, error) {
	var images []*image
	seqset := new(imap.SeqSet)
	seqset.AddNum(sequence)

	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	var err error
	go func() {
		err = client.Fetch(seqset, items, messages)
	}()

	msg := <-messages
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, fmt.Errorf("Server didn't return message")
	}

	r := msg.GetBody(section)
	if r == nil {
		return nil, fmt.Errorf("Server didn't return message body")
	}

	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, err
	}

	imgDate := time.Now()

	// Print some info about the message
	header := mr.Header
	if date, err := header.Date(); err == nil {
		imgDate = date
	}

	if len(e.from) > 0 {
		// only add images from emails in from whitelist
		if from, err := header.AddressList("From"); err == nil {
			found := false
			for _, good := range e.from {
				if e.hasAddress(from, good) {
					found = true
					break
				}
			}
			if !found {
				return nil, nil
			}
		}
	}
	if e.to != "" {
		if to, err := header.AddressList("To"); err == nil {
			if !e.hasAddress(to, e.to) {
				return nil, nil
			}
		}
	}

	// Process each message's part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch h := p.Header.(type) {
		case *mail.AttachmentHeader:
			ctype := p.Header.Get("Content-Type")
			if strings.Contains(ctype, "image") {
				split := strings.Split(ctype, ";")
				if len(split) < 1 {
					continue
				}
				ctype = split[0]
				filename, _ := h.Filename()

				key := fmt.Sprintf("%s.%d.%s", e.mailbox, msg.Uid, filename)
				_, err := getImage(key)
				if err == nil {
					// image already added
					continue
				}
				if err != bh.ErrNotFound {
					return nil, err
				}

				body, err := ioutil.ReadAll(p.Body)
				if err != nil {
					return nil, err
				}

				images = append(images, &image{
					Key:         key,
					Date:        imgDate,
					Data:        body,
					Provider:    e.name(),
					ContentType: ctype,
				})
			}
		}
	}
	return images, nil
}

func (e *email) hasAddress(addresses []*mail.Address, address string) bool {
	for _, addr := range addresses {
		a := addr.String()
		a = a[strings.Index(a, "<")+1 : strings.Index(a, ">")]
		if strings.ToLower(a) == strings.ToLower(address) {
			return true
		}
	}
	return false
}
