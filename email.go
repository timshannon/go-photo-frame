// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap/client"
)

type email struct {
	server   string
	port     string
	username string
	password string
	mailbox  string
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
	return nil
}

func (e *email) name() string { return "email" }
func (e *email) getImages(lastImage *image) ([]*image, error) {
	fmt.Println(e.server)

	var images []*image

	c, err := client.DialTLS(fmt.Sprintf("%s:%s", e.server, e.port), nil)
	if err != nil {
		return nil, err
	}

	defer c.Logout()

	if err := c.Login(e.username, e.password); err != nil {
		return nil, err
	}

	done := make(chan error, 1)

	// Select INBOX
	mbox, err := c.Select(e.mailbox, false)
	if err != nil {
		return nil, err
	}

	if mbox.Messages == 0 {
		return nil, nil
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	return images, nil
}
