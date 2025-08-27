package controller

import (
	"banking-app-be/components/errors"
	"banking-app-be/components/log"
	"banking-app-be/components/security"
	"banking-app-be/components/web"
	"banking-app-be/model/credential"
	"banking-app-be/model/user"
	"net/http"
	"strconv"

	userService "banking-app-be/components/user/service"

	"github.com/gorilla/mux"
)

type UserController struct {
	log         log.Logger
	UserService *userService.UserService
}

func NewUserController(userService *userService.UserService, log log.Logger) *UserController {
	return &UserController{
		log:         log,
		UserService: userService,
	}
}

func (userController *UserController) RegisterRoutes(router *mux.Router) {

	userRouter := router.PathPrefix("/user").Subrouter()
	guardedRouter := userRouter.PathPrefix("/").Subrouter()
	unguardedRouter := userRouter.PathPrefix("/").Subrouter()

	//Post
	unguardedRouter.HandleFunc("/login", userController.login).Methods(http.MethodPost)
	unguardedRouter.HandleFunc("/register-admin", userController.registerAdmin).Methods(http.MethodPost)
	guardedRouter.HandleFunc("/register-user", userController.registerUser).Methods(http.MethodPost)

	// Get
	guardedRouter.HandleFunc("/", userController.getAllUsers).Methods(http.MethodGet)
	guardedRouter.HandleFunc("/{id}", userController.getUserById).Methods(http.MethodGet)

	//Update
	guardedRouter.HandleFunc("/{id}", userController.updateUserById).Methods(http.MethodPut)

	// Delete
	guardedRouter.HandleFunc("/{id}", userController.deleteUserById).Methods(http.MethodDelete)

	guardedRouter.Use(security.MiddlewareAdmin)
}

func (controller *UserController) registerAdmin(w http.ResponseWriter, r *http.Request) {

	newUser := user.User{}

	err := web.UnmarshalJSON(r, &newUser)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse requested data", http.StatusBadRequest))
		return
	}

	err = controller.UserService.CreateAdmin(&newUser)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newUser)
}

func (controller *UserController) registerUser(w http.ResponseWriter, r *http.Request) {

	newUser := user.User{}

	err := web.UnmarshalJSON(r, &newUser)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse requested data", http.StatusBadRequest))
		return
	}

	newUser.CreatedBy, err = security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}

	newUser.Credentials.CreatedBy = newUser.CreatedBy

	err = controller.UserService.CreateUser(&newUser)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newUser)
}

func (controller *UserController) login(w http.ResponseWriter, r *http.Request) {

	userCredentials := credential.Credential{}
	claim := security.Claims{}

	err := web.UnmarshalJSON(r, &userCredentials)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse requested data", http.StatusBadRequest))
		return
	}

	err = controller.UserService.Login(&userCredentials, &claim)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	token, err := claim.GenerateToken()
	if err != nil {
		web.RespondError(w, err)
		return
	}

	// w.Header().Set("token", token)
	w.Header().Set("Authorization", "Bearer "+token)

	web.RespondJSON(w, http.StatusAccepted, map[string]string{
		"message": "Login successful",
		"token":   token,
	})
}

func (controller *UserController) getAllUsers(w http.ResponseWriter, r *http.Request) {
	allUsers := &[]user.UserDTO{}
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

	err = controller.UserService.GetAllUsers(allUsers, &totalCount, limit, offset)
	if err != nil {
		controller.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}
	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allUsers)
}

func (controller *UserController) getUserById(w http.ResponseWriter, r *http.Request) {

	var targetUser = &user.UserDTO{}

	parser := web.NewParser(r)

	userIdFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid user ID format"))
		return
	}

	targetUser.ID = userIdFromURL

	err = controller.UserService.GetUserByID(targetUser)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, targetUser)
}

func (controller *UserController) updateUserById(w http.ResponseWriter, r *http.Request) {

	var userToUpdate = user.User{}

	parser := web.NewParser(r)

	err := web.UnmarshalJSON(r, &userToUpdate)
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("unable to parse requested data", http.StatusBadRequest))
		return
	}

	userToUpdate.UpdatedBy, err = security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}

	userIdFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid user ID format"))
		return
	}

	userToUpdate.ID = userIdFromURL

	err = controller.UserService.UpdateUser(&userToUpdate)
	if err != nil {
		controller.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	updatedUser := user.UserDTO{}
	updatedUser.ID = userToUpdate.ID
	err = controller.UserService.GetUserByID(&updatedUser)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, updatedUser)
}

func (controller *UserController) deleteUserById(w http.ResponseWriter, r *http.Request) {

	userToDelete := user.User{}

	parser := web.NewParser(r)

	userIdFromURL, err := parser.GetUUID("id")
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid user ID format"))
		return
	}

	userToDelete.DeletedBy, err = security.ExtractUserIDFromToken(r)
	if err != nil {
		controller.log.Error(err.Error())
		web.RespondError(w, err)
		return
	}

	userToDelete.ID = userIdFromURL

	err = controller.UserService.Delete(&userToDelete)
	if err != nil {
		controller.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}
