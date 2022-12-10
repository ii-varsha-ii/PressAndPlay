package main

import (
	"context"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/proto"
)

type courtGRPCService struct {
	proto.UnimplementedCourtServer
}

func (c *courtGRPCService) GetCourt(ctx context.Context, courtModel *proto.CourtModel) (*proto.CourtModel, error) {
	myCourtModel, _, err := GetCourtByID(courtModel.GetId())
	if err != nil {
		return nil, err
	}
	protoCourtModel := &proto.CourtModel{
		Id:   myCourtModel.Id,
		Name: myCourtModel.Name,
		Address: &proto.Address{
			AddressLine1: myCourtModel.Address.AddressLine1,
			AddressLine2: myCourtModel.Address.AddressLine2,
			City:         myCourtModel.Address.City,
			State:        myCourtModel.Address.State,
			Country:      myCourtModel.Address.Country,
			Pincode:      myCourtModel.Address.Pincode,
		},
		Location:         myCourtModel.Location,
		Distance:         float32(myCourtModel.Distance),
		Phone:            myCourtModel.Phone,
		Rating:           float32(myCourtModel.Rating),
		RatingCount:      int32(myCourtModel.RatingCount),
		ManagerId:        myCourtModel.ManagerId,
		AvailableSlots:   convertSlots(myCourtModel.AvailableSlots),
		ImageUploadUrl:   myCourtModel.ImageUploadUrl,
		ImageDownloadUrl: myCourtModel.ImageDownloadUrl,
		SportType:        myCourtModel.SportType,
		Verified:         myCourtModel.Verified,
	}
	return protoCourtModel, nil
}

func getUserEvents(userID, courtID string) ([]*proto.EventModel, error) {
	events := &proto.EventModel{
		UserId:  userID,
		CourtId: courtID,
	}
	result, err := gRPCEventClient.GetEvents(ctx, events)
	if err != nil {
		return nil, err
	}
	return result.EventModelList, nil
}

func getUserByID(userID string) (*proto.UserModel, error) {
	user := &proto.UserModel{
		Id: userID,
	}
	result, err := gRPCUserClient.GetUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func convertSlots(slots Slots) []*proto.Slot {
	var protoSlots []*proto.Slot
	for _, slot := range slots {
		protoSlots = append(protoSlots, &proto.Slot{
			SlotId:        slot.SlotId,
			TimeStartHHMM: int32(slot.TimeStartHHMM),
			TimeEndHHMM:   int32(slot.TimeEndHHMM),
			Booked:        slot.Booked,
		})
	}
	return protoSlots
}
