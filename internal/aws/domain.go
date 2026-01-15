package aws

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/sunil-saini/astat/internal/cache"
	"github.com/sunil-saini/astat/internal/model"
)

func TraceDomain(ctx context.Context, cfg sdkaws.Config, domain string) (*model.TraceResult, error) {
	result := &model.TraceResult{Domain: domain}

	// Normalize input
	input := strings.TrimSuffix(domain, ".")
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")

	parts := strings.SplitN(input, "/", 2)
	host := parts[0]
	path := "/"
	if len(parts) > 1 {
		path = "/" + parts[1]
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 1. Resolve Route53 Record
	foundRecord := resolveRoute53Record(ctx, cfg, host)
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if foundRecord == nil {
		// Not in Route53
		hops, err := traceExternalDNS(ctx, host)
		if err != nil {
			return nil, err
		}
		result.Hops = append(result.Hops, model.TraceNode{
			Type:  model.NodeDNS,
			Name:  "External DNS",
			Value: fmt.Sprintf("Not in Route53. Dig results: %s", hops),
		})
		return result, nil
	}

	// In Route53
	r53Node := model.TraceNode{
		Type:  model.NodeRoute53,
		Name:  foundRecord.Name,
		Value: fmt.Sprintf("%s (%s)", foundRecord.Value, foundRecord.Type),
	}

	target := strings.TrimSuffix(foundRecord.Value, ".")

	// 2. Trace CloudFront
	if node, matched := traceCloudFront(ctx, cfg, target, domain, path); matched {
		r53Node.Children = append(r53Node.Children, *node)
		result.Hops = append(result.Hops, r53Node)
		return result, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 3. Trace Load Balancers
	lbs, _ := FetchLoadBalancers(ctx, cfg)
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	normalizedTarget := strings.TrimPrefix(target, "dualstack.")
	for _, lb := range lbs {
		if strings.TrimSuffix(lb.DNSName, ".") == normalizedTarget {
			lbNode := traceLoadBalancer(ctx, cfg, lb, host, path)
			r53Node.Children = append(r53Node.Children, lbNode)
			result.Hops = append(result.Hops, r53Node)
			return result, nil
		}
	}

	result.Hops = append(result.Hops, r53Node)
	return result, nil
}

func resolveRoute53Record(ctx context.Context, cfg sdkaws.Config, host string) *model.Route53Record {
	// 1. Check Route53 Records Cache
	var records []model.Route53Record
	if ok, _ := cache.Load(cache.Path(cache.Dir(), "route53-records"), &records); ok {
		for _, r := range records {
			if strings.TrimSuffix(r.Name, ".") == host {
				return &r
			}
		}
	}

	// 2. If not found, check Zones Cache and fetch only for that zone
	var zones []model.Route53HostedZone
	if ok, _ := cache.Load(cache.Path(cache.Dir(), "route53-zones"), &zones); ok {
		var matchedZone *model.Route53HostedZone
		for _, z := range zones {
			zoneName := strings.TrimSuffix(z.Name, ".")
			if strings.HasSuffix(host, zoneName) {
				if matchedZone == nil || len(z.Name) > len(matchedZone.Name) {
					matchedZone = &z
				}
			}
		}

		if matchedZone != nil {
			zoneID := strings.TrimPrefix(matchedZone.ID, "/hostedzone/")
			client := route53.NewFromConfig(cfg)

			var startName *string
			var startType types.RRType

			for {
				out, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
					HostedZoneId:    &zoneID,
					StartRecordName: startName,
					StartRecordType: startType,
				})
				if err != nil {
					break
				}

				for _, r := range out.ResourceRecordSets {
					if strings.TrimSuffix(*r.Name, ".") == host {
						return mapR53Record(matchedZone.Name, r)
					}
				}

				if !out.IsTruncated {
					break
				}
				startName = out.NextRecordName
				startType = out.NextRecordType
			}
		}
	}
	return nil
}

func traceCloudFront(ctx context.Context, cfg sdkaws.Config, target, host, path string) (*model.TraceNode, bool) {
	cfDists, _ := FetchCloudFront(ctx, cfg)
	for _, d := range cfDists {
		distDomain := strings.TrimSuffix(d.Domain, ".")
		distAliases := strings.Split(d.Aliases, ",")

		isMatch := (strings.EqualFold(distDomain, target))
		if !isMatch {
			for _, a := range distAliases {
				if matchHost(target, a) {
					isMatch = true
					break
				}
			}
		}

		if isMatch {
			cfNode := model.TraceNode{
				Type: model.NodeCloudFront,
				Name: fmt.Sprintf("Distribution (%s)", d.ID),
			}

			originDomain := d.DefaultOrigin
			matchedPattern := "Default (*)"

			for _, b := range d.Behaviors {
				if matchPath(path, b.PathPattern) {
					originDomain = d.Origins[b.TargetOriginID]
					matchedPattern = b.PathPattern
					break
				}
			}

			cfNode.Children = append(cfNode.Children, model.TraceNode{
				Type:  model.NodeOrigin,
				Name:  fmt.Sprintf("Origin (via %s)", matchedPattern),
				Value: originDomain,
			})
			return &cfNode, true
		}
	}
	return nil, false
}

func traceExternalDNS(ctx context.Context, domain string) (string, error) {
	resolver := &net.Resolver{}
	ips, err := resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return "", fmt.Errorf("domain not found: %w", err)
	}

	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}
	result := strings.Join(ipStrings, ", ")

	cname, err := resolver.LookupCNAME(ctx, domain)
	if err == nil && strings.TrimSuffix(cname, ".") != domain {
		result = fmt.Sprintf("%s -> %s", cname, result)
	}

	return result, nil
}

func mapR53Record(zoneName string, r types.ResourceRecordSet) *model.Route53Record {
	ttl := ""
	recordType := string(r.Type)
	if r.AliasTarget != nil {
		ttl = "ALIAS"
		recordType = "Alias+" + recordType
	} else if r.TTL != nil {
		ttl = fmt.Sprintf("%d", *r.TTL)
	}

	value := ""
	if r.AliasTarget != nil {
		value = *r.AliasTarget.DNSName
	} else if len(r.ResourceRecords) > 0 {
		value = *r.ResourceRecords[0].Value
	}

	return &model.Route53Record{
		ZoneName: zoneName,
		Name:     *r.Name,
		Type:     recordType,
		TTL:      ttl,
		Value:    value,
	}
}

func matchPath(path, pattern string) bool {
	return matchPattern(path, pattern)
}

func traceLoadBalancer(ctx context.Context, cfg sdkaws.Config, lb model.LoadBalancer, host, path string) model.TraceNode {
	nodeType := model.NodeALB
	switch lb.Type {
	case "classic":
		nodeType = model.NodeCLB
	case "network":
		nodeType = model.NodeNLB
	}

	lbNode := model.TraceNode{
		Type:  nodeType,
		Name:  fmt.Sprintf("%s (%s)", lb.Name, lb.Scheme),
		ID:    lb.ARN,
		Value: lb.DNSName,
	}

	ec2Names := getEC2Names()

	switch lb.Type {
	case "classic":
		return traceClassicLB(ctx, cfg, lbNode, lb, ec2Names)
	case "network":
		return traceNetworkLB(ctx, cfg, lbNode, lb, ec2Names)
	default:
		return traceApplicationLB(ctx, cfg, lbNode, lb, host, path, ec2Names)
	}
}

func getEC2Names() map[string]string {
	ec2Names := make(map[string]string)
	var instances []model.EC2Instance
	if ok, _ := cache.Load(cache.Path(cache.Dir(), "ec2"), &instances); ok {
		for _, inst := range instances {
			name := inst.Name
			if name == "" {
				name = inst.InstanceID
			}
			ec2Names[inst.InstanceID] = name
		}
	}
	return ec2Names
}

func traceClassicLB(ctx context.Context, cfg sdkaws.Config, lbNode model.TraceNode, lb model.LoadBalancer, ec2Names map[string]string) model.TraceNode {
	listeners, healths, _ := GetClassicLBDetails(ctx, cfg, lb.Name)
	for _, l := range listeners {
		lNode := model.TraceNode{
			Type: "Listener",
			Name: fmt.Sprintf("%s:%d", l.Protocol, l.Port),
		}
		for _, h := range healths {
			name := ec2Names[h.InstanceID]
			if name == "" {
				name = h.InstanceID
			}
			val := h.State
			if h.Reason != "" && h.Reason != "N/A" {
				val += " (" + h.Reason + ")"
			}
			status := "unhealthy"
			if h.State == "InService" {
				status = "healthy"
			}
			lNode.Children = append(lNode.Children, model.TraceNode{
				Type:   "Instance",
				Name:   name,
				Value:  val,
				Status: status,
			})
		}
		lbNode.Children = append(lbNode.Children, lNode)
	}
	return lbNode
}

func traceNetworkLB(ctx context.Context, cfg sdkaws.Config, lbNode model.TraceNode, lb model.LoadBalancer, ec2Names map[string]string) model.TraceNode {
	listeners, _ := FetchListeners(ctx, cfg, lb.ARN)
	for _, l := range listeners {
		listenerNode := model.TraceNode{
			Type: "Listener",
			Name: fmt.Sprintf("%s:%d", l.Protocol, l.Port),
		}
		for _, action := range l.DefaultActions {
			if action.TargetGroupARN != "" {
				tgNode := traceTargetGroup(ctx, cfg, action.TargetGroupARN, ec2Names)
				listenerNode.Children = append(listenerNode.Children, tgNode)
			}
		}
		lbNode.Children = append(lbNode.Children, listenerNode)
	}
	return lbNode
}

func traceApplicationLB(ctx context.Context, cfg sdkaws.Config, lbNode model.TraceNode, lb model.LoadBalancer, host, path string, ec2Names map[string]string) model.TraceNode {
	listeners, _ := FetchListeners(ctx, cfg, lb.ARN)
	for _, l := range listeners {
		listenerNode := model.TraceNode{
			Type: "Listener",
			Name: fmt.Sprintf("%s:%d", l.Protocol, l.Port),
		}

		rules, _ := FetchRules(ctx, cfg, l.ARN)
		sortRules(rules)
		var matchedRule *model.Rule
		var matchFound bool
		for _, r := range rules {
			if r.IsDefault {
				continue
			}
			match := true
			for _, cond := range r.Conditions {
				condMatch := false
				for _, val := range cond.Values {
					if cond.Field == "host-header" {
						if matchHost(host, val) {
							condMatch = true
							break
						}
					} else if cond.Field == "path-pattern" {
						if matchPath(path, val) {
							condMatch = true
							break
						}
					}
				}
				if !condMatch {
					match = false
					break
				}
			}
			if match {
				matchedRule = &r
				matchFound = true
				break
			}
		}

		if !matchFound {
			for _, action := range l.DefaultActions {
				if action.TargetGroupARN != "" {
					tgNode := traceTargetGroup(ctx, cfg, action.TargetGroupARN, ec2Names)
					tgNode.Name = "[Default] " + tgNode.Name
					listenerNode.Children = append(listenerNode.Children, tgNode)
				}
			}
		} else {
			condStr := ""
			for _, c := range matchedRule.Conditions {
				condStr += fmt.Sprintf("[%s:%s] ", c.Field, strings.Join(c.Values, ","))
			}
			ruleNode := model.TraceNode{
				Type: "Rule",
				Name: fmt.Sprintf("Priority %s: %s", matchedRule.Priority, condStr),
			}
			for _, action := range matchedRule.Actions {
				if action.TargetGroupARN != "" {
					tgNode := traceTargetGroup(ctx, cfg, action.TargetGroupARN, ec2Names)
					ruleNode.Children = append(ruleNode.Children, tgNode)
				}
			}
			listenerNode.Children = append(listenerNode.Children, ruleNode)
		}
		lbNode.Children = append(lbNode.Children, listenerNode)
	}
	return lbNode
}

func traceTargetGroup(ctx context.Context, cfg sdkaws.Config, tgARN string, ec2Names map[string]string) model.TraceNode {
	parts := strings.Split(tgARN, ":")
	tgName := "Unknown"
	if len(parts) > 5 {
		resParts := strings.Split(parts[5], "/")
		if len(resParts) > 1 {
			tgName = resParts[1]
		}
	}

	tgNode := model.TraceNode{
		Type: model.NodeTargetGroup,
		Name: tgName,
	}

	healths, _ := FetchTargetHealth(ctx, cfg, tgARN)
	hasHealthy := false
	for _, h := range healths {
		name := ec2Names[h.InstanceID]
		if name == "" {
			name = h.InstanceID
		}
		val := h.State
		if h.Reason != "" && h.Reason != "N/A" {
			val += " (" + h.Reason + ")"
		}

		status := "unhealthy"
		if h.State == "healthy" {
			status = "healthy"
			hasHealthy = true
		}

		tgNode.Children = append(tgNode.Children, model.TraceNode{
			Type:   "Target",
			Name:   name,
			Value:  val,
			Status: status,
		})
	}

	if hasHealthy {
		tgNode.Status = "healthy"
	} else {
		tgNode.Status = "unhealthy"
	}

	return tgNode
}

func matchHost(host, pattern string) bool {
	return matchPattern(host, pattern)
}

func matchPattern(text, pattern string) bool {
	rePattern := regexp.QuoteMeta(pattern)
	rePattern = strings.ReplaceAll(rePattern, `\*`, `.*`)
	rePattern = strings.ReplaceAll(rePattern, `\?`, `.`)
	rePattern = "^" + rePattern + "$"

	// Special case for ALB: if pattern ends with /*, it should also match the path without the trailing slash
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if text == prefix {
			return true
		}
	}

	matched, _ := regexp.MatchString(rePattern, text)
	return matched
}

func sortRules(rules []model.Rule) {
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].IsDefault {
			return false
		}
		if rules[j].IsDefault {
			return true
		}
		pi, _ := strconv.Atoi(rules[i].Priority)
		pj, _ := strconv.Atoi(rules[j].Priority)
		return pi < pj
	})
}
