package main

import (
	"net/http"
	"time"
)

func CreateEvent(eventsModel EventsModel) (EventsModel, int, error) {
	if err := validateEventModel(&eventsModel, true); err != nil {
		return EventsModel{}, http.StatusBadRequest, err
	}
	eventsDBData := convertModelToDBData(eventsModel)
	eventsDBData.BookingTimestamp = time.Now() // remove
	court, err := getCourtByID(eventsModel.CourtID)
	if err != nil {
		return EventsModel{}, http.StatusInternalServerError, err
	}
	for _, slot := range court.AvailableSlots {
		if slot.SlotId == eventsModel.SlotID && slot.Booked {
			eventsDBData.TimeStartHHMM = int(slot.TimeStartHHMM)
			eventsDBData.TimeEndHHMM = int(slot.TimeEndHHMM)
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
		eventDBData := EventsDBData{}
		copyEventDBData(&eventDBData, &event)
		eventDBData.Notified = true
		eventDBData.updateByID()
		//x, _, _ := eventsDBData.listByManagerID()
		//fmt.Println(x)
	}

	//fmt.Print("Events: ", eventsDBDataList)
	return eventsResult, http.StatusOK, nil
}
