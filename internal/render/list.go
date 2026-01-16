package render

import (
	"context"
	"fmt"
	"reflect"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/cache"
	"github.com/sunil-saini/astat/internal/logger"
	"github.com/sunil-saini/astat/internal/refresh"
)

// List handles the common logic for listing resources
// 1. loading from cache
// 2. refreshing if needed
// 3. automatically extracting headers and rows from the model
func List[T any](
	cmd *cobra.Command,
	cacheKey string,
	fetch func(context.Context, sdkaws.Config) ([]T, error),
) error {
	forceRefresh := viper.GetBool("refresh")
	cacheFile := cache.Path(cache.Dir(), cacheKey)

	var data []T
	hit, err := cache.Load(cacheFile, &data)
	if err != nil {
		logger.Error("cache read failed: %v", err)
		return err
	}

	ctx := cmd.Context()

	if !hit || forceRefresh {
		refresh.RefreshSync(ctx, cacheKey, fetch)
		_, err = cache.Load(cacheFile, &data)
		if err != nil {
			return err
		}
	}

	var headers []string
	var fields []int

	// Extract headers from struct tags
	var t T
	typ := reflect.TypeOf(t)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if header := field.Tag.Get("header"); header != "" {
			headers = append(headers, header)
			fields = append(fields, i)
		}
	}

	rows := make([][]string, 0, len(data))
	for _, item := range data {
		row := make([]string, 0, len(fields))
		val := reflect.ValueOf(item)
		for _, idx := range fields {
			fieldVal := val.Field(idx)
			row = append(row, fmt.Sprintf("%v", fieldVal.Interface()))
		}
		rows = append(rows, row)
	}

	return Print(TableData{
		Headers: headers,
		Rows:    rows,
		JSON:    data,
	})
}
