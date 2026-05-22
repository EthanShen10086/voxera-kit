package analytics

import "time"

// PathQuery defines parameters for user flow / path analysis.
type PathQuery struct {
	// TenantID scopes the analysis to a specific tenant.
	TenantID string
	// StartEvent restricts paths to those beginning with this event.
	// An empty value matches any entry event.
	StartEvent string
	// EndEvent restricts paths to those ending with this event.
	// An empty value matches any exit event.
	EndEvent string
	// From is the inclusive start of the analysis window.
	From time.Time
	// To is the exclusive end of the analysis window.
	To time.Time
	// MaxSteps is the maximum path depth to analyze (defaults to 5).
	MaxSteps int
	// MinCount is the minimum number of occurrences required to include a path.
	MinCount int64
}

// PathResult holds the computed path / flow analysis.
type PathResult struct {
	// Nodes lists the distinct events that appear in the path graph.
	Nodes []PathNode
	// Edges lists the transitions between events.
	Edges []PathEdge
	// Paths lists the specific end-to-end sequences users took.
	Paths []PathSequence
}

// PathNode represents an event node in the path graph.
type PathNode struct {
	// Name is the event name.
	Name string
	// Count is the total number of times users passed through this node.
	Count int64
}

// PathEdge represents a directed transition between two events.
type PathEdge struct {
	// From is the source event name.
	From string
	// To is the destination event name.
	To string
	// Count is the number of users who made this transition.
	Count int64
	// Percent is the fraction of users at the From node who transitioned here.
	Percent float64
}

// PathSequence represents a specific end-to-end path taken by users.
type PathSequence struct {
	// Steps lists the ordered event names in this path.
	Steps []string
	// Count is the number of users who followed this exact sequence.
	Count int64
}
