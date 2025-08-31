package controller

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	passbookService "banking-app-be/components/passbook/service"
	"banking-app-be/components/security"
	"banking-app-be/components/web"
	"banking-app-be/model/passbook"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PassbookController struct {
	log             log.Logger
	PassbookService *passbookService.PassbookService
}

func NewPassbookController(passbookService *passbookService.PassbookService, log log.Logger) *PassbookController {
	return &PassbookController{
		log:             log,
		PassbookService: passbookService,
	}
}

func (Controller *PassbookController) RegisterRoutes(router *mux.Router) {

	// http://localhost:8001/api/v1/banking-app/
	accountRouter := router.PathPrefix("/passbook").Subrouter()
	guardedRouter := accountRouter.PathPrefix("/").Subrouter()

	//Get
	guardedRouter.HandleFunc("/", Controller.getPassbookByAccountNo).Methods(http.MethodPost)

	guardedRouter.Use(security.MiddlewareUser)
}

func (controller *PassbookController) getPassbookByAccountNo(w http.ResponseWriter, r *http.Request) {

	var requestData struct {
		AccountNo string `json:"accountNo"`
	}

	err := web.UnmarshalJSON(r, &requestData)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("Unable to parse requested data", http.StatusBadRequest))
		return
	}

	passbook := []passbook.Transaction{}

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

	err = controller.PassbookService.GetPassbookByAccountNo(&passbook, userID, &requestData.AccountNo, &totalCount, limit, offset)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, passbook)

}
