package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun/schema"
)

const (
	UserTableName  = "user_data"
	UserTableAlias = "user"
)

type Role int

const (
	Role_Customer Role = iota
	Role_Manager
)

type Genders int

const (
	Genders_FEMALE = iota
	Genders_MALE
	Genders_OTHERS
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

type UserModel struct {
	Id          string  `json:"id" bun:"id"`
	FirstName   string  `json:"firstName" bun:"firstName"`
	LastName    string  `json:"lastName" bun:"lastName"`
	DateOfBirth string  `json:"dateOfBirth" bun:"dateOfBirth"`
	Gender      Genders `json:"gender" bun:"gender"`
	Address     Address `json:"address" bun:"address"`
	Phone       string  `json:"phone" bun:"phone"`
	Role        Role    `json:"role" bun:"role"`
	Email       string  `json:"email" bun:"email"`
	Password    string  `json:"password" bun:"password"`
	Verified    bool    `json:"verified" bun:"verified"`
}

type UserDBData struct {
	schema.BaseModel `bun:"table:user_data,alias:user"`
	Id               string            `json:"id" bun:"id,pk"`
	FirstName        string            `json:"firstName" bun:"firstName"`
	LastName         string            `json:"lastName" bun:"lastName"`
	DateOfBirth      string            `json:"dateOfBirth" bun:"dateOfBirth"`
	Gender           Genders           `json:"gender" bun:"gender"`
	Address          string            `json:"address" bun:"address"`
	Email            string            `json:"email" bun:"email,unique"`
	Phone            string            `json:"phone" bun:"phone"`
	Password         string            `json:"password" bun:"password"`
	Tags             map[string]string `json:"tags" bun:"tags"`
	Blob             []byte            `json:"blob" bun:"blob"`
	LastLogin        time.Time         `json:"lastLogin" bun:"lastLogin"`
	Verified         bool              `json:"verified" bun:"verified"`
	Role        Role    `json:"role" bun:"role"`
	LastSessionID    string            `json:"lastSessionID" bun:"lastSessionID"`
	CreatedAt        time.Time         `json:"createdAt"  bun:"createdAt" custom:"update_invalid"`
	UpdatedAt        time.Time         `json:"updatedAt" bun:"updatedAt"`
}

type UserDBOps interface {
	createUser() (int, error)
	login() (int, error)
	getByID() (int, error)
	getByEmail() (int, error)
	updateUser() (int, error)
	deleteByID() (int, error)
}

func (user *UserDBData) createUser() (int, error) {

	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Insert", UserTableName, err)
	}

	user.Id = uuid.New().String()
	user.Tags = map[string]string{}
	user.Blob = nil
	user.LastLogin = time.Time{}
	user.Verified = true
	user.LastSessionID = ""
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if _, err := dbClient.NewInsert().Model(user).Exec(context.TODO()); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Insert", UserTableName, err)
	}
	return http.StatusOK, nil

}

func (user *UserDBData) login() (int, error) {
	userGivenPassword := user.Password
	if statusCode, err := user.getByEmail(); err != nil {
		return statusCode, fmt.Errorf("exception while fetching user record by email. %v", err)
	}
	if userGivenPassword != user.Password {
		return http.StatusForbidden, fmt.Errorf("login exception. invalid password")
	}
	user.LastLogin = time.Now()
	sessionID, err := createNewSession(user)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("exception while creating session. %v", err)
	}
	user.LastSessionID = sessionID
	return user.updateUser()
}

func (user *UserDBData) getByID() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", UserTableName, err)
	}
	whereClausee := [] WhereClauseType{
		{
			ColumnName:   "id",
			RelationType: EQUAL,
			ColumnValue:  user.Id,
		},
	}
	userData, _, _, statusCode, err := readUtil(nil, whereClausee, nil, nil, nil, true)
	if err != nil {
		return statusCode, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", UserTableName, err)
	}
	copyUserDBData(user, &userData)
	return http.StatusOK, nil
}

func (user *UserDBData) getByEmail() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", UserTableName, err)
	}
	whereClausee := [] WhereClauseType{
		{
			ColumnName:   "email",
			RelationType: EQUAL,
			ColumnValue:  user.Email,
		},
	}
	userData, _, _, statusCode, err := readUtil(nil, whereClausee, nil, nil, nil, true)
	if err != nil {
		return statusCode, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", UserTableName, err)
	}
	copyUserDBData(user, &userData)
	return http.StatusOK, nil
}

func (user *UserDBData) updateUser() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Update", UserTableName, err)
	}
	user.UpdatedAt = time.Now()
	updateQuery := dbClient.NewUpdate().Model(user)
	updateQuery = updateQuery.WherePK()
	oldVersion := 0
	prepareUpdateQuery(updateQuery, &oldVersion, user, true, true)
	if _, err := updateQuery.Exec(context.TODO()); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Update", UserTableName, err)
	}
	return http.StatusOK, nil
}

func (user *UserDBData) deleteByID() (int, error) {
	if err := verifyDatabaseConnection(dbClient); err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Delete", UserTableName, err)
	}
	whereClauses := [] WhereClauseType{
		{
			ColumnName:   "id",
			RelationType: EQUAL,
			ColumnValue:  user.Id,
		},
	}

	deleteQuery := dbClient.NewDelete().
		Model(user)

	// prepare whereClause.
	queryStr, vals, err := createWhereClause(whereClauses)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Delete", UserTableName, err)
	}
	deleteQuery = deleteQuery.Where(queryStr, vals...)


	if _, err := deleteQuery.Exec(context.TODO()); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Delete", UserTableName, err)
	}
	return http.StatusOK, nil
}

func convertAddressToString(address Address) string {
	return fmt.Sprintf("%s;%s;%s;%s;%s;%s", address.AddressLine1, address.AddressLine2, address.City, address.State, address.Country, address.Pincode)
}

func convertStringToAddress(address string) Address {
	splitAddressString := strings.Split(address, ";")
	addressObj := Address{
		AddressLine1: splitAddressString[0],
		AddressLine2: splitAddressString[1],
		City:         splitAddressString[2],
		State:        splitAddressString[3],
		Country:      splitAddressString[4],
		Pincode:      splitAddressString[5],
	}
	return addressObj
}

func validateUserModel(userModel *UserModel, create bool) error {
	if userModel.Email == "" {
		return fmt.Errorf("invalid UserModel. Empty Email")
	}

	if userModel.FirstName == "" {
		return fmt.Errorf("invalid UserModel. Empty Firstname")
	}

	if userModel.LastName == "" {
		return fmt.Errorf("invalid UserModel. Empty LastName")
	}

	if userModel.DateOfBirth == "" {
		return fmt.Errorf("invalid UserModel. Empty DateOfBirth")
	}

	if userModel.Email == "" {
		return fmt.Errorf("invalid UserModel. Empty Email")
	}

	if _, err := mail.ParseAddress(userModel.Email); err != nil {
		return fmt.Errorf("invalid UserModel. Invalid Email %v", err)
	}

	if userModel.Phone == "" {
		return fmt.Errorf("invalid UserModel. Empty Phone")
	}
	re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
	if !re.MatchString(userModel.Phone) {
		return fmt.Errorf("invalid UserModel. Invalid Phonenumber %v", userModel.Phone)
	}

	if userModel.Password == "" {
		return fmt.Errorf("invalid UserModel. Empty Password")
	}
	if create {
		user := UserDBData{
			Email: userModel.Email,
		}
		if _, err := user.getByEmail(); err == nil {
			return fmt.Errorf("email already registered")
		}
	}
	return nil
}

func validateLoginModel(userModel *UserModel) error {
	if userModel.Email == "" {
		return fmt.Errorf("invalid UserModel. Empty Email")
	}

	user := UserDBData{
		Email: userModel.Email,
	}
	if _, err := user.getByEmail(); err != nil {
		return fmt.Errorf("email %s not found", userModel.Email)
	}

	if userModel.Password == "" {
		return fmt.Errorf("invalid UserModel. Empty Password")
	}
	return nil
}

func convertDBDataToModel(userDBData UserDBData, includePwd bool) UserModel {
	userModel := UserModel{
		Id:          userDBData.Id,
		FirstName:   userDBData.FirstName,
		LastName:    userDBData.LastName,
		DateOfBirth: userDBData.DateOfBirth,
		Gender:      userDBData.Gender,
		Address:     convertStringToAddress(userDBData.Address),
		Phone:       userDBData.Phone,
		Role:        userDBData.Role,
		Email:       userDBData.Email,
		Password:    "",
		Verified:    userDBData.Verified,
	}
	if includePwd {
		userModel.Password = userDBData.Password
	}
	return userModel
}

func convertModelToDBData(userModel UserModel) UserDBData {
	userDBData := UserDBData{
		Id:            userModel.Id,
		FirstName:     userModel.FirstName,
		LastName:      userModel.LastName,
		DateOfBirth:   userModel.DateOfBirth,
		Gender:        userModel.Gender,
		Address:       convertAddressToString(userModel.Address),
		Email:         userModel.Email,
		Phone:         userModel.Phone,
		Password:      userModel.Password,
		LastLogin:     time.Time{},
		Role:          userModel.Role,
	}
	return userDBData
}

func copyUserDBData(userDBData1, userDBData2 *UserDBData)  {

		userDBData1.Id =       userDBData2.Id
		userDBData1.FirstName=     userDBData2.FirstName
		userDBData1.LastName=      userDBData2.LastName
		userDBData1.DateOfBirth=   userDBData2.DateOfBirth
		userDBData1.Gender=        userDBData2.Gender
		userDBData1.Address=       userDBData2.Address
		userDBData1.Email=         userDBData2.Email
		userDBData1.Phone=         userDBData2.Phone
		userDBData1.Password=      userDBData2.Password
		userDBData1.Tags=          userDBData2.Tags
		userDBData1.Blob=          userDBData2.Blob
		userDBData1.LastLogin=     userDBData2.LastLogin
		userDBData1.Verified=      userDBData2.Verified
		userDBData1.Role=          userDBData2.Role
		userDBData1.LastSessionID= userDBData2.LastSessionID
		userDBData1.CreatedAt=     userDBData2.CreatedAt
		userDBData1.UpdatedAt=     userDBData2.UpdatedAt
}


