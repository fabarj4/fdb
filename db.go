package fdb

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type QueryExecer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Database struct {
	Dbname    string
	Dropfirst bool
}

func (d Database) Init(db *sql.DB) error {
	// check apakah database sudah dibuat
	exists := false
	query := fmt.Sprintf("select exists(SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower('%s'));", d.Dbname)
	if err := db.QueryRow(query).Scan(&exists); err != nil {
		return err
	}
	if d.Dropfirst && exists {
		query := fmt.Sprintf("DROP DATABASE %s", d.Dbname)
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	query = fmt.Sprintf("CREATE DATABASE %s", d.Dbname)
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil
}

type Table struct {
	Name            string
	PrimaryKey      string
	DstPrimary      interface{}
	DstPrimaryIndex int
	Fields          []string
	DstFields       []interface{}
	AutoIncrement   bool
	Data            interface{}
	ReturningID     bool
	Schema          string
}

//Connect : fungsi ini digunakan untuk melakukan koneksi dengan database
func Connect(user, password, dbName string) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbName)
	db, err := sql.Open("postgres", connStr)
	return db, err
}

//Create : fungsi ini digunakan untuk membuat tabel
func (t *Table) Create(db QueryExecer, item interface{}) error {
	queryFields, err := t.getQueryCreate(item)
	if err != nil {
		return err
	}
	if t.Name == "" {
		t.Name = strings.ToLower(reflect.TypeOf(item).Elem().Name())
	}
	if t.Schema != "" {
		query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", t.Schema)
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	query := fmt.Sprintf("CREATE TABLE %s(%s);", t.tableName(), strings.Join(queryFields, ","))
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil
}

//Insert : fungsi ini digunakan untuk memasukan data ke Table
func (t *Table) Insert(db QueryExecer, item interface{}) error {
	if err := t.setup(item, false, true); err != nil {
		return err
	}
	if t.AutoIncrement && t.ReturningID {
		if err := t.setup(item, true, true); err != nil {
			return err
		}
		query := fmt.Sprintf("INSERT INTO %s(%s) VALUES %s RETURNING %s", t.tableName(), strings.Join(t.getFieldWithoutPrimary(), ","), FieldsToVariables(t.Fields, true), t.PrimaryKey)
		if err := db.QueryRow(query, t.DstFields[0:len(t.DstFields)-1]...).Scan(t.DstFields[len(t.DstFields)-1]); err != nil {
			return err
		}
		return nil
	}

	query := fmt.Sprintf("INSERT INTO %s VALUES %s", t.tableName(), FieldsToVariables(t.Fields, false))
	_, err := db.Query(query, t.DstFields...)
	return err
}

//Delete : fungsi ini digunakan untuk menghapus data ke Table
func (t *Table) Delete(db QueryExecer, item interface{}) error {
	if err := t.setup(item, false, false); err != nil {
		return err
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", t.tableName(), t.PrimaryKey)
	_, err := db.Exec(query, t.DstPrimary)
	return err
}

//Update : fungsi ini digunakan untuk mengubah data ke Table
func (t *Table) Update(db QueryExecer, item interface{}, data map[string]interface{}) error {
	if err := t.setup(item, false, false); err != nil {
		return err
	}
	var kolom = []string{}
	var args []interface{}
	args = append(args, t.DstFields[t.DstPrimaryIndex])
	i := 2
	for key, value := range data {
		updateData := fmt.Sprintf("%v = $%d", strings.ToLower(key), i)
		kolom = append(kolom, updateData)
		args = append(args, value)
		i++
	}
	dataUpdate := strings.Join(kolom, " ,")
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $1", t.tableName(), dataUpdate, t.PrimaryKey)
	_, err := db.Exec(query, args...)
	return err
}

//Get : fungsi ini digunakan untuk mengambil data berdasarkan primary key
func (t *Table) Get(db QueryExecer, item interface{}) error {
	if err := t.setup(item, false, false); err != nil {
		return err
	}
	query := fmt.Sprintf("SELECT * FROM %v WHERE %v = $1 ", t.tableName(), t.PrimaryKey)
	err := db.QueryRow(query, t.DstFields[t.DstPrimaryIndex]).Scan(t.DstFields...)
	if err != nil {
		return err
	}
	return nil
}

//Gets : fungsi ini digunakan untuk mengambil data seluruh tabel
func (t *Table) Gets(db QueryExecer, item interface{}, c *Cursor) ([]interface{}, string, error) {
	var kolom = []string{}
	var args []interface{}
	var addOnsQuery []string
	var resultCursor string
	defultSort := fmt.Sprintf(" ORDER BY %s ASC", t.PrimaryKey)
	// var cursorData []string

	// var filter, sort, offset, limit string
	if c != nil {
		if c.Filters != "" {
			dataParams := strings.Split(c.Filters, ";")
			for i, v := range dataParams {
				temp := strings.Split(v, ",")
				where := fmt.Sprintf("%s %s $%d", strings.ToLower(temp[0]), temp[1], i+1)
				kolom = append(kolom, where)
				// arg, _ := url.QueryUnescape(temp[2])
				args = append(args, temp[2])
			}
			addOnsQuery = append(addOnsQuery, fmt.Sprintf(" WHERE %s", strings.Join(kolom, " AND ")))
		}
		if c.Sort != "" {
			dataSort := strings.Split(c.Sort, ";")
			temp := []string{}
			for _, v := range dataSort {
				temp = append(temp, v)
			}
			addOnsQuery = append(addOnsQuery, fmt.Sprintf(" ORDER BY %s", strings.Join(temp, ",")))
			defultSort = ""
		} else {
			addOnsQuery = append(addOnsQuery, fmt.Sprintf(" ORDER BY %s ASC", t.PrimaryKey))
			defultSort = ""
		}
		if c.Limit != "" {
			limitInt, err := strconv.Atoi(c.Limit)
			if err != nil {
				return nil, "", err
			}
			if c.Offset == "" {
				c.Offset = "0"
			}
			offsetInt, err := strconv.Atoi(c.Offset)
			addOnsQuery = append(addOnsQuery, fmt.Sprintf(" LIMIT %s OFFSET %s", c.Limit, c.Offset))
			if err != nil {
				return nil, "", err
			}
			offsetInt += limitInt
			c.Offset = fmt.Sprintf("%v", offsetInt)
		}
		resultCursor = c.SetCursor()
	}
	var query string
	if defultSort != "" {
		query = fmt.Sprintf("SELECT * FROM %s %s", t.tableName(), defultSort)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s %s", t.tableName(), strings.Join(addOnsQuery, " "))
	}
	data, err := db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}
	defer data.Close()
	var result []interface{}

	for data.Next() {
		temp := clone(item)
		if err := t.setup(temp, false, false); err != nil {
			return nil, "", err
		}
		if err = data.Scan(t.DstFields...); err != nil {
			return nil, "", err
		}
		result = append(result, temp)
	}

	if err = data.Err(); err != nil {
		return nil, "", err
	}

	return result, resultCursor, nil
}

// Clone : fungsi ini untuk menduplikat variable dengan alamat memori yang berbeda
func clone(data interface{}) interface{} {
	result := reflect.New(reflect.TypeOf(data).Elem())
	val := reflect.ValueOf(data).Elem()
	resultVal := result.Elem()
	for i := 0; i < val.NumField(); i++ {
		resultField := resultVal.Field(i)
		resultField.Set(val.Field(i))
	}
	return result.Interface()
}

//setup : fungsi ini digunakan untuk mengambil data fields,primaryKey, nama dari struct yang didaftarkan
func (t *Table) setup(item interface{}, primaryNotInclude, validateCheck bool) error {
	if t.Name == "" {
		t.Name = strings.ToLower(reflect.TypeOf(item).Elem().Name())
	}
	var result []interface{}
	var fields []string

	reflectValue := reflect.ValueOf(item)
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	var reflectType = reflectValue.Type()
	for i := 0; i < reflectValue.NumField(); i++ {
		rawTags := reflectType.Field(i).Tag.Get("fdb")
		if rawTags == "-" {
			continue
		} else if rawTags != "" {
			tags := strings.Split(rawTags, ";")
			check := false
			if strings.Contains(rawTags, "fieldName") && strings.Contains(rawTags, "fieldType") {
				check = true
			}
			field := ""
			primaryCheck := false
			for _, tag := range tags {
				tempTag := strings.Split(tag, ":")
				if len(tempTag) != 2 {
					return fmt.Errorf("format tags salah pada field : %v", reflectType.Field(i).Name)
				}
				if tempTag[len(tempTag)-1] == "" {
					return fmt.Errorf("value tag kososng pada field : %v", reflectType.Field(i).Name)
				}
				switch tempTag[0] {
				case "validate":
					validate, err := strconv.ParseBool(tempTag[len(tempTag)-1])
					if err != nil {
						return err
					}
					if validateCheck && validate && reflect.ValueOf(reflectValue.Field(i).Interface()).IsZero() {
						return fmt.Errorf("value field %s tidak boleh kosong", reflectType.Field(i).Name)
					}
				case "fieldName":
					field = tempTag[1]
				case "fieldLength":
				case "fieldType":
					if strings.ToLower(tempTag[1]) == "serial" {
						t.AutoIncrement = true
						t.ReturningID = true
					}
				case "primaryKey":
					checked, err := strconv.ParseBool(tempTag[len(tempTag)-1])
					if err != nil {
						return err
					}
					if checked {
						primaryCheck = true
					}
				default:
					return fmt.Errorf("perintah tidak dikenal : %s pada atribut : %s", tempTag[0], reflectType.Field(i).Name)
				}
			}
			if check {
				fields = append(fields, field)
				if primaryCheck {
					t.PrimaryKey = field
				}
			}
		}
		re, err := regexp.Compile(`[^\w]`)
		if err != nil {
			return err
		}
		temp := reflectType.Field(i).Name
		field := re.ReplaceAllString(temp, "")
		if strings.ToLower(field) == t.PrimaryKey {
			t.DstPrimaryIndex = i
			t.DstPrimary = reflectValue.Field(i).Addr().Interface()
			if primaryNotInclude {
				continue
			}
		}
		result = append(result, reflectValue.Field(i).Addr().Interface())
	}
	if primaryNotInclude {
		result = append(result, reflectValue.Field(t.DstPrimaryIndex).Addr().Interface())
	}
	t.DstFields = result
	t.Fields = fields
	return nil
}

//getQueryCreate : fungsi ini digunakan untuk menyiapkan data untuk create table
func (t *Table) getQueryCreate(item interface{}) ([]string, error) {
	reflectValue := reflect.ValueOf(item)
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}

	var reflectType = reflectValue.Type()
	type FieldConstructor struct {
		Name       string
		Length     string
		Type       string
		PrimaryKey bool
	}

	var fields []*FieldConstructor
	for i := 0; i < reflectValue.NumField(); i++ {
		rawTags := reflectType.Field(i).Tag.Get("fdb")
		if rawTags == "-" {
			continue
		} else if rawTags != "" {
			tags := strings.Split(rawTags, ";")
			field := &FieldConstructor{}
			check := false
			if strings.Contains(rawTags, "fieldName") && strings.Contains(rawTags, "fieldType") {
				check = true
			}
			for _, tag := range tags {
				tempTag := strings.Split(tag, ":")
				if len(tempTag) != 2 {
					return nil, fmt.Errorf("format tags salah pada field : %v", reflectType.Field(i).Name)
				}
				if tempTag[len(tempTag)-1] == "" {
					return nil, fmt.Errorf("value tag kososng pada field : %v", reflectType.Field(i).Name)
				}
				switch tempTag[0] {
				case "validate":
				case "fieldName":
					field.Name = tempTag[1]
				case "fieldType":
					field.Type = tempTag[1]
				case "fieldLength":
					field.Length = tempTag[1]
				case "primaryKey":
					checked, err := strconv.ParseBool(tempTag[len(tempTag)-1])
					if err != nil {
						return nil, err
					}
					if checked {
						field.PrimaryKey = true
					}
				default:
					return nil, fmt.Errorf("perintah tidak dikenal : %s pada atribut : %s", tempTag[0], reflectType.Field(i).Name)
				}
			}
			if check {
				fields = append(fields, field)
			}
		}
	}
	var queryFields []string
	for _, field := range fields {
		queryField := fmt.Sprintf("%s %s", field.Name, field.Type)
		if field.Length != "" {
			queryField = fmt.Sprintf("%s %s(%s)", field.Name, field.Type, field.Length)
		}
		if field.PrimaryKey {
			queryField += fmt.Sprintf(" PRIMARY KEY")
		}
		queryFields = append(queryFields, queryField)
	}
	return queryFields, nil
}

// getFieldWithoutPrimary : fungsi ini digunakan untuk mendapatkan susunan field tanpa primary key
func (t *Table) getFieldWithoutPrimary() []string {
	var result []string
	for _, value := range t.Fields {
		if value != t.PrimaryKey {
			result = append(result, value)
		}
	}
	return result
}
func (t *Table) tableName() string {
	result := t.Name
	if t.Schema != "" {
		result = fmt.Sprintf("%s.%s", t.Schema, t.Name)
	}
	return result
}

//FieldsToVariabel : digunakan untuk mendapatkan seberapa banyak $
func FieldsToVariables(fields []string, autoNumber bool) string {
	var params []string
	for i := 0; i < len(fields); i++ {
		if i+1 == len(fields) && autoNumber {
			continue
		}
		params = append(params, fmt.Sprintf("$%d", i+1))
	}
	return fmt.Sprintf("(%s)", strings.Join(params, ","))
}
