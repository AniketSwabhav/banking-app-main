package module

import (
	"banking-app-be/app"
	"banking-app-be/module/repository"
)

func RegisterModuleRoutes(app *app.App, repository repository.Repository) {
	log := app.Log
	log.Print("============Registering-Module-Routes==============")

	app.WG.Add(4)
	registerUserRoutes(app, repository)
	registerBankRoutes(app, repository)
	registerAccountRoutes(app, repository)
	app.WG.Done()
}
