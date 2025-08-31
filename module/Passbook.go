package module

import (
	"banking-app-be/app"
	"banking-app-be/components/passbook/controller"
	passbookService "banking-app-be/components/passbook/service"
	"banking-app-be/module/repository"
)

func registerPassbookRoutes(appObj *app.App, repository repository.Repository) {

	defer appObj.WG.Done()
	acountService := passbookService.NewPassbookService(appObj.DB, repository)

	accountController := controller.NewPassbookController(acountService, appObj.Log)

	appObj.RegisterControllerRoutes([]app.Controller{
		accountController,
	})
}
