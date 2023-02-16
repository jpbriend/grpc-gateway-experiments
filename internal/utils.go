package internal

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type Sorter[T any] struct {
	field string
}

func NewSorter[T any](field string) *Sorter[T] {
	return &Sorter[T]{field: field}
}

// Sort sorts the given slice of objects by the given field, ignoring case of the field name
// sorts using the string value of the field
func (s *Sorter[T]) Sort(o []*T) ([]*T, error) {
	f, found := reflect.TypeOf(*o[0]).FieldByNameFunc(func(str string) bool {
		return strings.EqualFold(strings.ToLower(s.field), strings.ToLower(str))
	})

	if !found {
		return nil, fmt.Errorf("field %s not found", s.field)
	} else {
		res := o
		sort.Slice(res, func(i, j int) bool {
			valI := reflect.ValueOf(*res[i]).FieldByIndex(f.Index).Interface()
			valJ := reflect.ValueOf(*res[j]).FieldByIndex(f.Index).Interface()

			return valI.(string) < valJ.(string)
		})
		return res, nil
	}
}
