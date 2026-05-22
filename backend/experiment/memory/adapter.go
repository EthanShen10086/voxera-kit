// Package memory provides an in-memory implementation of the experiment
// Manager interface with deterministic variant assignment and statistical
// significance testing.
package memory

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/EthanShen10086/voxera-kit/experiment"
)

type metricEvent struct {
	userID    string
	metricKey string
	value     float64
}

// Store is an in-memory experiment store that supports variant assignment,
// metric recording, and statistical significance computation.
type Store struct {
	mu          sync.RWMutex
	configs     map[string]experiment.Config
	assignments map[string]map[string]experiment.Assignment // expKey → userID → Assignment
	events      map[string][]metricEvent                    // expKey → events
}

// NewStore creates a new in-memory experiment store.
func NewStore() *Store {
	return &Store{
		configs:     make(map[string]experiment.Config),
		assignments: make(map[string]map[string]experiment.Assignment),
		events:      make(map[string][]metricEvent),
	}
}

// Create registers a new experiment definition in the store.
func (s *Store) Create(_ context.Context, cfg experiment.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.configs[cfg.Key]; exists {
		return fmt.Errorf("experiment %q already exists", cfg.Key)
	}

	if cfg.Status == "" {
		cfg.Status = experiment.StatusDraft
	}
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = time.Now()
	}

	s.configs[cfg.Key] = cfg
	s.assignments[cfg.Key] = make(map[string]experiment.Assignment)

	return nil
}

// Get retrieves an experiment by its unique key.
func (s *Store) Get(_ context.Context, key string) (*experiment.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.configs[key]
	if !ok {
		return nil, fmt.Errorf("experiment %q not found", key)
	}

	return &cfg, nil
}

// List returns all experiments matching the given status filter. An empty
// status returns all experiments.
func (s *Store) List(_ context.Context, status experiment.Status) ([]experiment.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []experiment.Config
	for _, cfg := range s.configs {
		if status == "" || cfg.Status == status {
			result = append(result, cfg)
		}
	}

	return result, nil
}

// Start transitions an experiment to the running state.
func (s *Store) Start(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[key]
	if !ok {
		return fmt.Errorf("experiment %q not found", key)
	}

	cfg.Status = experiment.StatusRunning
	cfg.StartedAt = time.Now()
	s.configs[key] = cfg

	return nil
}

// Stop ends an experiment and marks it as complete.
func (s *Store) Stop(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[key]
	if !ok {
		return fmt.Errorf("experiment %q not found", key)
	}

	cfg.Status = experiment.StatusComplete
	cfg.EndedAt = time.Now()
	s.configs[key] = cfg

	return nil
}

// Assign assigns a user to a variant deterministically using a hash of the
// experiment key and user ID. If the user is already assigned, the existing
// assignment is returned.
func (s *Store) Assign(_ context.Context, key string, userID string) (*experiment.Assignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, ok := s.configs[key]
	if !ok {
		return nil, fmt.Errorf("experiment %q not found", key)
	}

	if a, exists := s.assignments[key][userID]; exists {
		return &a, nil
	}

	variant := pickVariant(cfg.Variants, key, userID)

	a := experiment.Assignment{
		ExperimentKey: key,
		UserID:        userID,
		VariantKey:    variant,
		AssignedAt:    time.Now(),
	}
	s.assignments[key][userID] = a

	return &a, nil
}

// GetAssignment retrieves the current variant assignment for a user.
func (s *Store) GetAssignment(_ context.Context, key string, userID string) (*experiment.Assignment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, ok := s.assignments[key]
	if !ok {
		return nil, fmt.Errorf("experiment %q not found", key)
	}

	a, ok := users[userID]
	if !ok {
		return nil, nil //nolint:nilnil // nil assignment means the user has not been assigned yet
	}

	return &a, nil
}

// RecordMetric records a metric event for a user in an experiment.
func (s *Store) RecordMetric(_ context.Context, key string, userID string, metricKey string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.configs[key]; !ok {
		return fmt.Errorf("experiment %q not found", key)
	}

	s.events[key] = append(s.events[key], metricEvent{
		userID:    userID,
		metricKey: metricKey,
		value:     value,
	})

	return nil
}

// GetResults computes current experiment results with statistical significance
// analysis using z-tests for proportions and t-test approximations for
// continuous metrics.
func (s *Store) GetResults(_ context.Context, key string) (*experiment.Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.configs[key]
	if !ok {
		return nil, fmt.Errorf("experiment %q not found", key)
	}

	assignments := s.assignments[key]
	events := s.events[key]

	variantUsers := make(map[string]map[string]bool)
	for _, a := range assignments {
		if variantUsers[a.VariantKey] == nil {
			variantUsers[a.VariantKey] = make(map[string]bool)
		}
		variantUsers[a.VariantKey][a.UserID] = true
	}

	userVariant := make(map[string]string)
	for _, a := range assignments {
		userVariant[a.UserID] = a.VariantKey
	}

	type metricData struct {
		values []float64
		users  map[string]float64
	}

	perVariantMetric := make(map[string]map[string]*metricData) // variantKey → metricKey → data
	for _, v := range cfg.Variants {
		perVariantMetric[v.Key] = make(map[string]*metricData)
		for _, m := range cfg.Metrics {
			perVariantMetric[v.Key][m.Key] = &metricData{
				users: make(map[string]float64),
			}
		}
	}

	for _, e := range events {
		vk, ok := userVariant[e.userID]
		if !ok {
			continue
		}
		md, ok := perVariantMetric[vk][e.metricKey]
		if !ok {
			continue
		}
		md.values = append(md.values, e.value)
		md.users[e.userID] += e.value
	}

	var controlKey string
	for _, v := range cfg.Variants {
		if v.IsControl {
			controlKey = v.Key
			break
		}
	}

	var results []experiment.MetricResult

	for _, m := range cfg.Metrics {
		for _, v := range cfg.Variants {
			md := perVariantMetric[v.Key][m.Key]
			sampleSize := int64(len(variantUsers[v.Key]))

			mr := experiment.MetricResult{
				MetricKey:  m.Key,
				VariantKey: v.Key,
				SampleSize: sampleSize,
			}

			if sampleSize > 0 {
				switch m.Type {
				case experiment.MetricConversion:
					conversions := int64(len(md.users))
					mr.ConvRate = float64(conversions) / float64(sampleSize)
					mr.Mean = mr.ConvRate
					mr.Variance = mr.ConvRate * (1 - mr.ConvRate)
				default:
					var sum, sumSq float64
					for _, val := range md.values {
						sum += val
						sumSq += val * val
					}
					n := float64(len(md.values))
					if n > 0 {
						mr.Mean = sum / n
						if n > 1 {
							mr.Variance = (sumSq - sum*sum/n) / (n - 1)
						}
					}
				}
			}

			results = append(results, mr)
		}
	}

	winner := ""
	for _, m := range cfg.Metrics {
		if controlKey == "" {
			break
		}

		var controlResult *experiment.MetricResult
		for i := range results {
			if results[i].MetricKey == m.Key && results[i].VariantKey == controlKey {
				controlResult = &results[i]
				break
			}
		}
		if controlResult == nil {
			continue
		}

		for i := range results {
			r := &results[i]
			if r.MetricKey != m.Key || r.VariantKey == controlKey {
				continue
			}

			if r.SampleSize < 2 || controlResult.SampleSize < 2 {
				continue
			}

			var conf float64
			switch m.Type {
			case experiment.MetricConversion:
				conf = zTestProportion(
					controlResult.SampleSize, int64(math.Round(controlResult.ConvRate*float64(controlResult.SampleSize))),
					r.SampleSize, int64(math.Round(r.ConvRate*float64(r.SampleSize))),
				)
			default:
				conf = tTestTwoSample(
					controlResult.Mean, controlResult.Variance, controlResult.SampleSize,
					r.Mean, r.Variance, r.SampleSize,
				)
			}

			r.Confidence = conf
			r.IsSignificant = conf > 0.95

			if r.IsSignificant && r.Mean > controlResult.Mean && (winner == "" || r.Mean > controlResult.Mean) {
				winner = r.VariantKey
			}
		}
	}

	totalUsers := int64(0)
	for _, users := range variantUsers {
		totalUsers += int64(len(users))
	}

	return &experiment.Result{
		ExperimentKey: key,
		Status:        cfg.Status,
		TotalUsers:    totalUsers,
		Metrics:       results,
		Winner:        winner,
		StartedAt:     cfg.StartedAt,
		AnalyzedAt:    time.Now(),
	}, nil
}

// pickVariant selects a variant deterministically by hashing the experiment key
// and user ID, then mapping the hash into the weighted variant distribution.
func pickVariant(variants []experiment.Variant, expKey, userID string) string {
	totalWeight := 0
	for _, v := range variants {
		totalWeight += v.Weight
	}
	if totalWeight == 0 {
		if len(variants) > 0 {
			return variants[0].Key
		}
		return ""
	}

	h := sha256.New()
	h.Write([]byte(expKey + ":" + userID))
	sum := h.Sum(nil)
	val := binary.BigEndian.Uint32(sum[:4])
	bucket := int(val % uint32(totalWeight))

	cumulative := 0
	for _, v := range variants {
		cumulative += v.Weight
		if bucket < cumulative {
			return v.Key
		}
	}

	return variants[len(variants)-1].Key
}

// zTestProportion computes confidence for comparing two conversion rates using
// a pooled proportion z-test. Returns a confidence value in [0, 1].
func zTestProportion(n1, conv1, n2, conv2 int64) float64 {
	if n1 == 0 || n2 == 0 {
		return 0
	}

	p1 := float64(conv1) / float64(n1)
	p2 := float64(conv2) / float64(n2)
	pooled := float64(conv1+conv2) / float64(n1+n2)

	se := math.Sqrt(pooled * (1 - pooled) * (1/float64(n1) + 1/float64(n2)))
	if se == 0 {
		return 0
	}

	z := math.Abs(p1-p2) / se

	return normalCDF(z)
}

// tTestTwoSample computes confidence for comparing two sample means using
// Welch's t-test approximation. Returns a confidence value in [0, 1].
func tTestTwoSample(mean1, var1 float64, n1 int64, mean2, var2 float64, n2 int64) float64 {
	if n1 < 2 || n2 < 2 {
		return 0
	}

	se := math.Sqrt(var1/float64(n1) + var2/float64(n2))
	if se == 0 {
		return 0
	}

	t := math.Abs(mean1-mean2) / se

	return normalCDF(t)
}

// normalCDF approximates the one-sided cumulative distribution function for the
// standard normal distribution using the Abramowitz and Stegun approximation.
// The returned value represents 1 - p (confidence), not the raw CDF.
func normalCDF(z float64) float64 {
	const (
		a1 = 0.254829592
		a2 = -0.284496736
		a3 = 1.421413741
		a4 = -1.453152027
		a5 = 1.061405429
		p  = 0.3275911
	)

	sign := 1.0
	if z < 0 {
		sign = -1.0
	}
	z = math.Abs(z) / math.Sqrt(2)

	t := 1.0 / (1.0 + p*z)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-z*z)

	cdf := 0.5 * (1.0 + sign*y)

	return 1.0 - 2.0*(1.0-cdf)
}
