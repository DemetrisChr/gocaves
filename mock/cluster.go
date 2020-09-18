package mock

import (
	"time"

	"github.com/couchbaselabs/gocaves/mock/mocktime"
	"github.com/google/uuid"
)

// Cluster represents an instance of a mock cluster
type Cluster struct {
	id              string
	enabledFeatures []ClusterFeature
	numVbuckets     uint
	chrono          *mocktime.Chrono
	replicaLatency  time.Duration

	buckets []*Bucket
	nodes   []*ClusterNode

	currentConfig []byte
}

// NewClusterOptions allows the specification of initial options for a new cluster.
type NewClusterOptions struct {
	Chrono          *mocktime.Chrono
	EnabledFeatures []ClusterFeature
	NumVbuckets     uint
	InitialNode     NewNodeOptions
	ReplicaLatency  time.Duration
}

// NewCluster instantiates a new cluster instance.
func NewCluster(opts NewClusterOptions) (*Cluster, error) {
	if opts.Chrono == nil {
		opts.Chrono = &mocktime.Chrono{}
	}
	if opts.NumVbuckets == 0 {
		opts.NumVbuckets = 1024
	}
	if opts.ReplicaLatency == 0 {
		opts.ReplicaLatency = 50 * time.Millisecond
	}

	cluster := &Cluster{
		id:              uuid.New().String(),
		enabledFeatures: opts.EnabledFeatures,
		numVbuckets:     opts.NumVbuckets,
		chrono:          opts.Chrono,
		replicaLatency:  opts.ReplicaLatency,
		buckets:         nil,
		nodes:           nil,
	}

	// Since it doesn't make sense to have no nodes in a cluster, we force
	// one to be added here at creation time.  Theoretically nothing will break
	// if there are no nodes in the cluster, but this might change in the future.
	cluster.AddNode(opts.InitialNode)

	return cluster, nil
}

// ID returns the uuid of this cluster.
func (c *Cluster) ID() string {
	return c.id
}

func (c *Cluster) nodeUuids() []string {
	var out []string
	for _, node := range c.nodes {
		out = append(out, node.ID())
	}
	return out
}

// AddNode will add a new node to a cluster.
func (c *Cluster) AddNode(opts NewNodeOptions) (*ClusterNode, error) {
	node, err := NewClusterNode(c, opts)
	if err != nil {
		return nil, err
	}

	c.nodes = append(c.nodes, node)

	c.updateConfig()
	return node, nil
}

// AddBucket will add a new bucket to a cluster.
func (c *Cluster) AddBucket(opts NewBucketOptions) (*Bucket, error) {
	bucket, err := newBucket(c, opts)
	if err != nil {
		return nil, err
	}

	// Do an initial rebalance for the nodes we currently have
	bucket.UpdateVbMap(c.nodeUuids())

	c.buckets = append(c.buckets, bucket)

	c.updateConfig()
	return bucket, nil
}

// GetBucket will return a specific bucket from the cluster.
func (c *Cluster) GetBucket(name string) *Bucket {
	for _, bucket := range c.buckets {
		if bucket.Name() == name {
			return bucket
		}
	}
	return nil
}

// IsFeatureEnabled will indicate whether this cluster has a specific feature enabled.
func (c *Cluster) IsFeatureEnabled(feature ClusterFeature) bool {
	for _, supportedFeature := range c.enabledFeatures {
		if supportedFeature == feature {
			return true
		}
	}

	return false
}

func (c *Cluster) updateConfig() {
	c.currentConfig = nil
}
