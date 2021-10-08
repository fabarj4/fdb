package fdb

import (
	"fmt"
	"testing"
)

func TestCursor(t *testing.T) {
	cursor := &Cursor{
		Filters: "filters=coba,=,1",
		Sort:    "sort=id ASC",
		Limit:   "limit=5",
	}
	cursorString := "ZmlsdGVycz1jb2JhLD0sMSZzb3J0PWlkIEFTQyZsaW1pdD01"
	t.Run("Test set Cursor", func(t *testing.T) {
		result := cursor.SetCursor()
		fmt.Println(result)
		if result != cursorString {
			t.Fatalf("hasil output tidak sesuai")
		}
	})
	t.Run("Test Get Cursor", func(t *testing.T) {
		temp := &Cursor{}
		if err := temp.GetCursor(cursorString); err != nil {
			t.Fatalf("hasil output tidak sesuai")
		}
		if temp.Filters != cursor.Filters {
			t.Fatalf("hasil output tidak sesuai got : %s want :%s", temp.Filters, cursor.Filters)
		}
		if temp.Limit != cursor.Limit {
			t.Fatalf("hasil output tidak sesuai got : %s want :%s", temp.Limit, cursor.Limit)
		}
		if temp.Sort != cursor.Sort {
			t.Fatalf("hasil output tidak sesuai got : %s want :%s", temp.Sort, cursor.Sort)
		}
	})
}
