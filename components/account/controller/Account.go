package controller

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/components/security"
	"banking-app-be/components/web"
	"banking-app-be/model/account"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	accountService "banking-app-be/components/account/service"
)

type AccountController struct {
	log            log.Logger
	AccountService *accountService.AccountService
}

func NewAccountController(accountService *accountService.AccountService, log log.Logger) *AccountController {
	return &AccountController{
		log:            log,
		AccountService: accountService,
	}
}

func (Controller *AccountController) RegisterRoutes(router *mux.Router) {

	// http://localhost:8001/api/v1/banking-app/
	accountRouter := router.PathPrefix("/account").Subrouter()
	guardedRouter := accountRouter.PathPrefix("/").Subrouter()

	//Post
	guardedRouter.HandleFunc("/bank/{bankId}", Controller.createAccount).Methods(http.MethodPost)

	//Get
	guardedRouter.HandleFunc("/", Controller.getAllUserAccounts).Methods(http.MethodGet)

	guardedRouter.Use(security.MiddlewareUser)
}

func (controller *AccountController) createAccount(w http.ResponseWriter, r *http.Request) {

	newAccount := account.Account{}
	parser := web.NewParser(r)

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}
	newAccount.UserID = userID
	newAccount.CreatedBy = userID

	bankID, err := parser.GetUUID("bankId")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid bank ID format"))
		return
	}
	newAccount.BankID = bankID

	err = controller.AccountService.CreateAccount(&newAccount)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newAccount)
}

func (controller *AccountController) getAllUserAccounts(w http.ResponseWriter, r *http.Request) {

	allAccounts := &[]account.AccountDTO{}
	var totalCount int
	query := r.URL.Query()

	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5 //default
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 //default
	}

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unauthorized", http.StatusUnauthorized))
		return
	}

	err = controller.AccountService.GetAccountsByUser(userID, allAccounts, &totalCount, limit, offset)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allAccounts)
}
