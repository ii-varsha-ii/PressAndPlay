package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun/schema"
)

const (
	EventsTableName = "user_data"
)

type EventsModel struct {
	Id               string    `json:"id" bun:"id"`
	UserID           string    `json:"userId" bun:"userId"`
	ManagerID        string    `json:"managerId" bun:"managerId"`
	CourtID          string    `json:"courtId" bun:"courtId"`
	SlotID           string    `json:"slotId" bun:"slotId"`
	TimeStartHHMM    int       `json:"timeStartHHMM" bson:"timeStartHHMM"`
	TimeEndHHMM      int       `json:"timeEndHHMM" bson:"timeEndHHMM"`
	BookingTimestamp time.Time `json:"bookingTimestamp" bun:"bookingTimestamp"`
	Notified         bool      `json:"notified" bun:"notified"`
}

type EventsListModel struct {
	Id               string    `json:"id" bun:"id"`
	UserFirstName    string    `json:"userFirstName" bun:"userId"`
	UserLastName     string    `json:"courtId" bun:"courtId"`
	UserPhone        string    `json:"userContact" bun:"userContact"`
	CourtName        string    `json:"courtName" bun:"courtName"`
	SlotID           string    `json:"slotId" bun:"slotId"`
	TimeStartHHMM    int       `json:"timeStartHHMM" bson:"timeStartHHMM"`
	TimeEndHHMM      int       `json:"timeEndHHMM" bson:"timeEndHHMM"`
	BookingTimestamp time.Time `json:"bookingTimestamp" bun:"bookingTimestamp"`
}

type EventsDBData struct {
	schema.BaseModel `bun:"table:events_data"`
	Id               string            `json:"id" bun:"id"`
	UserID           string            `json:"userId" bun:"userId"`
	ManagerID        string            `json:"managerId" bun:"managerId"`
	CourtID          string            `json:"courtId" bun:"courtId"`
	SlotID           string            `json:"slotId" bun:"slotId"`
	TimeStartHHMM    int               `json:"timeStartHHMM" bson:"timeStartHHMM"`
	TimeEndHHMM      int               `json:"timeEndHHMM" bson:"timeEndHHMM"`
	BookingTimestamp time.Time         `json:"bookingTimestamp" bun:"bookingTimestamp"`
	Notified         bool              `json:"notified" bun:"notified"`
	Tags             map[string]string `json:"tags" bson:"tags,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"  bun:"createdAt,omitempty"`
	UpdatedAt        time.Time         `json:"updatedAt" bun:"updatedAt,omitempty"`
}
type EventResponse struct {
}
type EventsDBOps interface {
	createEvent() (int, error)
	listUnreadEvents() (int, error)
	getHistory() (int, error)
}

func (event *EventsDBData) createEvent() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Insert", EventsTableName, err)
	}
	event.Id = uuid.New().String()
	event.Tags = map[string]string{}
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	event.Notified = false

	if _, err := dbClient.NewInsert().Model(event).Exec(context.TODO()); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Insert", EventsTableName, err)
	}
	return http.StatusOK, nil
}

func (event *EventsDBData) getByUserIdAndCourtId() ([]EventsDBData, int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return []EventsDBData{}, http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", EventsTableName, err)
	}
	whereClause := []WhereClauseType{
		{
			ColumnName:   "userId",
			RelationType: EQUAL,
			ColumnValue:  event.UserID,
		},
		{
			ColumnName:   "courtId",
			RelationType: EQUAL,
			ColumnValue:  event.CourtID,
		},
	}
	_, eventData, _, statusCode, err := readUtil(nil, whereClause, nil, nil, nil, true)
	if err != nil {
		return []EventsDBData{}, statusCode, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", EventsTableName, err)
	}
	return eventData, http.StatusOK, nil
}

func (event *EventsDBData) listByManagerID() ([]EventsDBData, int, error) {

	if err := verifyDatabaseConnection(dbClient); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Insert", EventsTableName, err)
	}
	whereClause := []WhereClauseType{
		{
			ColumnName:   "managerId",
			RelationType: EQUAL,
			ColumnValue:  event.ManagerID,
		},
	}
	if event.Notified == false {
		whereClause = append(whereClause, WhereClauseType{
			ColumnName:   "notified",
			RelationType: EQUAL,
			ColumnValue:  false,
		})
	}

	orderByClause := []string{"bookingTimestamp"}
	_, eventsList, _, statusCode, err := readUtil(nil, whereClause, orderByClause, nil, nil, true)
	if err != nil {
		return nil, statusCode, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", EventsTableName, err)
	}

	return eventsList, http.StatusOK, nil
}

func (event *EventsDBData) updateByID() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Update", EventsTableName, err)
	}
	event.UpdatedAt = time.Now()
	updateQuery := dbClient.NewUpdate().Model(event)
	updateQuery = updateQuery.WherePK()
	oldVersion := 0
	prepareUpdateQuery(updateQuery, &oldVersion, event, true, true)
	if _, err := updateQuery.Exec(context.TODO()); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Update", EventsTableName, err)
	}
	return http.StatusOK, nil
}

func convertDBDataToModel(eventsDbData EventsDBData) EventsModel {
	eventsModel := EventsModel{
		Id:               eventsDbData.Id,
		UserID:           eventsDbData.UserID,
		ManagerID:        eventsDbData.ManagerID,
		CourtID:          eventsDbData.CourtID,
		SlotID:           eventsDbData.SlotID,
		TimeStartHHMM:    eventsDbData.TimeStartHHMM,
		TimeEndHHMM:      eventsDbData.TimeEndHHMM,
		BookingTimestamp: eventsDbData.BookingTimestamp,
		Notified:         eventsDbData.Notified,
	}
	return eventsModel
}
func convertModelToDBData(eventsModel EventsModel) EventsDBData {
	eventsDBData := EventsDBData{
		Id:               eventsModel.Id,
		UserID:           eventsModel.UserID,
		ManagerID:        eventsModel.ManagerID,
		CourtID:          eventsModel.CourtID,
		SlotID:           eventsModel.SlotID,
		TimeStartHHMM:    eventsModel.TimeStartHHMM,
		TimeEndHHMM:      eventsModel.TimeEndHHMM,
		BookingTimestamp: eventsModel.BookingTimestamp,
		Notified:         eventsModel.Notified,
	}
	return eventsDBData
}

func validateEventModel(eventModel *EventsModel, create bool) error {
	seen := map[string]bool{}
	if eventModel.UserID == "" {
		return fmt.Errorf("invalid EventModel. Empty UserId")
	}
	if eventModel.ManagerID == "" {
		return fmt.Errorf("invalid EventModel. Empty ManagerId")
	}
	if eventModel.CourtID == "" {
		return fmt.Errorf("invalid EventModel. Empty ManagerId")
	}
	if eventModel.SlotID == "" {
		return fmt.Errorf("invalid EventModel. Empty Slot ID")
	}
	if _, ok := seen[eventModel.SlotID]; ok {
		return fmt.Errorf("invalid EventModel. Duplicate slot id %s", eventModel.SlotID)
	}
	seen[eventModel.SlotID] = true
	if eventModel.TimeStartHHMM == 0 {
		return fmt.Errorf("invalid EventModel. Empty Time Start for slot %s", eventModel.SlotID)
	}

	if eventModel.TimeEndHHMM == 0 {
		return fmt.Errorf("invalid EventModel. Empty Time End for slot %s", eventModel.SlotID)
	}

	newLayout := "1504"
	_, err := time.Parse(newLayout, strconv.Itoa(eventModel.TimeStartHHMM))
	if err != nil {
		return fmt.Errorf("invalid EventModel. exception while parsing Time Start for slot %s. %v", eventModel.SlotID, err)

	}
	_, err = time.Parse(newLayout, strconv.Itoa(eventModel.TimeEndHHMM))
	if err != nil {
		return fmt.Errorf("invalid EventModel. exception while parsing Time End for slot %s. %v", eventModel.SlotID, err)

	}

	if eventModel.TimeStartHHMM > eventModel.TimeEndHHMM {
		return fmt.Errorf("invalid EventModel. Start Time greater than End Time for slot %s", eventModel.SlotID)
	}
	return nil
}
