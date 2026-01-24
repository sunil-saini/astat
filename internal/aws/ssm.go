package aws

import (
	"context"

	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchSSMParameters(ctx context.Context, cfg sdkaws.Config) ([]model.SSMParameter, error) {
	client := ssm.NewFromConfig(cfg)

	var params []model.SSMParameter
	var nextToken *string

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		out, err := client.DescribeParameters(ctx, &ssm.DescribeParametersInput{
			NextToken:  nextToken,
			MaxResults: sdkaws.Int32(50),
		})
		if err != nil {
			return nil, err
		}

		for _, p := range out.Parameters {
			params = append(params, model.SSMParameter{
				Name:         *p.Name,
				Type:         string(p.Type),
				LastModified: p.LastModifiedDate.Format("2006-01-02 15:04:05"),
				ModifiedBy:   shortenARN(*p.LastModifiedUser),
			})
		}

		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	return params, nil
}

func GetSSMParameter(ctx context.Context, cfg sdkaws.Config, name string) (string, error) {
	client := ssm.NewFromConfig(cfg)

	out, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: sdkaws.Bool(true),
	})
	if err != nil {
		return "", err
	}

	return *out.Parameter.Value, nil
}

func shortenARN(arn string) string {
	if !strings.HasPrefix(arn, "arn:aws:") {
		return arn
	}

	// Extract the resource part after the last colon
	lastColon := strings.LastIndex(arn, ":")
	if lastColon == -1 {
		return arn
	}
	resource := arn[lastColon+1:]

	// For assumed roles: assumed-role/RoleName/SessionName -> assumed-role/RoleName
	if strings.HasPrefix(resource, "assumed-role/") {
		parts := strings.Split(resource, "/")
		if len(parts) >= 2 {
			return "assumed-role/" + parts[1]
		}
	}

	// For other resources (user/Name, role/Name), take the last segment
	parts := strings.Split(resource, "/")
	return parts[len(parts)-1]
}
