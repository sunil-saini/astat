package aws

import (
	"context"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchSQSQueues(ctx context.Context, cfg sdkaws.Config) ([]model.SQSQueue, error) {
	client := sqs.NewFromConfig(cfg)

	paginator := sqs.NewListQueuesPaginator(client, &sqs.ListQueuesInput{})

	var queues []model.SQSQueue
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, url := range page.QueueUrls {
			name := url[strings.LastIndex(url, "/")+1:]
			qType := "Standard"
			if strings.HasSuffix(name, ".fifo") {
				qType = "FIFO"
			}

			queues = append(queues, model.SQSQueue{
				Name: name,
				Type: qType,
			})
		}
	}

	return queues, nil
}
