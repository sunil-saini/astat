package aws

import (
	"context"
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchLambdaFunctions(ctx context.Context, cfg sdkaws.Config) ([]model.LambdaFunction, error) {
	client := lambda.NewFromConfig(cfg)

	var funcs []model.LambdaFunction
	var marker *string

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		out, err := client.ListFunctions(ctx, &lambda.ListFunctionsInput{
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}

		for _, f := range out.Functions {
			funcs = append(funcs, model.LambdaFunction{
				Name:         sdkaws.ToString(f.FunctionName),
				Runtime:      string(f.Runtime),
				LastModified: sdkaws.ToString(f.LastModified),
				Memory:       fmt.Sprintf("%d", sdkaws.ToInt32(f.MemorySize)),
				Timeout:      fmt.Sprintf("%d", sdkaws.ToInt32(f.Timeout)),
			})
		}

		if out.NextMarker == nil {
			break
		}
		marker = out.NextMarker
	}

	return funcs, nil
}
