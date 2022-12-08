package main

import (
	"context"
	"github.com/adarshsrinivasan/PressAndPlay/user/proto"
)

type userGRPCService struct {
	proto.UnimplementedUserServer
}

func (u *userGRPCService) GetUser(ctx context.Context, userModel *proto.UserModel) (*proto.UserModel, error) {
	myUserModel, _, err := GetUserByID(userModel.GetId())
	if err != nil {
		return nil, err
	}
	protoUserModel := &proto.UserModel{
		Id:          myUserModel.Id,
		FirstName:   myUserModel.FirstName,
		LastName:    myUserModel.LastName,
		DateOfBirth: myUserModel.DateOfBirth,
		Gender:      proto.Gender(myUserModel.Gender),
		Address:     &proto.Address{
			AddressLine1: myUserModel.Address.AddressLine1,
			AddressLine2: myUserModel.Address.AddressLine2,
			City:         myUserModel.Address.City,
			State:        myUserModel.Address.State,
			Country:      myUserModel.Address.Country,
			Pincode:      myUserModel.Address.Pincode,
		},
		Phone:       myUserModel.Phone,
		Role:        proto.Role(myUserModel.Role),
		Email:       myUserModel.Email,
		Password:    myUserModel.Password,
		Verified:    myUserModel.Verified,
	}
	return protoUserModel, nil
}
