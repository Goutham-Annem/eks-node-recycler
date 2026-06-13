# eks-node-recycler

> Gracefully recycle stale or underutilized EKS nodes — keep your cluster fresh without downtime.

[![Go Report Card](https://goreportcard.com/badge/github.com/goutham-annem/eks-node-recycler)](https://goreportcard.com/report/github.com/goutham-annem/eks-node-recycler)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Problem

EKS nodes accumulate drift over time: stale AMIs, leaked memory, orphaned processes, and wasted spend from idle underutilized nodes. Manual recycling is error-prone and doesn't scale.

## Solution

`eks-node-recycler` is a Go CLI that:
- Identifies nodes older than a configurable threshold (default: 7 days)
- Detects chronically idle nodes via CloudWatch CPU metrics
- Gracefully drains each node (`kubectl drain`) before terminating via EC2 API
- Triggers ASG replacement with fresh, patched instances
- Supports dry-run mode for safe previewing

## Installation

```bash
go install github.com/goutham-annem/eks-node-recycler@latest
```

Or download a pre-built binary from [Releases](../../releases).

## Usage

### List eligible nodes
```bash
eks-node-recycler list --cluster my-prod-cluster --region us-east-1
```

### Recycle stale nodes (dry run first!)
```bash
eks-node-recycler recycle \
  --cluster my-prod-cluster \
  --region us-east-1 \
  --max-age 168h \
  --cpu-threshold 10 \
  --dry-run
```

### Actually recycle
```bash
eks-node-recycler recycle \
  --cluster my-prod-cluster \
  --region us-east-1 \
  --max-age 168h \
  --cpu-threshold 10
```

## Required IAM Permissions

```json
{
  "Effect": "Allow",
  "Action": [
    "eks:ListNodegroups",
    "eks:DescribeNodegroup",
    "ec2:DescribeInstances",
    "ec2:TerminateInstances",
    "autoscaling:DescribeAutoScalingGroups",
    "cloudwatch:GetMetricStatistics"
  ],
  "Resource": "*"
}
```

## Configuration Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--cluster` | required | EKS cluster name |
| `--region` | `us-east-1` | AWS region |
| `--max-age` | `168h` | Node age before recycling |
| `--cpu-threshold` | `10.0` | Avg CPU % below which node is idle |
| `--dry-run` | `false` | Preview without terminating |

## Architecture

```
eks-node-recycler
├── List node groups (EKS API)
├── Describe instances (EC2 API)
├── Pull 7-day avg CPU (CloudWatch)
├── Filter: age > max-age OR cpu < threshold
├── kubectl drain (graceful eviction)
└── ec2:TerminateInstances → ASG replaces
```

## Contributing

PRs welcome. Please run `make lint test` before submitting.

## License

MIT — by [Goutham Annem](https://linkedin.com/in/goutham-annem)
