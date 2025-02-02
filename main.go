package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/pelletier/go-toml/v2"
	"github.com/robfig/cron/v3"
)

type Conf struct {
	CronConf Cron `json:"cron" toml:"cron"`
}

type Cron struct {
	LogLevel string `json:"log_level" toml:"log_level"`
	Tasks    []Task `json:"tasks" toml:"tasks"`
}

type Task struct {
	Name     string `json:"name" toml:"name"`
	Schedule string `json:"schedule" toml:"schedule"`
	Init     string `json:"init" toml:"init"`
	Cmd      string `json:"cmd" toml:"cmd"`
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

	c := cron.New()
	for _, t := range conf.CronConf.Tasks {
		if len(t.Init) > 0 {
			init := exec.Command("sh", "-c", t.Init)
			if err := init.Run(); err != nil {
				log.Printf("task %s run init cmd fail: %s\n", t.Name, err)
			}
		}

		c.AddFunc(t.Schedule, func() {
			cmd := exec.Command("sh", "-c", t.Cmd)
			if err := cmd.Run(); err != nil {
				log.Printf("task %s run fail: %s\n", t.Name, err)
			}
		})
	}

	c.Start()

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, os.Kill)
	<-quitChan

	c.Stop()
	log.Println("zeus quit")
}
