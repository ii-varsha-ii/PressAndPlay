package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"

	_ "github.com/lib/pq"
	"github.com/oiime/logrusbun"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	POSTGRES_HOST_ENV     = "POSTGRES_HOST"
	POSTGRES_PORT_ENV     = "POSTGRES_PORT"
	POSTGRES_USERNAME_ENV = "POSTGRES_USERNAME"
	POSTGRES_PASSWORD_ENV = "POSTGRES_PASSWORD"
	POSTGRES_DB_ENV       = "POSTGRES_DB"

	DEFAULT_POOL_SIZE      = 8   // default connection pool size.
	DEFAULT_IDLE_TIMEOUTS  = -1  // never timeout/close an idle connection.
	DEFAULT_LOG_SLOW_QUERY = 100 // log db queries slower than 100ms by default

	SCHEMA_NAME    = "user_schema"
	INDEX_TYPE_GIN = "gin"
)

const (
	// NULL_VALUE To check whether a value is NULL in postgres
	NULL_VALUE = "null"
	// NOT_NULL_VALUE To check if a value is not NULL in postgres
	NOT_NULL_VALUE = "not null"
)

type RelationType int

const (
	NONE RelationType = iota
	EQUAL
	NOT_EQUAL
	IN
	NOT_IN
	IS
	LIKE
	ANY
)

func (r RelationType) String() string {
	switch r {
	case EQUAL:
		return "="
	case NOT_EQUAL:
		return "!="
	case IN:
		return "in"
	case NOT_IN:
		return "not in"
	case IS:
		return "is"
	case LIKE:
		return "like"
	case ANY:
		return "any"
	default:
		return ""
	}
}

type Cursor struct {
	PageNum      int
	PageSize     int
	TotalPages   uint32
	TotalRecords uint32
	PageToken    string
	OrderBy      string
	Limit        int
	Offset       int
}

type IndexParams struct {
	Name        string
	Type        string
	TableName   string
	ColumnNames []string
}

type WhereClauseType struct {
	ColumnName   string
	RelationType RelationType
	ColumnValue  interface{}
	TableAlias   string
	JsonOperator string
}

func newPostgresClient() (*bun.DB, error) {
	host := common.GetEnv(POSTGRES_HOST_ENV, "localhost")
	port := common.GetEnv(POSTGRES_PORT_ENV, "5432")
	username := common.GetEnv(POSTGRES_USERNAME_ENV, "admin")
	password := common.GetEnv(POSTGRES_PASSWORD_ENV, "admin")
	dbName := common.GetEnv(POSTGRES_DB_ENV, "pressandplay")
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(fmt.Sprintf("%s:%s", host, port)),
		pgdriver.WithUser(username),
		pgdriver.WithPassword(password),
		pgdriver.WithDatabase(dbName),
		pgdriver.WithApplicationName(SERVICE_NAME),
		pgdriver.WithConnParams(map[string]interface{}{
			"search_path": SCHEMA_NAME,
		}),
		pgdriver.WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		})))
	sqldb.SetConnMaxIdleTime(DEFAULT_IDLE_TIMEOUTS)
	sqldb.SetMaxIdleConns(DEFAULT_POOL_SIZE)
	sqldb.SetMaxOpenConns(DEFAULT_POOL_SIZE)
	sqlDB, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbName))
	if err != nil {
		return nil, err
	}
	db := bun.NewDB(sqlDB, pgdialect.New())
	logrusObj := logrus.New()
	logrusObj.SetFormatter(&logrus.TextFormatter{DisableQuote: true})
	db.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{
		Logger:          logrusObj,
		LogSlow:         DEFAULT_LOG_SLOW_QUERY,
		QueryLevel:      logrus.DebugLevel,
		SlowLevel:       logrus.WarnLevel,
		ErrorLevel:      logrus.ErrorLevel,
		MessageTemplate: "{{.Operation}}[{{.Duration}}]: {{.Query}}",
		ErrorTemplate:   "{{.Operation}}[{{.Duration}}]: {{.Query}}: {{.Error}}",
	}))
	return db, db.Ping()
}

func createUserTable() error {
	if dbClient == nil {
		err := fmt.Errorf("exception while creating user table. database connection not created")
		logrus.Errorf(err.Error())
		return err
	}

	_, err := dbClient.ExecContext(context.TODO(), "CREATE SCHEMA IF NOT EXISTS ?", bun.Ident(SCHEMA_NAME))
	if err != nil {
		err := fmt.Errorf("exception while creating user schema %v. %v", SCHEMA_NAME, err)
		logrus.Errorf(err.Error())
		return err
	}

	_, err = dbClient.ExecContext(context.TODO(), "CREATE EXTENSION IF NOT EXISTS pg_trgm")
	if err != nil {
		err := fmt.Errorf("exception while creating pg_trgm extention. %v", err)
		logrus.Errorf(err.Error())
		return err
	}

	_, err = dbClient.ExecContext(context.TODO(), "CREATE EXTENSION IF NOT EXISTS btree_gin")
	if err != nil {
		err := fmt.Errorf("exception while creating btree_gin extention. %v", err)
		logrus.Errorf(err.Error())
		return err
	}

	tableSchemaPtr := reflect.New(reflect.TypeOf(UserDBData{}))
	createTableQuery := dbClient.NewCreateTable().
		Model(tableSchemaPtr.Interface()).
		IfNotExists()

	_, err = createTableQuery.Exec(context.TODO())
	if err != nil {
		err := fmt.Errorf("exception while creaiting user table. %v", err)
		logrus.Errorf(err.Error())
		return err
	}

	// create indexes
	emailIndex := IndexParams{
		Name:        "idx_user_email",
		Type:        INDEX_TYPE_GIN,
		TableName:   UserTableName,
		ColumnNames: []string{"email"},
	}
	emailIndexCreateQuery := createIndexCreateQuery(emailIndex)
	_, err = dbClient.ExecContext(context.TODO(), emailIndexCreateQuery)
	if err != nil {
		err := fmt.Errorf("exception while creaiting index in email column of user table. %v", err)
		logrus.Errorf(err.Error())
		return err
	}
	return nil
}

func createIndexCreateQuery(indexCreateParam IndexParams) string {
	var buffer bytes.Buffer

	if strings.ToLower(indexCreateParam.Type) == "unique" {
		buffer.WriteString("create unique index if not exists ")
	} else {
		buffer.WriteString("create index if not exists ")
	}
	buffer.WriteString(indexCreateParam.Name)
	buffer.WriteString(" on ")

	if indexCreateParam.TableName == "" {
		buffer.WriteString(UserTableName)
	} else {
		buffer.WriteString(indexCreateParam.TableName)
	}
	if strings.ToLower(indexCreateParam.Type) != "unique" && indexCreateParam.Type != "" {
		buffer.WriteString(" using ")
		buffer.WriteString(indexCreateParam.Type)
	}
	buffer.WriteString(" ")
	buffer.WriteString("(")
	prepColumnList := strings.Join(indexCreateParam.ColumnNames, ",")
	buffer.WriteString(prepColumnList)
	buffer.WriteString(")")

	return buffer.String()
}

func verifyDatabaseConnection(databaseConnection *bun.DB) error {
	if databaseConnection == nil {
		return fmt.Errorf("database connection not initialized")
	}
	return nil
}

func createWhereClause(whereClause []WhereClauseType) (string, []interface{}, error) {
	var values []interface{}
	var buffer bytes.Buffer

	flag := true
	for i := 0; i < len(whereClause); i++ {
		val := whereClause[i]

		if flag {
			flag = false
		} else {
			buffer.WriteString(" and ")
		}
		var relType string
		if val.RelationType.String() != "" {
			relType = val.RelationType.String()
		} else {
			relType = "="
		}

		// useful column name, so it works in where clauses with joins too
		var fullColumnName string             // placeholder for column name
		fullColumnNameValues := []bun.Ident{} // values of placeholder
		if val.TableAlias == "" {
			fullColumnName = "?TableAlias.?" // ?TableAlias is filled by the bun ORM
		} else {
			fullColumnName = "?.?"
			fullColumnNameValues = append(fullColumnNameValues, bun.Ident(val.TableAlias))
		}
		fullColumnNameValues = append(fullColumnNameValues, bun.Ident(strings.ToLower(val.ColumnName)))

		switch strings.ToLower(relType) {
		case "like":
			buffer.WriteString(fullColumnName)
			buffer.WriteString(val.JsonOperator)
			buffer.WriteString(" ")
			buffer.WriteString(relType)
			buffer.WriteString(" ")
			buffer.WriteString("?")
			colValue, ok := val.ColumnValue.(string)
			if !ok {
				return "", nil, fmt.Errorf("exception while creating where query for tabel %s. Column value not string type", fullColumnName)
			}
			values = append(values, fullColumnNameValues[0])
			if len(fullColumnNameValues) > 1 {
				values = append(values, fullColumnNameValues[1])
			}
			values = append(values, "%"+colValue+"%")
		case "in":
			buffer.WriteString(fullColumnName)
			buffer.WriteString(val.JsonOperator)
			buffer.WriteString(" ")
			buffer.WriteString(relType)
			buffer.WriteString(" ")
			buffer.WriteString("(" + "?" + ")")
			values = append(values, fullColumnNameValues[0])
			if len(fullColumnNameValues) > 1 {
				values = append(values, fullColumnNameValues[1])
			}
			values = append(values, val.ColumnValue)
		case "is":
			buffer.WriteString(fullColumnName)
			buffer.WriteString(val.JsonOperator)
			buffer.WriteString(" ")
			buffer.WriteString(relType)
			buffer.WriteString(" ")
			colValue, ok := val.ColumnValue.(string)
			if !ok {
				return "", nil, fmt.Errorf("exception while creating where query for tabel %s. Column value not string type", fullColumnName)
			}
			if colValue == NULL_VALUE || colValue == NOT_NULL_VALUE {
				buffer.WriteString(colValue)
			} else {
				return "", nil, fmt.Errorf("only null and not null values are supported")
			}
			values = append(values, fullColumnNameValues[0])
			if len(fullColumnNameValues) > 1 {
				values = append(values, fullColumnNameValues[1])
			}
		case "any":
			buffer.WriteString("'" + val.ColumnValue.(string) + "'")
			buffer.WriteString(" = ")
			buffer.WriteString(relType)
			buffer.WriteString("(" + fullColumnName + val.JsonOperator + ")")
			values = append(values, fullColumnNameValues[0])
			if len(fullColumnNameValues) > 1 {
				values = append(values, fullColumnNameValues[1])
			}
		default:
			switch val.ColumnValue.(type) {
			case int8, uint8, int16, uint16, int32, uint32, int64, int, uint, uint64, float32, float64:
				buffer.WriteString(fullColumnName)
				buffer.WriteString(val.JsonOperator)
				buffer.WriteString(" ")
				buffer.WriteString(relType)
				buffer.WriteString(" ")

				valPlaceholder := "'" + "?" + "'"
				buffer.WriteString(valPlaceholder)
				values = append(values, fullColumnNameValues[0])
				if len(fullColumnNameValues) > 1 {
					values = append(values, fullColumnNameValues[1])
				}
				values = append(values, val.ColumnValue)

			case string, bool:
				buffer.WriteString(fullColumnName)
				buffer.Write([]byte(val.JsonOperator))
				buffer.WriteString(" ")
				buffer.WriteString(relType)
				buffer.WriteString(" ")
				buffer.WriteString("?")

				values = append(values, fullColumnNameValues[0])
				if len(fullColumnNameValues) > 1 {
					values = append(values, fullColumnNameValues[1])
				}
				values = append(values, val.ColumnValue)
			default:
				buffer.WriteString(fullColumnName)
				buffer.WriteString(val.JsonOperator)
				buffer.WriteString(" ")
				buffer.WriteString(relType)
				buffer.WriteString(" ")
				buffer.WriteString("?")
				values = append(values, fullColumnNameValues[0])
				if len(fullColumnNameValues) > 1 {
					values = append(values, fullColumnNameValues[1])
				}
				values = append(values, val.ColumnValue)
			}
		}
	}
	return buffer.String(), values, nil
}

func prepareUpdateQuery(q *bun.UpdateQuery, oldVersion *int, data interface{}, igVersionCheck, colListEmpty bool) { // nolint
	v := reflect.ValueOf(data).Elem()
	for i := 0; i < v.NumField(); i++ {
		valueField := v.Field(i)
		typeField := v.Type().Field(i)
		kindType := valueField.Kind()
		if kindType == reflect.Ptr || typeField.Type.String() == "schema.BaseModel" {
			continue
		}
		pgTag := typeField.Tag.Get("bun")
		customPgTag := typeField.Tag.Get("custom")
		// Expecting first tag as columnName
		columnName := strings.Split(pgTag, ",")[0]
		if valueField.CanSet() {
			if !strings.Contains(customPgTag, "update_invalid") && colListEmpty {
				q = q.Column(strings.TrimSpace(columnName))
			}
			if !igVersionCheck {
				if typeField.Name == "Version" && kindType == reflect.Int {
					oval := valueField.Interface().(int)
					oldVersion = &oval
					newVersionInt64 := int64(*oldVersion + 1)
					valueField.SetInt(newVersionInt64)
					q = q.Where("Version = ?", *oldVersion)
				}
			}
		}
	}
}

func createColumnList(columnList []string) string {
	var buffer bytes.Buffer
	if columnList != nil {
		gblen := len(columnList)
		if gblen >= 1 {
			flag := true
			for i := 0; i < gblen; i++ {
				if flag {
					flag = false
				} else {
					buffer.WriteString(", ")
				}
				buffer.WriteString(strings.ToLower(columnList[i]))
			}
		}
	}
	return buffer.String()
}

func createOrderBy(orderByClause []string) (string, error) {
	var buffer bytes.Buffer
	if orderByClause != nil {
		oblen := len(orderByClause)
		if oblen >= 1 {
			flag := true
			for _, value := range orderByClause {
				kv := strings.Split(value, ":")
				if len(kv) != 2 {
					return "", fmt.Errorf("invalid orderBy param %v, it should be of type fieldName:sortingType", value)
				}
				if flag {
					flag = false
				} else {
					buffer.WriteString(", ")
				}
				buffer.WriteString(strings.TrimSpace(strings.ToLower(kv[0])))
				buffer.WriteString(" ")
				buffer.WriteString(strings.TrimSpace(kv[1]))
			}
		}
	}
	return buffer.String(), nil
}

func readUtil(pagination *Cursor, whereClausefilters []WhereClauseType,
	orderByClause []string, groupByClause, selectedColumns []string, singleRecord bool) (UserDBData, []UserDBData, *Cursor, int, error) {

	if err := verifyDatabaseConnection(dbClient); err != nil {
		return UserDBData{}, nil, nil, http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", UserTableName, err)
	}
	var (
		singleUser    = UserDBData{}
		listUser      []UserDBData
		newOffset     int
		newPagination = new(Cursor)
		readQuery     *bun.SelectQuery
	)
	if singleRecord {
		readQuery = dbClient.NewSelect().
			Model(&singleUser)
	} else {
		readQuery = dbClient.NewSelect().
			Model(&listUser)
	}

	colListStr := createColumnList(selectedColumns)
	if colListStr != "" {
		readQuery = readQuery.ColumnExpr(colListStr)
	}

	if len(whereClausefilters) != 0 {
		queryStr, vals, err := createWhereClause(whereClausefilters)
		if err != nil {
			return UserDBData{}, nil, nil, http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", UserTableName, err)
		}
		readQuery = readQuery.Where(queryStr, vals...)
	}

	groupCols := createColumnList(groupByClause)
	if groupCols != "" {
		readQuery = readQuery.GroupExpr(groupCols)
	}

	orderByCols, err := createOrderBy(orderByClause)
	if err != nil {
		return UserDBData{}, nil, nil, http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. %v", "Read", UserTableName, err)
	}
	if orderByCols != "" {
		readQuery = readQuery.OrderExpr(orderByCols)
	}

	if !singleRecord && pagination != nil {
		if pagination.PageSize <= 0 {
			return UserDBData{}, nil, nil, http.StatusBadRequest, fmt.Errorf("unable to Perform %s Operation on Table: %s. Invalid PazeSize %v", "Read",
				UserTableName, pagination.PageSize)
		}

		if pagination.PageSize != 0 {
			var offset int
			var cErr error
			if pagination.PageToken == "" {
				offset = 0
				if pagination.PageNum > 0 {
					offset = (pagination.PageNum - 1) * pagination.PageSize
				}
			} else {
				offset, cErr = strconv.Atoi(pagination.PageToken)
				if cErr != nil {
					return UserDBData{}, nil, nil, http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. Exception while converting pagetoken to offset. %v", "Read",
						UserTableName, cErr)
				}
			}
			readQuery = readQuery.Limit(pagination.PageSize).Offset(offset)
			newOffset = pagination.PageSize + offset
			newPagination.PageNum = offset/pagination.PageSize + 1
			newPagination.PageSize = pagination.PageSize
			newPagination.PageToken = strconv.Itoa(newOffset)
		}
	}

	if count, err := readQuery.ScanAndCount(context.TODO()); err != nil {
		return UserDBData{}, nil, nil, http.StatusInternalServerError, fmt.Errorf("unable to Perform %s Operation on Table: %s. Exception while reading data. %v", "Read",
			UserTableName, err)
	} else if !singleRecord && pagination != nil {
		if newOffset >= count {
			newPagination.PageToken = ""
		}
		newPagination.TotalRecords = uint32(count)
		if newPagination.PageSize > 0 {
			if count%newPagination.PageSize != 0 {
				newPagination.TotalPages = uint32(count/newPagination.PageSize) + 1
			} else {
				newPagination.TotalPages = uint32(count / newPagination.PageSize)
			}
		}
		return singleUser, listUser, newPagination, http.StatusOK, nil
	}

	return singleUser, listUser, nil, http.StatusOK, nil
}
