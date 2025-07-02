package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

var MySqlStringTypes = []string{
	"char", "varchar", "binary", "varbinery",
	"tinyblob", "tinytext", "text", "blob",
	"mediumtext", "mediumblob", "longtext", "longblob",
	"enum", "set"}
var MysqlIntTypes = []string{
	"tinyint", "smallint", "mediumint",
	"int", "integer", "bigint"}

type Table struct {
	Name       string
	PrimaryKey string
	Conn       *sql.DB
	Columns    []Column
}

type Column struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
}

type NotExistError struct {
}

func (e *NotExistError) Error() string {
	return "record not found"
}

type InvalidFieldError struct {
	Column string
}

func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("field %s have invalid type", e.Column)
}

type ReadFieldError struct {
	Field string
}

func (e *ReadFieldError) Error() string {
	return "can't read field " + e.Field
}

type QueryError struct {
	Err error
}

func (e *QueryError) Error() string {
	return "failed to exec query: " + e.Err.Error()
}

func (table *Table) fetchColumns() error {
	rows, err := table.Conn.Query("SHOW COLUMNS FROM `" + table.Name + "`;")
	defer rows.Close()

	if err != nil {
		return &QueryError{err}
	}

	for rows.Next() {
		column := Column{}
		err = rows.Scan(&column.Field, &column.Type, &column.Null, &column.Key, &column.Default, &column.Extra)
		if err != nil {
			return &QueryError{err}
		}
		table.Columns = append(table.Columns, column)
		if column.Key == "PRI" {
			table.PrimaryKey = column.Field
		}
	}

	return nil
}

func (table *Table) fetchRecord(id int64) (any, error) {
	item := make(map[string]any)
	values := make([]any, len(table.Columns))
	for i, _ := range values {
		values[i] = new(interface{})
	}

	q := "SELECT * FROM " + table.Name + " WHERE " + table.PrimaryKey + " = ?;"
	err := table.Conn.QueryRow(q, id).Scan(values...)
	if err == sql.ErrNoRows {
		return map[string]any{}, &NotExistError{}
	} else if err != nil {
		return map[string]any{}, &QueryError{err}
	}

	err = table.mapRowValues(values, item)
	if err != nil {
		return map[string]any{}, err
	}

	return map[string]any{"record": item}, nil
}

func (table *Table) listRecords(limit, offset int) (any, error) {
	q := "SELECT * FROM " + table.Name + " LIMIT ? OFFSET ?;"
	rows, err := table.Conn.Query(q, limit, offset)
	if err != nil {
		return []map[string]any{}, &QueryError{err}
	}

	result := make([]map[string]any, 0, 2)
	for rows.Next() {
		values := make([]any, len(table.Columns))
		item := make(map[string]any)
		for i, _ := range values {
			values[i] = new(interface{})
		}
		if err := rows.Scan(values...); err != nil {
			return []map[string]any{}, &QueryError{err}
		}

		err := table.mapRowValues(values, item)
		if err != nil {
			return []map[string]any{}, err
		}
		result = append(result, item)
	}
	rows.Close()

	return map[string]any{"records": result}, nil
}

func (table *Table) mapRowValues(values []any, item map[string]any) error {
	for i, f := range table.Columns {
		value := *values[i].(*interface{})
		switch {
		case value == nil:
			item[f.Field] = nil
		case f.isStringValue():
			if stringValue, ok := value.([]byte); ok {
				item[f.Field] = string(stringValue)
			} else {
				return &ReadFieldError{f.Field}
			}
		case f.isIntValue():
			if intValue, ok := value.(int64); ok {
				item[f.Field] = intValue
			} else {
				return &ReadFieldError{f.Field}
			}
		default:
			return &ReadFieldError{f.Field}
		}
	}
	return nil
}

func (table *Table) insertRow(byteParams []byte) (any, error) {
	columnNames := []string{}
	values := []any{}
	params := map[string]any{}
	err := json.Unmarshal(byteParams, &params)
	if err != nil {
		return map[string]int64{}, errors.New("can't parse params")
	}

	for _, f := range table.Columns {
		value, ok := params[f.Field]
		if !ok {
			if f.isStringValue() && !f.allowNull() {
				value = ""
			} else {
				continue
			}
		}

		if f.Field == table.PrimaryKey {
			continue
		}

		err = f.validateParamsValue(value)
		if err != nil {
			return map[string]int64{}, err
		}

		columnNames = append(columnNames, f.Field)
		values = append(values, value)
	}

	columnNamesQList := strings.Join(columnNames, ", ")
	paramsPlaces := strings.Join(slices.Repeat([]string{"?"}, len(values)), ", ")
	q := "INSERT INTO " + table.Name + "(" + columnNamesQList + ")" + " VALUES (" + paramsPlaces + ");"

	res, err := table.Conn.Exec(q, values...)
	if err != nil {
		return map[string]int64{}, &QueryError{err}
	}
	id, err := res.LastInsertId()
	if err != nil {
		return map[string]int64{}, &QueryError{err}
	}

	idRes := map[string]int64{table.PrimaryKey: id}
	return idRes, nil
}

func (table *Table) updateRecord(id int, byteParams []byte) (any, error) {
	exists := new(any)
	q := "SELECT '' FROM " + table.Name + " WHERE " + table.PrimaryKey + " = ?;"
	err := table.Conn.QueryRow(q, id).Scan(exists)
	if err != nil {
		return map[string]int64{}, err
	}

	columnNames := []string{}
	values := []any{}
	params := map[string]any{}
	err = json.Unmarshal(byteParams, &params)
	if err != nil {
		return []map[string]any{}, errors.New("can't parse params")
	}
	for _, f := range table.Columns {
		value, ok := params[f.Field]
		if !ok {
			continue
		}

		if f.Field == table.PrimaryKey {
			return map[string]int64{}, &InvalidFieldError{f.Field}
		}

		err = f.validateParamsValue(value)
		if err != nil {
			return map[string]int64{}, err
		}

		columnNames = append(columnNames, f.Field)
		values = append(values, value)
	}

	setExp := []string{}
	for _, c := range columnNames {
		setExp = append(setExp, c+"=?")
	}
	fullSetExp := strings.Join(setExp, ", ")

	updateQueryString := "UPDATE " + table.Name + " SET " + fullSetExp + " WHERE " + table.PrimaryKey + " = ?;"
	values = append(values, id)
	res, err := table.Conn.Exec(updateQueryString, values...)
	if err != nil {
		return map[string]int64{}, &QueryError{err}
	}
	rowUpdated, err := res.RowsAffected()
	if err != nil {
		return map[string]int64{}, &QueryError{err}
	}

	return map[string]int64{"updated": rowUpdated}, nil
}

func (table *Table) deleteRecord(id int) (any, error) {
	deleteQuery := "DELETE FROM " + table.Name + " WHERE id = ?;"
	res, err := table.Conn.Exec(deleteQuery, id)
	if err != nil {
		return map[string]int64{}, &QueryError{err}
	}

	rowsDeleted, err := res.RowsAffected()
	if err != nil {
		return map[string]int64{}, &QueryError{err}
	}

	return map[string]int64{"deleted": rowsDeleted}, nil
}

func (c *Column) allowNull() bool {
	return c.Null == "YES"
}

func (c *Column) isIntValue() bool {
	typeName := strings.Split(c.Type, "(")[0]
	return slices.Contains(MysqlIntTypes, typeName)
}

func (c *Column) isStringValue() bool {
	typeName := strings.Split(c.Type, "(")[0]
	return slices.Contains(MySqlStringTypes, typeName)
}

func (c *Column) validateParamsValue(value any) error {
	if c.isStringValue() {
		switch value.(type) {
		case string:
			return nil
		case nil:
			if c.Null == "NO" {
				return &InvalidFieldError{c.Field}
			}
		default:
			return &InvalidFieldError{c.Field}
		}
	}

	if c.isIntValue() {
		switch value.(type) {
		case int, int8, int16, int32, int64:
			return nil
		case nil:
			if !c.allowNull() {
				return &InvalidFieldError{c.Field}
			}
		default:
			return &InvalidFieldError{c.Field}
		}
	}

	return nil
}

type DbExplorer struct {
	DB     *sql.DB
	Tables []*Table
}

func (dbEx *DbExplorer) fetchAllTables() error {
	rows, err := dbEx.DB.Query("SHOW TABLES;")
	if err != nil {
		return &QueryError{err}
	}

	for rows.Next() {
		table := Table{Conn: dbEx.DB}
		err = rows.Scan(&table.Name)
		if err != nil {
			return &QueryError{err}
		}
		dbEx.Tables = append(dbEx.Tables, &table)
	}
	rows.Close()

	return nil
}

func (dbEx *DbExplorer) fetchColumns() error {
	for _, t := range dbEx.Tables {
		err := t.fetchColumns()
		if err != nil {
			return err
		}
	}

	return nil
}

func (dbEx *DbExplorer) tableList() []string {
	res := make([]string, 0, 2)
	for _, t := range dbEx.Tables {
		res = append(res, t.Name)
	}
	return res
}

func (dbEx *DbExplorer) findTable(name string) *Table {
	var table *Table
	table = nil
	for _, t := range dbEx.Tables {
		if t.Name == name {
			table = t
			break
		}
	}
	return table
}

type ApiError struct {
	Error string `json:"error"`
}

type ApiResponse struct {
	Response any `json:"response"`
}

func responseWithError(w http.ResponseWriter, errStatus int, errText string) {
	w.WriteHeader(errStatus)
	res, err := json.Marshal(ApiError{Error: errText})
	if err != nil {
		panic(err)
	}
	w.Write(res)
}

func (dbEx *DbExplorer) listTablesHandler(w http.ResponseWriter, r *http.Request) {
	tables := map[string][]string{}
	tablesList := dbEx.tableList()
	tables["tables"] = append(tables["tables"], tablesList...)

	res, err := json.Marshal(ApiResponse{Response: tables})
	if err != nil {
		panic(err)
	}
	w.Write(res)
}

func (dbEx *DbExplorer) listRecordsHandler(w http.ResponseWriter, r *http.Request) {
	tableName := strings.Split(strings.Split(r.URL.Path[1:], "/")[0], "?")[0]
	table := dbEx.findTable(tableName)
	if table == nil {
		responseWithError(w, http.StatusNotFound, "unknown table")
		return
	}

	limit, offset := 5, 0
	if limitParam := r.FormValue("limit"); limitParam != "" {
		if limitInt, err := strconv.Atoi(limitParam); err == nil {
			limit = limitInt
		}
	}
	if offsetParam := r.FormValue("offset"); offsetParam != "" {
		if offsetInt, err := strconv.Atoi(offsetParam); err == nil {
			offset = offsetInt
		}
	}

	res, err := table.listRecords(limit, offset)
	if err != nil {
		responseWithError(w, http.StatusUnprocessableEntity, "failed to list table")
		return
	}
	resJSON, err := json.Marshal(ApiResponse{Response: res})
	if err != nil {
		panic(err)
	}
	w.Write(resJSON)
}

func (dbEx *DbExplorer) fetchRecordHandler(w http.ResponseWriter, r *http.Request) {
	tableAndId := strings.Split(r.URL.Path[1:], "/")
	tableName := tableAndId[0]
	table := dbEx.findTable(tableName)
	if table == nil {
		responseWithError(w, http.StatusNotFound, "unknow table")
	}

	strID := tableAndId[1]
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "id must be int")
		return
	}

	res, err := table.fetchRecord(id)
	if errNotEx, ok := err.(*NotExistError); ok {
		responseWithError(w, http.StatusNotFound, errNotEx.Error())
		return
	} else if err != nil {
		responseWithError(w, http.StatusUnprocessableEntity, "failed to list table")
		return
	}

	resJSON, err := json.Marshal(ApiResponse{Response: res})
	if err != nil {
		panic(err)
	}
	w.Write(resJSON)
}

func (dbEx *DbExplorer) createHandler(w http.ResponseWriter, r *http.Request) {
	tableName := strings.Split(r.URL.Path[1:], "/")[0]
	table := dbEx.findTable(tableName)
	if table == nil {
		responseWithError(w, http.StatusNotFound, "unknown table")
		return
	}

	buff := []byte{}
	buff, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	res, err := table.insertRow(buff)
	switch err.(type) {
	case nil:
		resJSON, err := json.Marshal(ApiResponse{Response: res})
		if err != nil {
			panic(err)
		}
		w.Write(resJSON)
	case *InvalidFieldError:
		fmt.Println(err.Error())
		responseWithError(w, http.StatusBadRequest, err.Error())
	default:
		fmt.Println(err.Error())
		responseWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func (dbEx *DbExplorer) updateHandler(w http.ResponseWriter, r *http.Request) {
	tableAndID := strings.Split(r.URL.Path[1:], "/")
	tableName := tableAndID[0]
	table := dbEx.findTable(tableName)
	if table == nil {
		responseWithError(w, http.StatusNotFound, "unknown table")
		return
	}
	id, err := strconv.Atoi(tableAndID[1])
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	paramsBuff := []byte{}
	paramsBuff, err = io.ReadAll(r.Body)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "invalid params")
		return
	}

	res, err := table.updateRecord(id, paramsBuff)
	switch err.(type) {
	case nil:
		resJSON, err := json.Marshal(ApiResponse{Response: res})
		if err != nil {
			panic(err)
		}
		w.Write(resJSON)
	case *NotExistError:
		responseWithError(w, http.StatusNotFound, err.Error())
	case *InvalidFieldError:
		responseWithError(w, http.StatusBadRequest, err.Error())
	default:
		responseWithError(w, http.StatusInternalServerError, err.Error())
	}
}

func (dbEx *DbExplorer) deleteHandler(w http.ResponseWriter, r *http.Request) {
	tableAndID := strings.Split(r.URL.Path[1:], "/")
	tableName := tableAndID[0]
	table := dbEx.findTable(tableName)
	if table == nil {
		responseWithError(w, http.StatusNotFound, "unknown table")
		return
	}
	id, err := strconv.Atoi(tableAndID[1])
	if err != nil {
		responseWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	res, err := table.deleteRecord(id)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resJSON, err := json.Marshal(ApiResponse{Response: res})
	if err != nil {
		panic(err)
	}
	w.Write(resJSON)
}

func (dbEx *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/" {
			dbEx.listTablesHandler(w, r)
		} else {
			tableAndId := strings.Split(r.URL.Path[1:], "/")
			if len(tableAndId) > 1 {
				dbEx.fetchRecordHandler(w, r)
			} else {
				dbEx.listRecordsHandler(w, r)
			}
		}
	case http.MethodPut:
		dbEx.createHandler(w, r)
	case http.MethodPost:
		dbEx.updateHandler(w, r)
	case http.MethodDelete:
		dbEx.deleteHandler(w, r)
	default:
		responseWithError(w, http.StatusNotAcceptable, "not acceptable")
	}
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	dbEx := &DbExplorer{}
	dbEx.DB = db
	err := dbEx.fetchAllTables()
	if err != nil {
		return nil, err
	}
	err = dbEx.fetchColumns()
	if err != nil {
		return nil, err
	}

	return dbEx, nil
}
