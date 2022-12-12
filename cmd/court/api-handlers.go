package main

import (
	"encoding/json"
	"fmt"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	API_PREFIX    = "/api/v1/court"
	ID_URL_REGEX  = "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"
	SLOT_ID_REGEX = "[0-9]+"
)

func createCourtHandler(w http.ResponseWriter, r *http.Request) {
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		common.RespondWithStatusCode(w, http.StatusOK, nil)
	}
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("getUserHandler: Invalid session. Please login again"))
		return
	}
	userID := getUserIdFromSession(r.Header.Get("User-Session-Id"))
	if userID == "" {
		common.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("rateCourtHandler: exception while reading user id from session"))
		return
	}
	var courtModel CourtModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&courtModel); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("createCourtHandler: exception while parsing request. %v", err))
		return
	}
	defer r.Body.Close()
	courtModel.ManagerId = userID
	if updatedUser, statusCode, err := CreateCourt(courtModel); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("createUserHandler: exception while creating user. %v", err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusCreated, "", updatedUser)
	}
}

func listCourtHandler(w http.ResponseWriter, r *http.Request) {
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		common.RespondWithStatusCode(w, http.StatusOK, nil)
	}
	if r.Header.Get("User-Session-Id") != "" && !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("listCourtHandler: Invalid session. Please login again"))
		return
	}
	vars := mux.Vars(r)

	if resultCourts, statusCode, err := ListCourt(r.Header.Get("location")); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("listCourtHandler: exception while fetching courts list for location %s. %v",
			vars["location"], err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusOK, "", resultCourts)
	}
}

func getCourtHandler(w http.ResponseWriter, r *http.Request) {
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		common.RespondWithStatusCode(w, http.StatusOK, nil)
	}
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("getCourtHandler: Invalid session. Please login again"))
		return
	}
	vars := mux.Vars(r)

	if resultCourt, statusCode, err := GetCourtByID(vars["id"]); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("getCourtHandler: exception while fetching court %s. %v",
			vars["id"], err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusOK, "", resultCourt)
	}
}

func deleteCourtHandler(w http.ResponseWriter, r *http.Request) {
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		common.RespondWithStatusCode(w, http.StatusOK, nil)
	}
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("deleteCourtHandler: Invalid session. Please login again"))
		return
	}
	vars := mux.Vars(r)

	if statusCode, err := DeleteCourtByID(vars["id"]); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("deleteCourtHandler: exception while deleting court %s. %v",
			vars["id"], err))
		return
	} else {
		common.RespondWithStatusCode(w, http.StatusAccepted, nil)
	}
}

func rateCourtHandler(w http.ResponseWriter, r *http.Request) {
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		common.RespondWithStatusCode(w, http.StatusOK, nil)
	}
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("rateCourtHandler: Invalid session. Please login again"))
		return
	}
	userID := getUserIdFromSession(r.Header.Get("User-Session-Id"))
	if userID == "" {
		common.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("rateCourtHandler: exception while reading user id from session"))
		return
	}
	var courtModel CourtModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&courtModel); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("rateCourtHandler: exception while parsing request. %v", err))
		return
	}
	vars := mux.Vars(r)
	if resultCourt, statusCode, err := RateCourtByID(vars["id"], userID, courtModel.Rating); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("rateCourtHandler: exception while upating rating of court %s. %v",
			vars["id"], err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusOK, "", resultCourt)
	}
}

func bookCourtHandler(w http.ResponseWriter, r *http.Request) {
	// Stop here if its Preflighted OPTIONS request
	if r.Method == "OPTIONS" {
		common.RespondWithStatusCode(w, http.StatusOK, nil)
	}
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("rateCourtHandler: Invalid session. Please login again"))
		return
	}
	userID := getUserIdFromSession(r.Header.Get("User-Session-Id"))
	if userID == "" {
		common.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("rateCourtHandler: exception while reading user id from session"))
		return
	}
	vars := mux.Vars(r)
	if resultCourt, statusCode, err := BookCourtByID(vars["court-id"], vars["slot-id"], userID); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("rateCourtHandler: exception while upating rating of court %s. %v",
			vars["id"], err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusOK, "", resultCourt)
	}
}

func initializeMuxRoutes() {
	httpRouter = mux.NewRouter()
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, "create"),
		createCourtHandler).Methods("POST", "OPTIONS")
	httpRouter.HandleFunc(fmt.Sprintf("%s", API_PREFIX),
		listCourtHandler).Methods("GET", "OPTIONS")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, fmt.Sprintf("{id:%s}", ID_URL_REGEX)),
		getCourtHandler).Methods("GET", "OPTIONS")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, fmt.Sprintf("{id:%s}", ID_URL_REGEX)),
		deleteCourtHandler).Methods("DELETE", "OPTIONS")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s/%s", API_PREFIX, fmt.Sprintf("{id:%s}", ID_URL_REGEX), "rating"),
		rateCourtHandler).Methods("POST", "OPTIONS")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s/%s/%s/%s", API_PREFIX,
		fmt.Sprintf("{court-id:%s}", ID_URL_REGEX),
		"slot",
		fmt.Sprintf("{slot-id:%s}", SLOT_ID_REGEX),
		"book"),
		bookCourtHandler).Methods("POST", "OPTIONS")
}
