package module

import (
	"banking-app-be/app"
	"banking-app-be/components/account/controller"
	accountService "banking-app-be/components/account/service"
	"banking-app-be/module/repository"
)

func registerAccountRoutes(appObj *app.App, repository repository.Repository) {

	defer appObj.WG.Done()
	acountService := accountService.NewAccountService(appObj.DB, repository)

	accountController := controller.NewAccountController(acountService, appObj.Log)

	appObj.RegisterControllerRoutes([]app.Controller{
		accountController,
	})
}
