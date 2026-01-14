package aws

import (
	"context"
	"fmt"
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
	wg.Add(1)
	go func() {
		defer wg.Done()
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
	}()

	// Fetch v2 Load Balancers (ALB/NLB)
	wg.Add(1)
	go func() {
		defer wg.Done()
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
					ARN:       *lb.LoadBalancerArn,
				})
			}
			mu.Unlock()
		}
	}()

	wg.Wait()
	return lbs, nil
}

func FetchTargetGroups(ctx context.Context, cfg sdkaws.Config, lbARN *string) ([]model.TargetGroup, error) {
	client := elbv2.NewFromConfig(cfg)
	out, err := client.DescribeTargetGroups(ctx, &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: lbARN,
	})
	if err != nil {
		return nil, err
	}

	var tgs []model.TargetGroup
	for _, tg := range out.TargetGroups {
		tgs = append(tgs, model.TargetGroup{
			Name:       *tg.TargetGroupName,
			Protocol:   string(tg.Protocol),
			Port:       *tg.Port,
			TargetType: string(tg.TargetType),
			ARN:        *tg.TargetGroupArn,
		})
	}
	return tgs, nil
}

func FetchListeners(ctx context.Context, cfg sdkaws.Config, lbARN string) ([]model.Listener, error) {
	client := elbv2.NewFromConfig(cfg)
	out, err := client.DescribeListeners(ctx, &elbv2.DescribeListenersInput{
		LoadBalancerArn: &lbARN,
	})
	if err != nil {
		return nil, err
	}

	var listeners []model.Listener
	for _, l := range out.Listeners {
		var actions []model.Action
		for _, a := range l.DefaultActions {
			tgArn := ""
			if a.TargetGroupArn != nil {
				tgArn = *a.TargetGroupArn
			}
			actions = append(actions, model.Action{
				Type:           string(a.Type),
				TargetGroupARN: tgArn,
			})
		}
		listeners = append(listeners, model.Listener{
			ARN:            *l.ListenerArn,
			Protocol:       string(l.Protocol),
			Port:           *l.Port,
			DefaultActions: actions,
		})
	}
	return listeners, nil
}

func FetchRules(ctx context.Context, cfg sdkaws.Config, listenerARN string) ([]model.Rule, error) {
	client := elbv2.NewFromConfig(cfg)
	out, err := client.DescribeRules(ctx, &elbv2.DescribeRulesInput{
		ListenerArn: &listenerARN,
	})
	if err != nil {
		return nil, err
	}

	var rules []model.Rule
	for _, r := range out.Rules {
		var conditions []model.Condition
		for _, c := range r.Conditions {
			var values []string
			if c.HostHeaderConfig != nil {
				values = c.HostHeaderConfig.Values
			}
			if c.PathPatternConfig != nil {
				values = c.PathPatternConfig.Values
			}
			field := ""
			if c.Field != nil {
				field = *c.Field
			}
			conditions = append(conditions, model.Condition{
				Field:  field,
				Values: values,
			})
		}

		var actions []model.Action
		for _, a := range r.Actions {
			tgArn := ""
			if a.TargetGroupArn != nil {
				tgArn = *a.TargetGroupArn
			}
			actions = append(actions, model.Action{
				Type:           string(a.Type),
				TargetGroupARN: tgArn,
			})
		}

		rules = append(rules, model.Rule{
			ARN:        *r.RuleArn,
			Priority:   *r.Priority,
			IsDefault:  *r.IsDefault,
			Conditions: conditions,
			Actions:    actions,
		})
	}
	return rules, nil
}

func FetchTargetHealth(ctx context.Context, cfg sdkaws.Config, tgARN string) ([]model.InstanceHealth, error) {
	client := elbv2.NewFromConfig(cfg)
	out, err := client.DescribeTargetHealth(ctx, &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: &tgARN,
	})
	if err != nil {
		return nil, err
	}

	var health []model.InstanceHealth
	for _, th := range out.TargetHealthDescriptions {
		reason := ""
		if th.TargetHealth.Reason != "" {
			reason = string(th.TargetHealth.Reason)
		}
		health = append(health, model.InstanceHealth{
			InstanceID: *th.Target.Id,
			State:      string(th.TargetHealth.State),
			Reason:     reason,
		})
	}
	return health, nil
}

func GetClassicLBDetails(ctx context.Context, cfg sdkaws.Config, lbName string) ([]model.Listener, []model.InstanceHealth, error) {
	client := elb.NewFromConfig(cfg)
	out, err := client.DescribeLoadBalancers(ctx, &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{lbName},
	})
	if err != nil {
		return nil, nil, err
	}

	if len(out.LoadBalancerDescriptions) == 0 {
		return nil, nil, fmt.Errorf("LB not found")
	}

	lb := out.LoadBalancerDescriptions[0]
	var listeners []model.Listener
	for _, l := range lb.ListenerDescriptions {
		protocol := ""
		if l.Listener.Protocol != nil {
			protocol = *l.Listener.Protocol
		}
		listeners = append(listeners, model.Listener{
			Protocol: protocol,
			Port:     int32(l.Listener.LoadBalancerPort),
		})
	}

	var healths []model.InstanceHealth
	hOut, err := client.DescribeInstanceHealth(ctx, &elb.DescribeInstanceHealthInput{
		LoadBalancerName: &lbName,
	})
	if err == nil {
		for _, s := range hOut.InstanceStates {
			state := ""
			if s.State != nil {
				state = *s.State
			}
			reason := ""
			if s.Description != nil {
				reason = *s.Description
			}
			healths = append(healths, model.InstanceHealth{
				InstanceID: *s.InstanceId,
				State:      state,
				Reason:     reason,
			})
		}

		return listeners, healths, nil
	}

	// Fallback to just IDs if health check fails
	for _, inst := range lb.Instances {
		healths = append(healths, model.InstanceHealth{
			InstanceID: *inst.InstanceId,
		})
	}

	return listeners, healths, nil
}
