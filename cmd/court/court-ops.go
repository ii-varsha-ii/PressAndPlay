package main

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/adarshsrinivasan/PressAndPlay/libraries/proto"
	"github.com/google/uuid"
)

func CreateCourt(courtModel CourtModel) (CourtModel, int, error) {
	if err := validateCourtModel(&courtModel, true); err != nil {
		return CourtModel{}, http.StatusBadRequest, err
	}
	courtModel.Id = uuid.New().String()
	imageDownloadUrl, imageUploadUrl, err := genImageDownloadAndUploadUrl(courtModel.Id)
	if err != nil {
		return CourtModel{}, http.StatusInternalServerError, err
	}
	courtModel.ImageDownloadUrl = imageDownloadUrl
	courtModel.ImageUploadUrl = imageUploadUrl
	if statusCode, err := courtModel.createCourt(); err != nil {
		return CourtModel{}, statusCode, err
	}
	return courtModel, http.StatusOK, nil
}

func ListCourt(location string) ([]*CourtListModel, int, error) {
	courtModel := CourtModel{}
	courts, statusCode, err := courtModel.listCourt()
	if err != nil {
		return nil, statusCode, err
	}
	var courtList []*CourtListModel
	for _, court := range courts {
		court.Distance, err = calculateDistance(location, court.Location)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		courtList = append(courtList, &CourtListModel{
			Id:                  court.Id,
			Name:                court.Name,
			Distance:            court.Distance,
			Rating:              court.Rating,
			AvailableSlotsCount: getAvailableSlots(court.AvailableSlots),
			ImageUploadUrl:      court.ImageUploadUrl,
			ImageDownloadUrl:    court.ImageDownloadUrl,
			SportType:           court.SportType,
		})
	}
	sort.Slice(courtList[:], func(i, j int) bool {
		if courtList[i].Distance == courtList[j].Distance {
			return courtList[i].Rating < courtList[j].Rating
		}
		return courtList[i].Distance < courtList[j].Distance
	})
	return courtList, http.StatusOK, nil
}

func GetCourtByID(courtID string) (CourtModel, int, error) {
	courtModel := CourtModel{Id: courtID}

	if statusCode, err := courtModel.getByID(); err != nil {
		return CourtModel{}, statusCode, err
	}
	//getAvailableSlots(courtModel.AvailableSlots)
	return courtModel, http.StatusOK, nil
}

func DeleteCourtByID(courtID string) (int, error) {
	courtModel := CourtModel{Id: courtID}

	if statusCode, err := courtModel.deleteByID(); err != nil {
		return statusCode, err
	}
	notifyCourtDeletedEvent(courtID)
	deleteObjectFromCloud(courtID)
	return http.StatusOK, nil
}

func RateCourtByID(courtID, userID string, rate float64) (CourtModel, int, error) {
	courtModel := CourtModel{Id: courtID}

	if statusCode, err := courtModel.getByID(); err != nil {
		return CourtModel{}, statusCode, err
	}

	result, err := checkUserBooked(courtID, userID)
	if err != nil {
		return CourtModel{}, http.StatusInternalServerError, fmt.Errorf("exception while fetching user events. %v", err)
	}

	userModel, err := getUserByID(userID)
	if err != nil {
		return CourtModel{}, http.StatusInternalServerError, fmt.Errorf("exception while validating user. %v", err)
	}
	if userModel.Role != proto.Role_ROLE_CUSTOMER {
		return CourtModel{}, http.StatusInternalServerError, fmt.Errorf("given user id %s is not of type customer", userID)
	}

	if !result {
		return CourtModel{}, http.StatusBadRequest, fmt.Errorf("invalid rating attempt. user can provide rating 24 hrs after the time of appointment creation")
	}
	if rate > 5 || rate < 0 {
		return CourtModel{}, http.StatusBadRequest, fmt.Errorf("invalid rating attempt. accepted rating value are 0,1,2,3,4,5")
	}
	totalRating := courtModel.Rating * float64(courtModel.RatingCount)
	totalRating += float64(rate)
	courtModel.RatingCount += 1
	courtModel.Rating = totalRating / float64(courtModel.RatingCount)
	if statusCode, err := courtModel.updateCourt(); err != nil {
		return CourtModel{}, statusCode, err
	}
	return courtModel, http.StatusOK, nil
}

func BookCourtByID(courtID, slotID, userID string) (CourtModel, int, error) {
	courtModel := CourtModel{Id: courtID}
	if statusCode, err := courtModel.getByID(); err != nil {
		return CourtModel{}, statusCode, err
	}
	userModel, err := getUserByID(userID)
	if err != nil {
		return CourtModel{}, http.StatusInternalServerError, fmt.Errorf("exception while validating user. %v", err)
	}
	if userModel.Role != proto.Role_ROLE_CUSTOMER {
		return CourtModel{}, http.StatusInternalServerError, fmt.Errorf("given user id %s is not of type customer", userID)
	}
	for _, slot := range courtModel.AvailableSlots {
		if slot.SlotId == slotID {
			if slot.Booked {
				return CourtModel{}, http.StatusBadRequest, fmt.Errorf("slot %s in court %s already booked", slotID, courtID)
			} else {
				slot.Booked = true
				if statusCode, err := courtModel.updateCourt(); err != nil {
					return CourtModel{}, statusCode, err
				}
				notifySlotBookedEvent(userID, courtModel.ManagerId, courtID, slotID)
				return courtModel, http.StatusOK, nil
			}
		}
	}
	return CourtModel{}, http.StatusBadRequest, fmt.Errorf("slot %s not found in court %s", slotID, courtID)
}

func checkUserBooked(courtID, userID string) (bool, error) {
	return true, nil
	//events, err := getUserEvents(userID, courtID)
	//if err != nil {
	//	return false, err
	//}
	//currentTime := time.Now()
	//for _, event := range events {
	//	if currentTime.Sub(event.CreatedAt.AsTime()).Hours() >= 24 {
	//		return true, nil
	//	}
	//}
	//return false, nil
}

func calculateDistance(givenLocation string, courtLocation string) (float64, error) {
	if givenLocation == "" {
		return -1, nil
	}
	location1 := strings.Split(givenLocation, ",")
	if len(location1) != 2 {
		return 0, fmt.Errorf("malformed Location, format: \"<latitude>,<longitude>\"")
	}
	lat1, err := strconv.ParseFloat(strings.TrimSpace(location1[0]), 64)
	if err != nil {
		return 0, fmt.Errorf("malformed Location, invalid latitude")
	}
	lng1, err := strconv.ParseFloat(strings.TrimSpace(location1[1]), 64)
	if err != nil {
		return 0, fmt.Errorf("imalformed Location, invalid longitude")
	}

	location2 := strings.Split(courtLocation, ",")
	lat2, _ := strconv.ParseFloat(strings.TrimSpace(location2[0]), 64)
	lng2, _ := strconv.ParseFloat(strings.TrimSpace(location2[1]), 64)

	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}
	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	return dist, nil
}

func getAvailableSlots(slots Slots) int {
	openSlots := 0
	//newLayout := "1504"
	//currentTime, _ := time.Parse(newLayout, time.Now().Format(newLayout))
	//for _, slot := range slots {
	//	slotStartTime, _ := time.Parse(newLayout, strconv.Itoa(slot.TimeStartHHMM))
	//	if currentTime.Sub(slotStartTime).Minutes() < 0 {
	//		openSlots++
	//	} else {
	//		slot.Booked = true
	//	}
	//}
	for _, slot := range slots {
		if !slot.Booked {
			openSlots++
		}
	}
	return openSlots
}
