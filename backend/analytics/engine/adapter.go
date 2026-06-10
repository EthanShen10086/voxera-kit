// Package engine provides a self-hosted in-memory analytics engine that
// implements analytics.Store. It is suitable for development, testing, and
// small-to-medium workloads. For production at scale, use the PostHog
// adapter backed by ClickHouse.
package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/analytics"
)

// Engine is an in-memory analytics store that computes funnels, retention,
// and path analysis on-the-fly over its stored events.
type Engine struct {
	mu       sync.RWMutex
	events   []analytics.Event
	profiles map[string]*analytics.UserProfile // keyed by UserID
	sessions map[string]struct{}               // set of seen session IDs
	aliases  map[string]string                 // previousID -> newID
}

// compile-time interface check.
var _ analytics.Store = (*Engine)(nil)

// New creates a ready-to-use in-memory Engine.
func New() *Engine {
	return &Engine{
		profiles: make(map[string]*analytics.UserProfile),
		sessions: make(map[string]struct{}),
		aliases:  make(map[string]string),
	}
}

// Track records a single event and updates the associated user profile.
func (e *Engine) Track(_ context.Context, event analytics.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.trackLocked(event)
	return nil
}

// TrackBatch records multiple events in one call.
func (e *Engine) TrackBatch(_ context.Context, events []analytics.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, ev := range events {
		e.trackLocked(ev)
	}
	return nil
}

func (e *Engine) trackLocked(event analytics.Event) {
	e.events = append(e.events, event)

	p := e.getOrCreateProfile(event.UserID, event.TenantID)
	p.EventCount++
	if p.FirstSeen.IsZero() || event.Timestamp.Before(p.FirstSeen) {
		p.FirstSeen = event.Timestamp
	}
	if event.Timestamp.After(p.LastSeen) {
		p.LastSeen = event.Timestamp
	}
	if event.SessionID != "" {
		if _, ok := e.sessions[event.UserID+"|"+event.SessionID]; !ok {
			e.sessions[event.UserID+"|"+event.SessionID] = struct{}{}
			p.SessionCount++
		}
	}

	updateAttribution(p, event)
}

func updateAttribution(p *analytics.UserProfile, ev analytics.Event) {
	tp := analytics.TouchPoint{
		Source:   ev.Ctx.UTMSource,
		Medium:   ev.Ctx.UTMMedium,
		Campaign: ev.Ctx.UTMCampaign,
		Term:     ev.Ctx.UTMTerm,
		Content:  ev.Ctx.UTMContent,
		Referrer: ev.Ctx.Referrer,
	}
	if tp.Source == "" && tp.Medium == "" && tp.Referrer == "" {
		return
	}
	if p.Attribution.FirstTouch.Source == "" && p.Attribution.FirstTouch.Referrer == "" {
		p.Attribution.FirstTouch = tp
	}
	p.Attribution.LastTouch = tp
}

// Identify sets or merges properties on a user profile.
func (e *Engine) Identify(_ context.Context, profile analytics.UserProfile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	p := e.getOrCreateProfile(profile.UserID, profile.TenantID)
	for k, v := range profile.Properties {
		p.Properties[k] = v
	}
	return nil
}

// Alias links a previous anonymous ID to a new identified user ID by
// copying all events and profile data from the old ID to the new one.
func (e *Engine) Alias(_ context.Context, previousID, newID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.aliases[previousID] = newID
	for i := range e.events {
		if e.events[i].UserID == previousID {
			e.events[i].UserID = newID
		}
	}
	if old, ok := e.profiles[previousID]; ok {
		merged := e.getOrCreateProfile(newID, old.TenantID)
		merged.EventCount += old.EventCount
		merged.SessionCount += old.SessionCount
		if !old.FirstSeen.IsZero() && (merged.FirstSeen.IsZero() || old.FirstSeen.Before(merged.FirstSeen)) {
			merged.FirstSeen = old.FirstSeen
		}
		if old.LastSeen.After(merged.LastSeen) {
			merged.LastSeen = old.LastSeen
		}
		for k, v := range old.Properties {
			if _, exists := merged.Properties[k]; !exists {
				merged.Properties[k] = v
			}
		}
		delete(e.profiles, previousID)
	}
	return nil
}

func (e *Engine) getOrCreateProfile(userID, tenantID string) *analytics.UserProfile {
	p, ok := e.profiles[userID]
	if !ok {
		p = &analytics.UserProfile{
			UserID:     userID,
			TenantID:   tenantID,
			Properties: make(map[string]any),
		}
		e.profiles[userID] = p
	}
	return p
}

// ---------------------------------------------------------------------------
// Query methods
// ---------------------------------------------------------------------------

// QueryFunnel computes conversion metrics for a multi-step funnel.
func (e *Engine) QueryFunnel(_ context.Context, q analytics.FunnelQuery) (*analytics.FunnelResult, error) {
	if len(q.Steps) == 0 {
		return &analytics.FunnelResult{}, nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	filtered := e.filterEvents(q.TenantID, q.From, q.To)

	// Group events by user, sorted by timestamp.
	byUser := groupByUser(filtered)

	type completion struct {
		stepTimes []time.Time
	}

	var completions []completion
	stepEntered := make([]int64, len(q.Steps))

	for _, events := range byUser {
		sort.Slice(events, func(i, j int) bool {
			return events[i].Timestamp.Before(events[j].Timestamp)
		})

		var stepIdx int
		var comp completion
		for _, ev := range events {
			if stepIdx >= len(q.Steps) {
				break
			}
			step := q.Steps[stepIdx]
			if ev.Name == step.Name && matchFilters(ev, step.Filters) {
				comp.stepTimes = append(comp.stepTimes, ev.Timestamp)
				if stepIdx == 0 || (q.WindowSec <= 0 ||
					ev.Timestamp.Sub(comp.stepTimes[0]).Seconds() <= float64(q.WindowSec)) {
					stepEntered[stepIdx]++
					stepIdx++
				} else {
					break
				}
			}
		}
		if len(comp.stepTimes) > 0 {
			completions = append(completions, comp)
		}
	}

	result := &analytics.FunnelResult{
		Steps: make([]analytics.FunnelStepResult, len(q.Steps)),
	}
	var totalTimes []int64
	for i, s := range q.Steps {
		result.Steps[i].Name = s.Name
		result.Steps[i].EnteredCount = stepEntered[i]
		if i > 0 {
			result.Steps[i].DroppedCount = stepEntered[i-1] - stepEntered[i]
			if stepEntered[i-1] > 0 {
				result.Steps[i].ConvertRate = float64(stepEntered[i]) / float64(stepEntered[i-1])
			}
		} else {
			result.Steps[i].ConvertRate = 1.0
		}
	}

	for _, comp := range completions {
		if len(comp.stepTimes) == len(q.Steps) {
			d := int64(comp.stepTimes[len(comp.stepTimes)-1].Sub(comp.stepTimes[0]).Seconds())
			totalTimes = append(totalTimes, d)
		}
	}

	// Compute per-step median times.
	for i := 1; i < len(q.Steps); i++ {
		var diffs []int64
		for _, comp := range completions {
			if len(comp.stepTimes) > i {
				d := int64(comp.stepTimes[i].Sub(comp.stepTimes[i-1]).Seconds())
				diffs = append(diffs, d)
			}
		}
		result.Steps[i].MedianTime = median(diffs)
	}

	if stepEntered[0] > 0 {
		result.OverallRate = float64(stepEntered[len(stepEntered)-1]) / float64(stepEntered[0])
	}
	result.MedianTime = median(totalTimes)

	return result, nil
}

// QueryRetention computes cohort retention over multiple periods.
func (e *Engine) QueryRetention(_ context.Context, q analytics.RetentionQuery) (*analytics.RetentionResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	filtered := e.filterEvents(q.TenantID, q.From, q.To)

	// Find users who performed the cohort event, grouped by period.
	cohortUsers := make(map[string]map[string]time.Time) // period -> userID -> first ts

	for _, ev := range filtered {
		if ev.Name != q.CohortEvent {
			continue
		}
		p := timeToPeriod(ev.Timestamp, q.Granularity)
		if cohortUsers[p] == nil {
			cohortUsers[p] = make(map[string]time.Time)
		}
		if existing, ok := cohortUsers[p][ev.UserID]; !ok || ev.Timestamp.Before(existing) {
			cohortUsers[p][ev.UserID] = ev.Timestamp
		}
	}

	// Index return events by user.
	returnEvents := make(map[string][]time.Time) // userID -> timestamps
	for _, ev := range filtered {
		if ev.Name != q.ReturnEvent {
			continue
		}
		returnEvents[ev.UserID] = append(returnEvents[ev.UserID], ev.Timestamp)
	}

	// Build sorted period list.
	var periods []string
	for p := range cohortUsers {
		periods = append(periods, p)
	}
	sort.Strings(periods)

	result := &analytics.RetentionResult{}
	for _, period := range periods {
		users := cohortUsers[period]
		row := analytics.CohortRow{
			Period:     period,
			CohortSize: int64(len(users)),
			Retained:   make([]int64, q.Periods),
			Rates:      make([]float64, q.Periods),
		}

		for userID, cohortTS := range users {
			rEvents := returnEvents[userID]
			for pi := 0; pi < q.Periods; pi++ {
				start := advancePeriod(cohortTS, q.Granularity, pi+1)
				end := advancePeriod(cohortTS, q.Granularity, pi+2)
				for _, rt := range rEvents {
					if !rt.Before(start) && rt.Before(end) {
						row.Retained[pi]++
						break
					}
				}
			}
		}
		for i := range row.Retained {
			if row.CohortSize > 0 {
				row.Rates[i] = float64(row.Retained[i]) / float64(row.CohortSize)
			}
		}
		result.Cohorts = append(result.Cohorts, row)
	}

	return result, nil
}

// QueryPath computes user flow / path analysis between events.
func (e *Engine) QueryPath(_ context.Context, q analytics.PathQuery) (*analytics.PathResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	maxSteps := q.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 5
	}

	filtered := e.filterEvents(q.TenantID, q.From, q.To)
	bySession := groupBySession(filtered)

	nodeCount := make(map[string]int64)
	edgeCount := make(map[string]int64) // "from->to"
	pathCount := make(map[string]int64) // "a|b|c"

	for _, events := range bySession {
		sort.Slice(events, func(i, j int) bool {
			return events[i].Timestamp.Before(events[j].Timestamp)
		})

		var steps []string
		for _, ev := range events {
			if len(steps) >= maxSteps {
				break
			}
			if q.StartEvent != "" && len(steps) == 0 && ev.Name != q.StartEvent {
				continue
			}
			steps = append(steps, ev.Name)
			nodeCount[ev.Name]++
			if len(steps) > 1 {
				key := steps[len(steps)-2] + "->" + ev.Name
				edgeCount[key]++
			}
			if q.EndEvent != "" && ev.Name == q.EndEvent {
				break
			}
		}
		if len(steps) > 0 {
			pathKey := strings.Join(steps, "|")
			pathCount[pathKey]++
		}
	}

	result := &analytics.PathResult{}
	for name, cnt := range nodeCount {
		result.Nodes = append(result.Nodes, analytics.PathNode{Name: name, Count: cnt})
	}
	sort.Slice(result.Nodes, func(i, j int) bool {
		return result.Nodes[i].Count > result.Nodes[j].Count
	})

	for key, cnt := range edgeCount {
		parts := strings.SplitN(key, "->", 2)
		pct := float64(0)
		if nodeCount[parts[0]] > 0 {
			pct = float64(cnt) / float64(nodeCount[parts[0]])
		}
		result.Edges = append(result.Edges, analytics.PathEdge{
			From:    parts[0],
			To:      parts[1],
			Count:   cnt,
			Percent: pct,
		})
	}
	sort.Slice(result.Edges, func(i, j int) bool {
		return result.Edges[i].Count > result.Edges[j].Count
	})

	for key, cnt := range pathCount {
		if q.MinCount > 0 && cnt < q.MinCount {
			continue
		}
		result.Paths = append(result.Paths, analytics.PathSequence{
			Steps: strings.Split(key, "|"),
			Count: cnt,
		})
	}
	sort.Slice(result.Paths, func(i, j int) bool {
		return result.Paths[i].Count > result.Paths[j].Count
	})

	return result, nil
}

// QueryEvents returns raw events matching the given filters.
func (e *Engine) QueryEvents(_ context.Context, q analytics.EventQuery) ([]analytics.Event, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var results []analytics.Event
	nameSet := make(map[string]struct{}, len(q.Names))
	for _, n := range q.Names {
		nameSet[n] = struct{}{}
	}

	for _, ev := range e.events {
		if q.UserID != "" && ev.UserID != q.UserID {
			continue
		}
		if q.TenantID != "" && ev.TenantID != q.TenantID {
			continue
		}
		if len(nameSet) > 0 {
			if _, ok := nameSet[ev.Name]; !ok {
				continue
			}
		}
		if !q.From.IsZero() && ev.Timestamp.Before(q.From) {
			continue
		}
		if !q.To.IsZero() && !ev.Timestamp.Before(q.To) {
			continue
		}
		results = append(results, ev)
	}

	if q.OrderBy == "timestamp_desc" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Timestamp.After(results[j].Timestamp)
		})
	} else {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Timestamp.Before(results[j].Timestamp)
		})
	}

	if q.Offset > 0 && q.Offset < len(results) {
		results = results[q.Offset:]
	} else if q.Offset >= len(results) {
		return nil, nil
	}
	if q.Limit > 0 && q.Limit < len(results) {
		results = results[:q.Limit]
	}

	return results, nil
}

// QueryUserProfile returns the aggregated profile for a single user.
func (e *Engine) QueryUserProfile(_ context.Context, userID string) (*analytics.UserProfile, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	p, ok := e.profiles[userID]
	if !ok {
		return nil, fmt.Errorf("user %q not found", userID)
	}

	counts := make(map[string]int64)
	for _, ev := range e.events {
		if ev.UserID == userID {
			counts[ev.Name]++
		}
	}
	var topEvents []analytics.EventCount
	for name, cnt := range counts {
		topEvents = append(topEvents, analytics.EventCount{Name: name, Count: cnt})
	}
	sort.Slice(topEvents, func(i, j int) bool {
		return topEvents[i].Count > topEvents[j].Count
	})

	cp := *p
	cp.TopEvents = topEvents
	return &cp, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (e *Engine) filterEvents(tenantID string, from, to time.Time) []analytics.Event {
	var out []analytics.Event
	for _, ev := range e.events {
		if tenantID != "" && ev.TenantID != tenantID {
			continue
		}
		if !from.IsZero() && ev.Timestamp.Before(from) {
			continue
		}
		if !to.IsZero() && !ev.Timestamp.Before(to) {
			continue
		}
		out = append(out, ev)
	}
	return out
}

func groupByUser(events []analytics.Event) map[string][]analytics.Event {
	m := make(map[string][]analytics.Event)
	for _, ev := range events {
		m[ev.UserID] = append(m[ev.UserID], ev)
	}
	return m
}

func groupBySession(events []analytics.Event) map[string][]analytics.Event {
	m := make(map[string][]analytics.Event)
	for _, ev := range events {
		key := ev.SessionID
		if key == "" {
			key = ev.UserID
		}
		m[key] = append(m[key], ev)
	}
	return m
}

func matchFilters(ev analytics.Event, filters map[string]any) bool {
	for k, v := range filters {
		if ev.Properties[k] != v {
			return false
		}
	}
	return true
}

func median(vals []int64) int64 {
	if len(vals) == 0 {
		return 0
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	mid := len(vals) / 2
	if len(vals)%2 == 0 {
		return (vals[mid-1] + vals[mid]) / 2
	}
	return vals[mid]
}

func timeToPeriod(t time.Time, g analytics.Granularity) string {
	switch g {
	case analytics.GranularityWeek:
		y, w := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", y, w)
	case analytics.GranularityMonth:
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02")
	}
}

func advancePeriod(t time.Time, g analytics.Granularity, n int) time.Time {
	switch g {
	case analytics.GranularityWeek:
		return t.AddDate(0, 0, 7*n)
	case analytics.GranularityMonth:
		return t.AddDate(0, n, 0)
	default:
		return t.AddDate(0, 0, n)
	}
}
