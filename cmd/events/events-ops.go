package main

import (
	"fmt"
	"net/http"
)

func CreateEvent(eventsModel EventsModel) (EventsModel, int, error) {
	if err := validateEventModel(&eventsModel, true); err != nil {
		return EventsModel{}, http.StatusBadRequest, err
	}
	eventsDBData := convertModelToDBData(eventsModel)
	court, err := getCourtByID(eventsModel.CourtID)
	if err != nil {
		return EventsModel{}, http.StatusInternalServerError, err
	}
	for _, slot := range court.AvailableSlots {
		if slot.SlotId == eventsModel.SlotID && slot.Booked {
			eventsDBData.TimeStartHHMM = int(slot.TimeStartHHMM)
			eventsDBData.TimeEndHHMM = int(slot.TimeEndHHMM)
		} else {
			return EventsModel{}, http.StatusInternalServerError, fmt.Errorf("slot %s not found", eventsModel.SlotID)
		}
	}
	if statusCode, err := eventsDBData.createEvent(); err != nil {
		return EventsModel{}, statusCode, err
	}
	eventModel := convertDBDataToModel(eventsDBData)
	return eventModel, http.StatusCreated, nil
}

func GetEventsByUserIdAndCourtId(userID string, courtID string) ([]*EventsModel, int, error) {
	eventsDBData := EventsDBData{UserID: userID, CourtID: courtID}
	eventsData, statusCode, err := eventsDBData.getByUserIdAndCourtId()
	if err != nil {
		return []*EventsModel{}, statusCode, err
	}
	var eventsResult []*EventsModel
	for _, event := range eventsData {
		eventModel := convertDBDataToModel(event)
		eventsResult = append(eventsResult, &eventModel)
	}
	return eventsResult, http.StatusOK, nil
}

func GetHistory(managerId string) ([]*EventsListModel, int, error) {
	eventsModel := EventsModel{ManagerID: managerId, Notified: true} // bad design: change later
	eventsDBData := convertModelToDBData(eventsModel)
	eventsDBDataList, statusCode, err := eventsDBData.listByManagerID()
	if err != nil {
		return []*EventsListModel{}, statusCode, err
	}
	var eventsResult []*EventsListModel
	for _, event := range eventsDBDataList {
		eventModel := convertDBDataToModel(event)
		user, _ := getUserByID(eventModel.UserID)
		court, _ := getCourtByID(eventModel.CourtID)
		eventsResult = append(eventsResult, &EventsListModel{
			Id:               event.Id,
			UserFirstName:    user.FirstName,
			UserLastName:     user.LastName,
			UserPhone:        user.Phone,
			CourtName:        court.Name,
			SlotID:           event.SlotID,
			TimeStartHHMM:    event.TimeStartHHMM,
			TimeEndHHMM:      event.TimeEndHHMM,
			BookingTimestamp: event.BookingTimestamp,
		})
	}
	return eventsResult, http.StatusOK, nil
}

func ListUnreadEvents(managerId string) ([]*EventsListModel, int, error) {
	eventsModel := EventsModel{ManagerID: managerId, Notified: false}
	eventsDBData := convertModelToDBData(eventsModel)
	eventsDBDataList, statusCode, err := eventsDBData.listByManagerID()
	if err != nil {
		return []*EventsListModel{}, statusCode, err
	}
	var eventsResult []*EventsListModel
	for _, event := range eventsDBDataList {
		eventModel := convertDBDataToModel(event)
		user, _ := getUserByID(eventModel.UserID)
		court, _ := getCourtByID(eventModel.CourtID)
		eventsResult = append(eventsResult, &EventsListModel{
			Id:               event.Id,
			UserFirstName:    user.FirstName,
			UserLastName:     user.LastName,
			UserPhone:        user.Phone,
			CourtName:        court.Name,
			SlotID:           event.SlotID,
			TimeStartHHMM:    event.TimeStartHHMM,
			TimeEndHHMM:      event.TimeEndHHMM,
			BookingTimestamp: event.BookingTimestamp,
		})
		event.Notified = true
		event.updateByID()
	}
	return eventsResult, http.StatusOK, nil
}
