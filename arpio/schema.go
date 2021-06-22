package arpio

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func GetFirstElementAsMap(s interface{}) (m map[string]interface{}, ok bool) {
	var slice []interface{}

	switch s.(type) {
	case *schema.Set:
		slice = s.(*schema.Set).List()
	case *[]interface{}:
		slice = *(s.(*[]interface{}))
	case []interface{}:
		slice = s.([]interface{})
	default:
		panic("not a *schema.Set or *[]interface{}")
	}

	if len(slice) == 0 {
		return nil, false
	}
	if slice[0] == nil {
		return nil, false
	}
	return slice[0].(map[string]interface{}), true
}

func FromSetToStringList(v interface{}) []string {
	if v != nil {
		return TypifyStringList(v.(*schema.Set).List())
	}
	return []string{}
}

func FromSetToStringMap(v interface{}) map[string]string {
	if v != nil {
		return TypifyStringMap(v.(map[string]interface{}))
	}
	return map[string]string{}
}

func TypifyStringList(values []interface{}) []string {
	list := make([]string, 0, len(values))
	for _, val := range values {
		list = append(list, val.(string))
	}
	return list
}

func UntypifyStringList(values []string) []interface{} {
	list := make([]interface{}, 0, len(values))
	for _, val := range values {
		list = append(list, val)
	}
	return list
}

func TypifyStringMap(values map[string]interface{}) map[string]string {
	m := map[string]string{}
	for k, v := range values {
		m[k] = v.(string)
	}
	return m
}

func UntypifyStringMap(values map[string]string) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range values {
		m[k] = v
	}
	return m
}

func ParseRFC3339Timestamp(ts string) (*time.Time, error) {
	if ts == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
