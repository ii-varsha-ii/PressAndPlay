package main

import (
	"encoding/json"
	"fmt"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	API_PREFIX   = "/api/v1/user"
	ID_URL_REGEX = "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"
)

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		common.RespondWithStatusCode(w, http.StatusOK, nil)
	}
	var user UserModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("createUserHandler: exception while parsing request. %v", err))
		return
	}
	defer r.Body.Close()

	if updatedUser, statusCode, err := CreateUser(user); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("createUserHandler: exception while creating user. %v", err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusCreated, "", updatedUser)
	}
}

func loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var user UserModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("userLoginHandler: exception while parsing request. %v", err))
		return
	}
	defer r.Body.Close()

	if updatedUser, sessionID, statusCode, err := LoginUser(user); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("userLoginHandler: exception while authenticating user. %v", err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusCreated, sessionID, updatedUser)
	}
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("getUserHandler: Invalid session. Please login again"))
		return
	}

	if resultUser, statusCode, err := GetUserByID(vars["id"]); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("getUserHandler: exception while fetching user %s. %v",
			vars["id"], err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusOK, "", resultUser)
	}
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("deleteUserHandler: Invalid session. Please login again"))
		return
	}

	if statusCode, err := DeleteUserByID(vars["id"]); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("deleteUserHandler: exception while fetching user %s. %v",
			vars["id"], err))
		return
	} else {
		common.RespondWithStatusCode(w, http.StatusAccepted, nil)
	}
}

func initializeMuxRoutes() {
	httpRouter = mux.NewRouter()
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, "create"),
		createUserHandler).Methods("POST")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, "login"),
		loginUserHandler).Methods("POST")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, fmt.Sprintf("{id:%s}", ID_URL_REGEX)),
		getUserHandler).Methods("GET")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, fmt.Sprintf("{id:%s}", ID_URL_REGEX)),
		deleteUserHandler).Methods("DELETE")
}
