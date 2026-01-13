package aws

import (
	"context"

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
				Name:    *f.FunctionName,
				Runtime: string(f.Runtime),
				Handler: *f.Handler,
				Region:  cfg.Region,
			})
		}

		if out.NextMarker == nil {
			break
		}
		marker = out.NextMarker
	}

	return funcs, nil
}
