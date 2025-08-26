package main

import (
	"banking-app-be/app"
	"banking-app-be/components/config"
	"banking-app-be/components/log"
	"banking-app-be/docs"
	"banking-app-be/module"
	"banking-app-be/module/repository"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var environment = "local"

func main() {
	env := config.Environment(environment)

	log := log.GetLogger()
	log.Info("Starting main in ", env, ".")

	config.InitializeGlobalConfig(env)

	if env == config.Local {
		docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", config.PORT.GetStringValue())
	}

	db := app.NewDBConnection(log)
	if db == nil {
		log.Fatalf("Db connection failed.")
	}
	defer func() {
		db.Close()
		log.Info("Db closed")
	}()

	var wg sync.WaitGroup
	var repository = repository.NewGormRepository()

	app := app.NewApp("Banking App", db, log, &wg, repository)
	app.Init()

	module.RegisterModuleRoutes(app, repository)

	go func() {
		err := app.StartServer()
		if err != nil {
			app.Log.Print("Error in starting Server")
			stopApp(app)
		}
	}()

	app.Log.Print("Server Started")

	module.Configure(app)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	stopApp(app)
}

func stopApp(app *app.App) {
	app.Stop()
	app.Log.Print("App stopped.")
	os.Exit(0)
}
