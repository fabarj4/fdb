package fdb

import (
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

type Dummy struct {
	ID       int
	Username string
	Password string
}

func TestDB(t *testing.T) {
	database := &Database{
		Dbname:    "fdb",
		Dbtest:    "test_fdb",
		Dropfirst: true,
	}

	t.Run("test create database", func(t *testing.T) {
		db, err := Connect("postgres", "postgres", "postgres")
		if err != nil {
			t.Fatalf("error : %s", err.Error())
		}
		if err := database.Init(db); err != nil {
			t.Fatal(err)
		}
		db.Close()
	})

	db, err := Connect("postgres", "postgres", "test_fdb")
	if err != nil {
		t.Fatalf("error : %s", err.Error())
	}
	defer db.Close()

	temp := []*Dummy{
		{Username: "fabar", Password: "123456"},
		{Username: "falbar", Password: "123456"},
		{Username: "farbar", Password: "123456"},
		{Username: "faibar", Password: "123456"},
		{Username: "fasbar", Password: "123456"},
		{Username: "fabbar", Password: "123456"},
	}

	tbl := &Table{
		Name:          "users",
		PrimaryKey:    "id",
		Fields:        []string{"id", "username", "password"},
		ReturningID:   true,
		AutoIncrement: true,
	}

	t.Run("test create table", func(t *testing.T) {
		query := `CREATE TABLE users(
			id serial primary key,
			username varchar(60),
			password text
		);`
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test insert table", func(t *testing.T) {
		for _, item := range temp {
			if err = tbl.Insert(db, item); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("test update table", func(t *testing.T) {
		du := map[string]interface{}{
			"password": "asd",
			"username": "asd2",
		}
		if err := tbl.Update(db, temp[0], du); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test get table", func(t *testing.T) {
		user := &Dummy{ID: 2}
		if err := tbl.Get(db, user); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test gets table", func(t *testing.T) {
		temp := &Dummy{}
		data, _, err := tbl.Gets(db, temp, nil)
		if err != nil {
			t.Fatal(err)
		}
		result := make([]*Dummy, len(data))
		for index, item := range data {
			result[index] = item.(*Dummy)
		}
		for _, item := range result {
			fmt.Println(item)
		}
	})

	t.Run("test gets table with params sort", func(t *testing.T) {
		cursor := Cursor{
			Sort: "id DESC",
		}
		temp := &Dummy{}
		data, _, err := tbl.Gets(db, temp, &cursor)
		if err != nil {
			t.Fatal(err)
		}
		result := make([]*Dummy, len(data))
		for index, item := range data {
			result[index] = item.(*Dummy)
		}
		for _, item := range result {
			fmt.Println(item)
		}
	})

	t.Run("test delete table", func(t *testing.T) {
		for _, item := range temp {
			if err = tbl.Delete(db, item); err != nil {
				t.Fatal(err)
			}
		}
	})
}
