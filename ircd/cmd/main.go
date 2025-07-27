// SPDX-License-Identifier: BSD-2-Clause
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	cfg "github.com/0xilly/aircd/ircd/internal/config"
	"github.com/0xilly/aircd/ircd/internal/core"
	_ "github.com/0xilly/aircd/ircd/internal/core/command"
	"github.com/0xilly/aircd/ircd/internal/transport"
)

func main() {
	cfgPath := flag.String("config", "aircd.toml", "path to config file")
	flag.Parse()

	conf, created, err := cfg.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if created {
		log.Printf("wrote default config to %s", *cfgPath)
	}

	srv := core.NewServer(conf)

	// listeners...
	if conf.Server.ListenTCP != "" {
		go func() {
			if err := transport.ListenTCP(conf.Server.ListenTCP, srv, conf.Limits.MaxLineBytes); err != nil {
				log.Printf("tcp listener error: %v", err)
			}
		}()
	}
	if conf.Server.ListenWS != "" {
		go func() {
			if err := transport.ListenWS(conf.Server.ListenWS, conf.Server.WSPath, srv,
				conf.Limits.MaxLineBytes, conf.Server.TLS.Enable, conf.Server.TLS.CertFile, conf.Server.TLS.KeyFile); err != nil {
				log.Printf("ws listener error: %v", err)
			}
		}()
	}

	// ---- MOTD reload on SIGHUP / SIGUSR1 ----
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGUSR1)
	go func() {
		for range sigc {
			if err := srv.ReloadMOTD(); err != nil {
				log.Printf("reload motd: %v", err)
			} else {
				log.Printf("MOTD reloaded")
			}
		}
	}()

	select {}
}
