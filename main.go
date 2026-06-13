package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/spf13/cobra"
)

var (
	clusterName    string
	region         string
	maxAge         time.Duration
	cpuThreshold   float64
	dryRun         bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "eks-node-recycler",
		Short: "Gracefully recycle stale or underutilized EKS nodes",
		Long: `eks-node-recycler identifies EKS node groups with stale nodes
(based on age or CPU utilization) and gracefully drains + terminates them,
triggering ASG replacement with fresh, up-to-date instances.`,
	}

	recycleCmd := &cobra.Command{
		Use:   "recycle",
		Short: "Scan and recycle eligible nodes",
		RunE:  runRecycle,
	}

	recycleCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "EKS cluster name (required)")
	recycleCmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "AWS region")
	recycleCmd.Flags().DurationVar(&maxAge, "max-age", 7*24*time.Hour, "Max node age before recycling (e.g. 168h)")
	recycleCmd.Flags().Float64Var(&cpuThreshold, "cpu-threshold", 10.0, "Avg CPU % below which node is considered idle")
	recycleCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing")
	recycleCmd.MarkFlagRequired("cluster")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all nodes and their recycling eligibility",
		RunE:  runList,
	}
	listCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "EKS cluster name (required)")
	listCmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "AWS region")
	listCmd.Flags().DurationVar(&maxAge, "max-age", 7*24*time.Hour, "Max node age before recycling")
	listCmd.Flags().Float64Var(&cpuThreshold, "cpu-threshold", 10.0, "Avg CPU % idle threshold")
	listCmd.MarkFlagRequired("cluster")

	rootCmd.AddCommand(recycleCmd, listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRecycle(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}

	eksClient := eks.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	nodes, err := getEligibleNodes(ctx, eksClient, ec2Client)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		fmt.Println("No nodes eligible for recycling.")
		return nil
	}

	for _, node := range nodes {
		fmt.Printf("[%s] Age: %s | Avg CPU: %.1f%%\n", node.InstanceID, node.Age.Round(time.Hour), node.AvgCPU)
		if dryRun {
			fmt.Printf("  [DRY-RUN] Would drain and terminate %s\n", node.InstanceID)
			continue
		}
		if err := drainAndTerminate(ctx, ec2Client, node.InstanceID); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR terminating %s: %v\n", node.InstanceID, err)
		} else {
			fmt.Printf("  Terminated %s successfully\n", node.InstanceID)
		}
	}
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}

	eksClient := eks.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	nodes, err := getEligibleNodes(ctx, eksClient, ec2Client)
	if err != nil {
		return err
	}

	fmt.Printf("%-20s %-12s %-10s %-10s\n", "INSTANCE-ID", "AGE", "AVG-CPU%", "ELIGIBLE")
	fmt.Println("--------------------------------------------------------------")
	for _, node := range nodes {
		fmt.Printf("%-20s %-12s %-10.1f %-10v\n",
			node.InstanceID,
			node.Age.Round(time.Hour).String(),
			node.AvgCPU,
			node.Eligible,
		)
	}
	return nil
}

type NodeInfo struct {
	InstanceID string
	Age        time.Duration
	AvgCPU    float64
	Eligible  bool
}

// getEligibleNodes queries EKS node groups and returns nodes past max-age or below cpu-threshold.
func getEligibleNodes(ctx context.Context, eksClient *eks.Client, ec2Client *ec2.Client) ([]NodeInfo, error) {
	// In a real implementation: list node groups, describe instances, pull CloudWatch metrics
	// This scaffold demonstrates the structure — integrate aws-sdk-go-v2 calls here.
	fmt.Printf("Scanning cluster: %s (region: %s)\n", clusterName, region)
	fmt.Printf("Max age: %s | CPU threshold: %.1f%%\n\n", maxAge, cpuThreshold)
	return []NodeInfo{}, nil
}

func drainAndTerminate(ctx context.Context, ec2Client *ec2.Client, instanceID string) error {
	// 1. kubectl drain --ignore-daemonsets --delete-emptydir-data <node>
	// 2. ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{InstanceIds: []string{instanceID}})
	return nil
}
