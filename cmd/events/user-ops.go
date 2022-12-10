package main

import (
	"net/http"
)

func CreateUser(userModel UserModel) (UserModel, int, error) {
	if err := validateUserModel(&userModel, true); err != nil {
		return userModel, http.StatusBadRequest, err
	}
	userDBData := convertModelToDBData(userModel)
	statusCode, err := userDBData.createUser()
	if err != nil {
		return userModel, statusCode, err
	}
	updatedUserModel := convertDBDataToModel(userDBData, false)
	return updatedUserModel, http.StatusOK, nil
}

func LoginUser(userModel UserModel) (UserModel, string, int, error) {
	if err := validateLoginModel(&userModel); err != nil {
		return userModel, "", http.StatusBadRequest, err
	}
	userDBData := convertModelToDBData(userModel)
	statusCode, err := userDBData.login()
	if err != nil {
		return userModel, "", statusCode, err
	}

	updatedUserModel := convertDBDataToModel(userDBData, false)
	return updatedUserModel, userDBData.LastSessionID, http.StatusOK, nil
}

func GetUserByID(userID string) (UserModel, int, error) {
	userDBData := UserDBData{Id: userID}
	statusCode, err := userDBData.getByID()
	if err != nil {
		return UserModel{}, statusCode, err
	}
	updatedUserModel := convertDBDataToModel(userDBData, false)
	return updatedUserModel, http.StatusOK, nil
}

func DeleteUserByID(userID string) (int, error) {
	userDBData := UserDBData{Id: userID}
	if statusCode, err := userDBData.deleteByID(); err != nil {
		return statusCode, err
	}
	notifyUserDeletedEvent(userID)
	return http.StatusOK, nil
}
