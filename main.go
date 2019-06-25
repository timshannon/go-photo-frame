// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/viper"
)

func init() {
	go func() {
		//Capture program shutdown, to make sure everything shuts down nicely
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for sig := range c {
			if sig == os.Interrupt {
				fmt.Println("Go Photo Frame is shutting down")
				closeStore()
				os.Exit(0)
			}
		}
	}()
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/go-photo-frame/")
	viper.AddConfigPath("$HOME/.config/go-photo-frame/")
	viper.AddConfigPath(".")

	viper.SetDefault("port", "8080")
	viper.SetDefault("imageDuration", "3s")
	viper.SetDefault("maxImageCount", 1000)
	viper.SetDefault("imagePollDuration", "1h")
	viper.SetDefault("dataFile", "./images.db")
	viper.SetDefault("imageOrder", "default")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Fatal error loading config file: %s \n", err)
		os.Exit(1)
	}

	err = openStore(viper.GetString("dataFile"))
	if err != nil {
		fmt.Printf("Error opening data file: %s \n", err)
		os.Exit(1)

	}

	err = initializeProviders() // TODO: passin provider config
	if err != nil {
		fmt.Printf("Error initializing image providers: %s \n", err)
		os.Exit(1)
	}

	err = startServer(viper.GetString("port")) // TODO passing imageDuration?
	if err != nil {
		fmt.Printf("Error starting server: %s \n", err)
		os.Exit(1)
	}
}
