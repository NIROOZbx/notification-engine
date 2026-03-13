package main

import (
	"log"

	"github.com/NIROOZbx/notification-engine/services/backend/app"
	"github.com/NIROOZbx/notification-engine/services/backend/config"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	app,err := app.StartApp(cfg)
	if err != nil {
		log.Fatalf("cannot start app: %v", err)
	}

	if err := Run(app, cfg.Server.HTTPAddr); err != nil {
		log.Fatalf("cannot run program: %v", err)
	}
	

}
