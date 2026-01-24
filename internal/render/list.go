package render

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/cache"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/model"
	"github.com/sunil-saini/astat/internal/refresh"
	"github.com/sunil-saini/astat/internal/registry"
)

// List handles the common logic for listing resources
// 1. loading from cache
// 2. refreshing if needed
// 3. automatically extracting headers and rows from the model
// 4. filtering by search term if provided in args
func List(
	cmd *cobra.Command,
	args []string,
	serviceName string,
) error {
	service, err := getService(serviceName)
	if err != nil {
		return err
	}

	dataPtr, hit, err := loadCache(cmd.Context(), service)
	if err != nil {
		return err
	}

	if viper.GetBool("refresh") || !hit {
		refresh.RefreshSync(cmd.Context(), serviceName, service.Fetch)
		dataPtr, _, err = loadCache(cmd.Context(), service)
		if err != nil {
			return err
		}
	}

	headers, fields := extractHeaders(service.Model)
	items := model.ToAnySlice(dataPtr.Elem().Interface())

	searchTerm := ""
	if len(args) > 0 {
		searchTerm = strings.ToLower(args[0])
	}

	rows, filteredData := filterRows(items, fields, searchTerm)

	return Print(TableData{
		Headers: headers,
		Rows:    rows,
		JSON:    filteredData,
	})
}

func getService(name string) (*registry.Service, error) {
	for _, s := range registry.Registry {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("unknown service: %s", name)
}

func loadCache(ctx context.Context, service *registry.Service) (reflect.Value, bool, error) {
	if err := ctx.Err(); err != nil {
		return reflect.Value{}, false, err
	}

	cacheFile := cache.Path(cache.Dir(), service.Name)
	sliceType := reflect.SliceOf(reflect.TypeOf(service.Model))
	dataPtr := reflect.New(sliceType)

	hit, err := cache.Load(cacheFile, dataPtr.Interface())
	if err != nil {
		logger.Error("cache read failed: %v", err)
		return reflect.Value{}, false, err
	}
	return dataPtr, hit, nil
}

func extractHeaders(m any) ([]string, []int) {
	var headers []string
	var fields []int

	typ := reflect.TypeOf(m)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if header := field.Tag.Get("header"); header != "" {
			headers = append(headers, header)
			fields = append(fields, i)
		}
	}
	return headers, fields
}

func filterRows(items []any, fields []int, searchTerm string) ([][]string, []any) {
	var filteredData []any
	rows := make([][]string, 0)

	for _, item := range items {
		row := make([]string, 0, len(fields))
		val := reflect.ValueOf(item)
		match := searchTerm == ""

		for _, idx := range fields {
			fieldVal := val.Field(idx)
			cell := fmt.Sprintf("%v", fieldVal.Interface())
			row = append(row, cell)

			if !match && strings.Contains(strings.ToLower(cell), searchTerm) {
				match = true
			}
		}

		if match {
			rows = append(rows, row)
			filteredData = append(filteredData, item)
		}
	}
	return rows, filteredData
}
