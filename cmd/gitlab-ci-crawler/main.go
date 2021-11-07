package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/deichindianer/gitlab-ci-crawler/internal/storage/neo4j"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/conf/v2"
	"github.com/deichindianer/gitlab-ci-crawler/internal/crawler"
)

var cfg crawler.Config
var neo4jcfg neo4j.Config

func main() {
	rootCtx, rootCancel := context.WithCancel(context.Background())

	shutdownChan := make(chan os.Signal, 2)
	signal.Notify(shutdownChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		rootCancel()
	}()

	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return
		}
		log.Fatalf("parsing config: %s", err)
	}

	nHelp, err := conf.Parse("", &neo4jcfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(nHelp)
			return
		}
		log.Fatalf("parsing config: %s", err)
	}

	s, err := neo4j.New(neo4j.Config{
		Host:     neo4jcfg.Host,
		Username: neo4jcfg.Username,
		Password: neo4jcfg.Password,
		Realm:    "",
	})

	c, err := crawler.New(cfg, s)
	if err != nil {
		log.Fatalf("failed to setup crawler: %s\n", err)
	}

	if err := c.Crawl(rootCtx); err != nil {
		log.Fatalf("failed to gather project data: %s", err)
	}
}
