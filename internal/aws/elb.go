package aws

import (
	"context"
	"sync"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/sunil-saini/astat/internal/model"
)

func FetchLoadBalancers(ctx context.Context, cfg sdkaws.Config) ([]model.LoadBalancer, error) {
	var lbs []model.LoadBalancer
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Fetch v1 Load Balancers (Classic)
	wg.Go(func() {
		clientV1 := elb.NewFromConfig(cfg)
		if out, err := clientV1.DescribeLoadBalancers(ctx, &elb.DescribeLoadBalancersInput{}); err == nil {
			mu.Lock()
			for _, lb := range out.LoadBalancerDescriptions {
				lbs = append(lbs, model.LoadBalancer{
					Type:      "classic",
					Name:      *lb.LoadBalancerName,
					Scheme:    *lb.Scheme,
					CreatedAt: lb.CreatedTime.Format(time.RFC3339),
					DNSName:   *lb.DNSName,
				})
			}
			mu.Unlock()
		}
	})

	// Fetch v2 Load Balancers (ALB/NLB)
	wg.Go(func() {
		clientV2 := elbv2.NewFromConfig(cfg)
		if out, err := clientV2.DescribeLoadBalancers(ctx, &elbv2.DescribeLoadBalancersInput{}); err == nil {
			mu.Lock()
			for _, lb := range out.LoadBalancers {
				lbs = append(lbs, model.LoadBalancer{
					Type:      string(lb.Type),
					Name:      *lb.LoadBalancerName,
					DNSName:   *lb.DNSName,
					Scheme:    string(lb.Scheme),
					CreatedAt: lb.CreatedTime.Format(time.RFC3339),
				})
			}
			mu.Unlock()
		}
	})

	wg.Wait()
	return lbs, nil
}
