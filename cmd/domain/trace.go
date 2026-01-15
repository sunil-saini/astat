package domain

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/sunil-saini/astat/internal/aws"
	"github.com/sunil-saini/astat/internal/model"
)

var TraceCmd = &cobra.Command{
	Use:   "trace [domain/uri]",
	Short: "Trace a domain or URI request flow through AWS infrastructure",
	Long: `Trace a domain or URI request flow through AWS infrastructure.
This feature performs a deep inspection of your AWS setup to show
how a request for the given domain or URI is handled.

The trace includes:
- External DNS resolution (IPs, CNAMEs)
- Route53 Hosted Zones and Record Sets (A, CNAME, Alias)
- CloudFront Distributions (Aliases, Origins, Behaviors)
- Application, Network, and Classic Load Balancers
- Target Groups and Health Checks
- Filtered ALB Rules and Conditions
- Lambda Functions and EC2 Instance names`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		spinner, _ := pterm.DefaultSpinner.
			WithSequence("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏").
			WithRemoveWhenDone(true).
			Start(pterm.Cyan(fmt.Sprintf("Tracing %s...", domain)))

		ctx := cmd.Context()
		cfg, err := aws.LoadConfig(ctx)
		if err != nil {
			spinner.Fail(err)
			return err
		}

		result, err := aws.TraceDomain(ctx, cfg, domain)
		if err != nil {
			spinner.Fail(err)
			return err
		}

		spinner.Stop()
		pterm.Success.Printf("Trace complete for %s\n", pterm.Bold.Sprint(domain))
		pterm.Println()

		if len(result.Hops) == 0 {
			pterm.Warning.Println("No path found for domain")
			return nil
		}

		root := pterm.TreeNode{
			Text: pterm.Bold.Sprint(domain),
		}

		for _, hop := range result.Hops {
			root.Children = append(root.Children, convertToPTermNode(hop))
		}

		pterm.DefaultTree.WithRoot(root).Render()
		return nil
	},
}

func convertToPTermNode(node model.TraceNode) pterm.TreeNode {
	name := pterm.Bold.Sprint(node.Name)
	val := node.Value
	status := node.Status

	// Apply coloring based on status
	switch status {
	case "healthy":
		name = pterm.LightGreen(node.Name)
		if val != "" {
			val = pterm.LightGreen(val)
		}
	case "unhealthy":
		name = pterm.LightRed(node.Name)
		if val != "" {
			val = pterm.LightRed(val)
		}
	default:
		if val != "" {
			val = pterm.Yellow(val)
		}
	}

	text := fmt.Sprintf("[%s] %s", pterm.Cyan(node.Type), name)
	if val != "" {
		text += fmt.Sprintf(" -> %s", val)
	}

	pnode := pterm.TreeNode{
		Text: text,
	}

	for _, child := range node.Children {
		pnode.Children = append(pnode.Children, convertToPTermNode(child))
	}

	return pnode
}

func init() {
	DomainCmd.AddCommand(TraceCmd)
}
