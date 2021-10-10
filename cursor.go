package fdb

import (
	"encoding/base64"
	"strings"
)

type Cursor struct {
	Filters string
	Sort    string
	Limit   string
	Offset  string
}

func MapToCursor(item map[string]interface{}) *Cursor {
	result := &Cursor{}
	if item, ok := item["filters"]; ok {
		result.Filters = item.(string)
	}
	if item, ok := item["sort"]; ok {
		result.Sort = item.(string)
	}
	if item, ok := item["limit"]; ok {
		result.Limit = item.(string)
	}
	if item, ok := item["offset"]; ok {
		result.Offset = item.(string)
	}
	return result
}

func (c *Cursor) SetCursor() string {
	var result string
	temp := []string{}
	if c.Filters != "" {
		temp = append(temp, c.Filters)
	} else {
		temp = append(temp, "null")
	}
	if c.Sort != "" {
		temp = append(temp, c.Sort)
	} else {
		temp = append(temp, "null")
	}
	if c.Limit != "" {
		temp = append(temp, c.Limit)
	} else {
		temp = append(temp, "null")
	}
	if c.Offset != "" {
		temp = append(temp, c.Offset)
	} else {
		temp = append(temp, "null")
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
	if items[0] != "null" {
		c.Filters = items[0]
	}
	if items[1] != "null" {
		c.Sort = items[1]
	}
	if items[2] != "null" {
		c.Limit = items[2]
	}
	if items[3] != "null" {
		c.Offset = items[3]
	}
	return nil
}
