package main

import (
	"context"

	"github.com/adarshsrinivasan/PressAndPlay/libraries/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type eventsGRPCService struct {
	proto.UnimplementedEventsServer
}

func (u *eventsGRPCService) GetEventsByUserIdAndCourtId(ctx context.Context, eventsModel *proto.EventModel) (*proto.EventResponse, error) {
	myEventsModel, _, err := GetEventsByUserIdAndCourtId(eventsModel.UserId, eventsModel.CourtId)
	var result proto.EventResponse
	if err != nil {
		return nil, err
	}

	var protoEventsModel []*proto.EventModel
	for _, event := range myEventsModel {
		protoEventsModel = append(protoEventsModel, &proto.EventModel{
			Id:               event.Id,
			UserId:           event.UserID,
			ManagerId:        event.ManagerID,
			SlotId:           event.SlotID,
			BookingTimestamp: timestamppb.New(event.BookingTimestamp),
			TimeStartHHMM:    int32(event.TimeStartHHMM),
			TimeEndHHMM:      int32(event.TimeEndHHMM),
			Notified:         event.Notified,
		})
	}
	result.EventModelList = protoEventsModel
	return &result, nil
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

func getCourtByID(courtID string) (*proto.CourtModel, error) {
	court := &proto.CourtModel{
		Id: courtID,
	}
	result, err := gRPCCourtClient.GetCourt(ctx, court)
	if err != nil {
		return nil, err
	}
	return result, nil
}
