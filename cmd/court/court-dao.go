package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/proto"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	CourtTableName = "court_data"
)

type Address struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Pincode      string `json:"pincode"`
}

func (a *Address) ObjToString() string {
	bytesAddress, _ := json.Marshal(a)
	return string(bytesAddress)
}

func (a *Address) StringToObj(stringAddress string) {
	json.Unmarshal([]byte(stringAddress), a)
}

type Slot struct {
	SlotId        string `json:"slot_id" bson:"slot_id"`
	TimeStartHHMM int    `json:"time_start_hhmm" bson:"time_start_hhmm"`
	TimeEndHHMM   int    `json:"time_end_hhmm" bson:"time_end_hhmm"`
	Booked        bool   `json:"booked" bson:"booked"`
}

type Slots []*Slot

func (s *Slots) ObjToString() string {
	bytesAddress, _ := json.Marshal(s)
	return string(bytesAddress)
}

func (s *Slots) StringToObj(stringSlots string) {
	json.Unmarshal([]byte(stringSlots), s)
}

type CourtModel struct {
	Id               string            `json:"id" bson:"_id,omitempty"`
	Name             string            `json:"name" bson:"name,omitempty"`
	Address          Address           `json:"address" bson:"address,omitempty"`
	Location         string            `json:"location" bson:"location,omitempty"`
	Distance         float64           `json:"distance" bson:"distance"`
	Phone            string            `json:"phone" bson:"phone,omitempty"`
	Rating           float64           `json:"rating" bson:"rating,omitempty"`
	RatingCount      int               `json:"ratingCount" bson:"ratingCount, omitempty"`
	AvailableSlots   Slots             `json:"availableSlots" bson:"availableSlots,omitempty"`
	ImageUploadUrl   string            `json:"imageUploadUrl" bson:"imageUploadUrl"`
	ImageDownloadUrl string            `json:"imageDownloadUrl" bson:"imageDownloadUrl"`
	SportType        string            `json:"sportType" bson:"sportType,omitempty"`
	Tags             map[string]string `json:"tags" bson:"tags,omitempty"`
	Blob             []byte            `json:"blob" bson:"blob"`
	Verified         bool              `json:"verified" bson:"verified,omitempty"`
	ManagerId        string            `json:"managerId" bson:"managerId"`
	CreatedAt        time.Time         `json:"createdAt"  bson:"createdAt,omitempty"`
	UpdatedAt        time.Time         `json:"updatedAt" bson:"updatedAt,omitempty"`
}

type CourtListModel struct {
	Id                  string  `json:"id" bson:"_id,omitempty"`
	Name                string  `json:"name" bson:"name,omitempty"`
	Distance            float64 `json:"distance" bson:"distance"`
	Rating              float64 `json:"rating" bson:"rating,omitempty"`
	AvailableSlotsCount int     `json:"availableSlotsCount" bson:"availableSlotsCount, omitempty"`
	ImageUploadUrl      string  `json:"imageUploadUrl" bson:"imageUploadUrl"`
	ImageDownloadUrl    string  `json:"imageDownloadUrl" bson:"imageDownloadUrl"`
	SportType           string  `json:"sportType" bson:"sportType,omitempty"`
}

type CourtDBOps interface {
	createCourt() (int, error)
	getByID() (int, error)
	listCourt() ([]*CourtModel, int, error)
	updateCourt() (int, error)
	deleteByID() (int, error)
}

func (court *CourtModel) createCourt() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Insert", CourtTableName, err)
	}

	court.Id = uuid.New().String()
	court.Tags = map[string]string{}
	court.Blob = nil
	court.Rating = 5
	court.RatingCount = 1
	court.CreatedAt = time.Now()
	court.UpdatedAt = time.Now()

	_, err := dbClient.InsertOne(context.TODO(), court)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Insert", CourtTableName, err)

	}
	return http.StatusOK, nil

}

func (court *CourtModel) getByID() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", CourtTableName, err)
	}
	result := dbClient.FindOne(ctx, bson.M{"_id": court.Id})
	if result.Err() != nil {
		return http.StatusInternalServerError, fmt.Errorf("exception while performing %s Operation on Table: %s. %v", "Read", CourtTableName, result.Err())
	}
	if err := result.Decode(court); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("exception while parsing response %s Operation on Table: %s. %v", "Read", CourtTableName, result.Err())
	}
	return http.StatusOK, nil
}

func (court *CourtModel) getByLocation() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", CourtTableName, err)
	}
	result := dbClient.FindOne(ctx, bson.M{"location": court.Location})
	if result.Err() != nil {
		return http.StatusInternalServerError, fmt.Errorf("exception while performing %s Operation on Table: %s. %v", "Read", CourtTableName, result.Err())
	}
	if err := result.Decode(court); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("exception while parsing response %s Operation on Table: %s. %v", "Read", CourtTableName, result.Err())
	}
	return http.StatusOK, nil
}

func (court *CourtModel) listCourt() ([]*CourtModel, int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "List", CourtTableName, err)
	}

	cursor, err := dbClient.Find(ctx, bson.M{})
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("exception while performing %s Operation on Table: %s. %v", "List", CourtTableName, err)
	}
	var courts []*CourtModel
	if err = cursor.All(ctx, &courts); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("exception while parsing response %s Operation on Table: %s. %v", "Read", CourtTableName, err)
	}
	return courts, http.StatusOK, nil
}

func (court *CourtModel) updateCourt() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Update", CourtTableName, err)
	}
	court.UpdatedAt = time.Now()
	bsonCourt, _ := toDoc(court)
	update := bson.D{{"$set", bsonCourt}}
	if _, err := dbClient.UpdateByID(ctx, court.Id, update); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("exception while performing %s Operation on Table: %s. %v", "Update", CourtTableName, err)
	}
	return http.StatusOK, nil
}

func (court *CourtModel) deleteByID() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Delete", CourtTableName, err)
	}
	if _, err := dbClient.DeleteOne(ctx, bson.M{"_id": court.Id}); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("exception while performing %s Operation on Table: %s. %v", "Delete", CourtTableName, err)
	}
	return http.StatusOK, nil
}

func validateCourtModel(courtModel *CourtModel, create bool) error {
	seen := map[string]bool{}
	if courtModel.Name == "" {
		return fmt.Errorf("invalid CourtModel. Empty Name")
	}
	if courtModel.ManagerId == "" {
		return fmt.Errorf("invalid CourtModel. Empty ManagerId")
	}
	userModel, err := getUserByID(courtModel.ManagerId)
	if err != nil {
		return fmt.Errorf("invalid CourtModel. exception while validating manager. %v", err)
	}
	if userModel.Role != proto.Role_ROLE_MANAGER {
		return fmt.Errorf("invalid CourtModel. given user id %s is not of type manager", courtModel.ManagerId)
	}
	if courtModel.Address.AddressLine1 == "" {
		return fmt.Errorf("invalid CourtModel. Empty Address Line 1")
	}
	if courtModel.Address.City == "" {
		return fmt.Errorf("invalid CourtModel. Empty City")
	}
	if courtModel.Address.State == "" {
		return fmt.Errorf("invalid CourtModel. Empty State")
	}
	if courtModel.Address.Country == "" {
		return fmt.Errorf("invalid CourtModel. Empty Country")
	}
	if courtModel.Address.Pincode == "" {
		return fmt.Errorf("invalid CourtModel. Empty Pincode")
	}
	if courtModel.Location == "" {
		return fmt.Errorf("invalid CourtModel. Empty Location")
	}
	location := strings.Split(courtModel.Location, ",")
	if len(location) != 2 {
		return fmt.Errorf("invalid CourtModel. Malformed Location, format: \"<latitude>,<longitude>\"")
	}
	_, err = strconv.ParseFloat(strings.TrimSpace(location[0]), 64)
	if err != nil {
		return fmt.Errorf("invalid CourtModel. Malformed Location, invalid latitude")
	}
	_, err = strconv.ParseFloat(strings.TrimSpace(location[1]), 64)
	if err != nil {
		return fmt.Errorf("invalid CourtModel. Malformed Location, invalid longitude")
	}
	if courtModel.Phone == "" {
		return fmt.Errorf("invalid CourtModel. Empty Phone")
	}
	re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
	if !re.MatchString(courtModel.Phone) {
		return fmt.Errorf("invalid CourtModel. Invalid Phonenumber %v", courtModel.Phone)
	}
	if len(courtModel.AvailableSlots) == 0 {
		return fmt.Errorf("invalid CourtModel. Empty Available Slots")
	}
	for _, slot := range courtModel.AvailableSlots {
		if slot.SlotId == "" {
			return fmt.Errorf("invalid CourtModel. Empty Slot ID")
		}
		if _, ok := seen[slot.SlotId]; ok {
			return fmt.Errorf("invalid CourtModel. Duplicate slot id %s", slot.SlotId)
		}
		seen[slot.SlotId] = true
		if slot.TimeStartHHMM == 0 {
			return fmt.Errorf("invalid CourtModel. Empty Time Start for slot %s", slot.SlotId)
		}

		if slot.TimeEndHHMM == 0 {
			return fmt.Errorf("invalid CourtModel. Empty Time End for slot %s", slot.SlotId)
		}

		newLayout := "1504"
		_, err := time.Parse(newLayout, strconv.Itoa(slot.TimeStartHHMM))
		if err != nil {
			return fmt.Errorf("invalid CourtModel. exception while parsing Time Start for slot %s. %v", slot.SlotId, err)

		}
		_, err = time.Parse(newLayout, strconv.Itoa(slot.TimeEndHHMM))
		if err != nil {
			return fmt.Errorf("invalid CourtModel. exception while parsing Time End for slot %s. %v", slot.SlotId, err)

		}

		if slot.TimeStartHHMM > slot.TimeEndHHMM {
			return fmt.Errorf("invalid CourtModel. Start Time greater than End Time for slot %s", slot.SlotId)
		}
	}
	if create {
		if _, err := courtModel.getByLocation(); err == nil {
			return fmt.Errorf("court already registered")
		}
	}
	if courtModel.SportType == "" {
		return fmt.Errorf("invalid CourtModel. Empty Available Slots")
	}

	return nil
}

func toDoc(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}
