package module

import (
	"banking-app-be/app"
	"banking-app-be/module/repository"

	"banking-app-be/components/bank/controller"
	bankService "banking-app-be/components/bank/service"
)

func registerBankRoutes(appObj *app.App, repository repository.Repository) {

	defer appObj.WG.Done()
	bankService := bankService.NewBankService(appObj.DB, repository)

	bankController := controller.NewBankController(bankService, appObj.Log)

	appObj.RegisterControllerRoutes([]app.Controller{
		bankController,
	})

}
