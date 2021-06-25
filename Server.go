/*
 * Copyright (c) 2021.  D-Haven.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/D-Haven/fact-totem/version"
	"github.com/D-Haven/fact-totem/webapi"
	"github.com/heptiolabs/healthcheck"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const appName = "Fact-Totem"

func main() {
	err := ShowLogo(log.Writer())
	if err != nil {
		log.Fatalf("Error printing logo: %s", err)
	}

	err = version.Print(log.Writer())
	if err != nil {
		log.Fatalf("Error printing version: %s", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		// Don't kill the app, but report the error
		hostname = "--hostname lookup error: " + err.Error()
	}

	log.Printf("Server host name: %s", hostname)
	log.Printf("Configuring %s...", appName)

	config, err := GetValidatedConfig("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	ConfigureJwt(config)

	health := healthcheck.NewHandler()
	health.AddLivenessCheck("go-routinethreshold", healthcheck.GoroutineCountCheck(100))

	multiplexHandler := http.NewServeMux()
	multiplexHandler.Handle("/ready", health)
	multiplexHandler.Handle("/live", health)

	projectApi, err := webapi.NewApi(config.EventStore.Path, config.EventStore.EncryptionKey, config.EventStore.KeyDuration)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := projectApi.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	authHandler := webapi.AuthHandler(projectApi.Handle)
	multiplexHandler.Handle("/", authHandler)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    config.Server.Host + ":" + config.Server.Port,
		Handler: multiplexHandler,
	}

	go func() {
		useTLS := len(config.Server.TLS.CertFile) != 0 && len(config.Server.TLS.KeyFile) != 0

		if useTLS {
			err = server.ListenAndServeTLS(config.Server.TLS.CertFile, config.Server.TLS.KeyFile)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil {
			log.Printf("Error running service: %s", err)
		}
	}()

	log.Printf("%s started, listening on %s", appName, server.Addr)
	log.Print("K8s Readiness Check: /ready")
	log.Print("K8s Liveness Check: /live")

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Received OS Interrupt")
	case syscall.SIGTERM:
		log.Print("Received Termination Signal")
	}

	log.Printf("%s is shutting down...", appName)
	if err = server.Shutdown(context.Background()); err != nil {
		log.Printf("Shutdown error: %s", err)
	}

	log.Print("...Shutdown complete")
}
