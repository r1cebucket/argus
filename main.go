package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	"github.com/robfig/cron/v3"
)

type Conf struct {
	CronConf    Cron    `json:"cron" toml:"cron"`
	WatcherConf Watcher `json:"watcher" toml:"watcher"`
}

type Cron struct {
	LogLevel string     `json:"log_level" toml:"log_level"`
	Tasks    []CronTask `json:"tasks" toml:"tasks"`
}

type CronTask struct {
	Name     string   `json:"name" toml:"name"`
	Schedule string   `json:"schedule" toml:"schedule"`
	Init     []string `json:"init" toml:"init"`
	Cmd      []string `json:"cmd" toml:"cmd"`
}

type Watcher struct {
	LogLevel string        `json:"log_level" toml:"log_level"`
	Tasks    []WatcherTask `json:"tasks" toml:"tasks"`
}

type WatcherTask struct {
	Name string   `json:"name" toml:"name"`
	Path string   `json:"path" toml:"path"`
	Init []string `json:"init" toml:"init"`
	Cmd  []string `json:"cmd" toml:"cmd"`
}

func main() {
	confData, err := os.ReadFile("conf.toml")
	if err != nil {
		log.Printf("read conf fail: %s\n", err)
		return
	}

	var conf Conf

	if err := toml.Unmarshal(confData, &conf); err != nil {
		log.Printf("unmarshal conf fail: %s\n", err)
	}

	c := cornStart(conf.CronConf)
	watchers := watcherStart(conf.WatcherConf)

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt)
	<-quitChan

	c.Stop()
	for _, w := range watchers {
		w.Close()
	}

	log.Println("argus quit")
}

func cornStart(conf Cron) *cron.Cron {
	c := cron.New()

	for _, t := range conf.Tasks {
		if len(t.Init) > 0 {
			if err := execCmds(t.Init); err != nil {
				log.Printf("cron task %s run init cmd fail: %s\n", t.Name, err)
			}
		}

		c.AddFunc(t.Schedule, func() {
			if err := execCmds(t.Cmd); err != nil {
				log.Printf("cron task %s run fail: %s\n", t.Name, err)
			}
		})
	}

	c.Start()

	return c
}

func watcherStart(conf Watcher) []*fsnotify.Watcher {
	watchers := make([]*fsnotify.Watcher, 0, len(conf.Tasks))

	for _, t := range conf.Tasks {
		if len(t.Init) > 0 {
			if err := execCmds(t.Init); err != nil {
				log.Printf("watcher task %s run init cmd fail: %s\n", t.Name, err)
			}
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Printf("watcher task %s create watcher err: %s\n", t.Name, err)
		}
		err = watcher.Add(t.Path)
		if err != nil {
			log.Printf("watcher task %s add path err: %s\n", t.Name, err)
		}

		go func(t WatcherTask) {
			for {
				select {
				case _, ok := <-watcher.Events:
					if !ok {
						log.Printf("watcher task %s event not ok\n", t.Name)
						return
					}
					if err := execCmds(t.Cmd); err != nil {
						log.Printf("watcher task %s run fail: %s\n", t.Name, err)
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Printf("watcher task %s err: %s\n", t.Name, err)
				}
			}
		}(t)

		watchers = append(watchers, watcher)
	}

	return watchers
}

func execCmds(cmdStrs []string) error {
	for _, cmdStr := range cmdStrs {
		cmd := exec.Command("sh", "-c", cmdStr)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
