package fdb

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type Cursor struct {
	Filters string
	Sort    string
	Limit   string
	Offset  string
}

func (c *Cursor) SetCursor() string {
	var result string
	temp := []string{}
	if c.Filters != "" {
		temp = append(temp, c.Filters)
	}
	if c.Sort != "" {
		temp = append(temp, c.Sort)
	}
	if c.Limit != "" {
		temp = append(temp, c.Limit)
	}
	if c.Offset != "" {
		temp = append(temp, c.Offset)
	}
	result = base64.StdEncoding.EncodeToString([]byte(strings.Join(temp, "&")))
	return result
}

func (c *Cursor) GetCursor(data string) error {
	temp, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	items := strings.Split(string(temp), "&")
	fmt.Println(items)
	for _, item := range items {
		if strings.Contains(item, "filters") {
			c.Filters = item
		}
		if strings.Contains(item, "sort") {
			c.Sort = item
		}
		if strings.Contains(item, "limit") {
			c.Limit = item
		}
		if strings.Contains(item, "offset") {
			c.Offset = item
		}
	}
	return nil
}
