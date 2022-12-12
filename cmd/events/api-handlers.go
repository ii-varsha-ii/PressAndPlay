package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"

	"github.com/gorilla/mux"
)

const (
	API_PREFIX   = "/api/v1/events"
	ID_URL_REGEX = "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"
)

func createEvent(w http.ResponseWriter, r *http.Request) {
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("createEventHandler: Invalid session. Please login again"))
		return
	}
	var event EventsModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&event); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("createEventHandler: exception while parsing request. %v", err))
		return
	}
	defer r.Body.Close()

	if createdEvent, statusCode, err := CreateEvent(event); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("createEventHandler: exception while creating event. %v", err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusCreated, "", createdEvent)
	}
}

func listUnreadEventsHandler(w http.ResponseWriter, r *http.Request) {
	if !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("getUserHandler: Invalid session. Please login again"))
		return
	}
	managerId := getUserIdFromSession(r.Header.Get("User-Session-Id"))
	if resultEvents, statusCode, err := ListUnreadEvents(managerId); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("listUnreadEventsHandler: exception while fetching list for unread events. %v",
			err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusOK, "", resultEvents)
	}
}
func getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("User-Session-Id") != "" && !validateSessionID(r.Header.Get("User-Session-Id")) {
		common.RespondWithError(w, http.StatusForbidden, fmt.Sprintf("getUserHandler: Invalid session. Please login again"))
		return
	}
	userId := getUserIdFromSession(r.Header.Get("User-Session-Id"))
	//vars := mux.Vars(r)
	//TODO: validate Start and end date

	if resultEvents, statusCode, err := GetHistory(userId); err != nil {
		common.RespondWithError(w, statusCode, fmt.Sprintf("listUnreadEventsHandler: exception while fetching list for unread events. %v",
			err))
		return
	} else {
		common.RespondWithJSON(w, http.StatusOK, "", resultEvents)
	}
}

func initializeMuxRoutes() {
	httpRouter = mux.NewRouter()
	httpRouter.HandleFunc(fmt.Sprintf("%s", API_PREFIX),
		getHistoryHandler).Methods("GET")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, "create"),
		createEvent).Methods("POST")
	httpRouter.HandleFunc(fmt.Sprintf("%s/%s", API_PREFIX, "notifications"),
		listUnreadEventsHandler).Methods("GET")
}
