package controller

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/components/security"
	"banking-app-be/components/web"
	"banking-app-be/model/bank"
	"net/http"
	"strconv"

	bankService "banking-app-be/components/bank/service"

	"github.com/gorilla/mux"
)

type BankController struct {
	log         log.Logger
	BankService *bankService.BankService
}

func NewBankController(userService *bankService.BankService, log log.Logger) *BankController {
	return &BankController{
		log:         log,
		BankService: userService,
	}
}

func (Controller *BankController) RegisterRoutes(router *mux.Router) {

	bankRouter := router.PathPrefix("/bank").Subrouter()
	guardedRouter := bankRouter.PathPrefix("/").Subrouter()
	// unguardedRouter := userRouter.PathPrefix("/").Subrouter()

	//Post
	guardedRouter.HandleFunc("/register-bank", Controller.addBank).Methods(http.MethodPost)

	//Get
	guardedRouter.HandleFunc("/", Controller.getAllBanks).Methods(http.MethodGet)
	guardedRouter.HandleFunc("/{id}", Controller.getBankById).Methods(http.MethodGet)

	//Update
	bankRouter.HandleFunc("/{id}", Controller.updateBankById).Methods(http.MethodPut)

	//Delete
	bankRouter.HandleFunc("/{id}", Controller.deleteBankById).Methods(http.MethodDelete)

	guardedRouter.Use(security.MiddlewareAdmin)
}

func (controller *BankController) addBank(w http.ResponseWriter, r *http.Request) {
	newBank := bank.Bank{}
	if err := web.UnmarshalJSON(r, &newBank); err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse request data", http.StatusBadRequest))
		return
	}

	var err error
	newBank.CreatedBy, err = security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}

	if err := controller.BankService.CreateBank(&newBank); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newBank)
}

func (controller *BankController) getAllBanks(w http.ResponseWriter, r *http.Request) {
	allBanks := &[]bank.BankDTO{}
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

	err = controller.BankService.GetAllBanks(allBanks, &totalCount, limit, offset)
	if err != nil {
		controller.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}
	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allBanks)
}

func (controller *BankController) getBankById(w http.ResponseWriter, r *http.Request) {

	var targetBank = &bank.BankDTO{}

	parser := web.NewParser(r)

	bankIdFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid bank ID format"))
		return
	}

	targetBank.ID = bankIdFromURL

	err = controller.BankService.GetBankByID(targetBank)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, targetBank)
}

func (controller *BankController) updateBankById(w http.ResponseWriter, r *http.Request) {
	updatedBank := bank.Bank{}
	parser := web.NewParser(r)

	if err := web.UnmarshalJSON(r, &updatedBank); err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse request data", http.StatusBadRequest))
		return
	}

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}

	updatedBank.UpdatedBy = userID

	bankIdFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid bank ID format"))
		return
	}

	updatedBank.ID = bankIdFromURL

	if err := controller.BankService.UpdateBank(&updatedBank); err != nil {
		web.RespondError(w, err)
		return
	}

	// updatedBankDTO := bank.BankDTO{}
	// err = controller.BankService.GetBankByID(&updatedBankDTO)
	// if err != nil {
	// 	web.RespondError(w, err)
	// 	return
	// }

	web.RespondJSON(w, http.StatusOK, updatedBank)
}

func (controller *BankController) deleteBankById(w http.ResponseWriter, r *http.Request) {
	bankToDelete := bank.Bank{}
	parser := web.NewParser(r)

	userID, err := security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}
	bankToDelete.DeletedBy = userID

	bankIdFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid bank ID format"))
		return
	}
	bankToDelete.ID = bankIdFromURL

	if err := controller.BankService.DeleteBank(&bankToDelete); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "Bank deleted successfully"})
}
