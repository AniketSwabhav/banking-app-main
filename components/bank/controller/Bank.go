package controller

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/components/security"
	"banking-app-be/components/web"
	"banking-app-be/model/bank"
	"net/http"

	bankService "banking-app-be/components/bank/service"

	"github.com/gorilla/mux"
)

type BankController struct {
	log         log.Logger
	UserService *bankService.BankService
}

func NewBankController(userService *bankService.BankService, log log.Logger) *BankController {
	return &BankController{
		log:         log,
		UserService: userService,
	}
}

func (bankController *BankController) RegisterRoutes(router *mux.Router) {

	userRouter := router.PathPrefix("/bank").Subrouter()
	guardedRouter := userRouter.PathPrefix("/").Subrouter()
	// unguardedRouter := userRouter.PathPrefix("/").Subrouter()

	guardedRouter.HandleFunc("/register-bank", bankController.addBank).Methods(http.MethodPost)

	guardedRouter.Use(security.MiddlewareAdmin)
}

func (controller *BankController) addBank(w http.ResponseWriter, r *http.Request) {

	newBank := bank.Bank{}

	err := web.UnmarshalJSON(r, &newBank)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse requested data", http.StatusBadRequest))
		return
	}

	err = controller.UserService.CreateBank(&newBank)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newBank)
}
