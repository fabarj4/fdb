package fdb

import (
	"fmt"
	"testing"
)

func TestCursor(t *testing.T) {
	cursor := &Cursor{
		Filters: "coba,=,1",
		Sort:    "id ASC",
		Limit:   "5",
	}
	cursorString := "Y29iYSw9LDEmaWQgQVNDJjUmbnVsbA=="
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
		fmt.Println(temp)
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
