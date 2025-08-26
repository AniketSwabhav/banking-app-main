package module

import (
	"banking-app-be/app"
	"banking-app-be/components/user/controller"
	userService "banking-app-be/components/user/service"
	"banking-app-be/module/repository"
)

func registerUserRoutes(appObj *app.App, repository repository.Repository) {

	defer appObj.WG.Done()
	userService := userService.NewUserService(appObj.DB, repository)

	userController := controller.NewUserController(userService, appObj.Log)

	appObj.RegisterControllerRoutes([]app.Controller{
		userController,
	})
}
