package validation

import (
	"encoding/json"
	"flag"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

const (
	// LocalRateLimitStrat represents a ingestion rate limiting strategy that enforces the limit
	// on a per distributor basis.
	//
	// The actual effective rate limit will be N times higher, where N is the number of distributor replicas.
	LocalIngestionRateStrategy = "local"

	// GlobalRateLimitStrat represents a ingestion rate limiting strategy that enforces the rate
	// limiting globally, configuring a per-distributor local rate limiter as "ingestion_rate / N",
	// where N is the number of distributor replicas (it's automatically adjusted if the
	// number of replicas change).
	//
	// The global strategy requires the distributors to form their own ring, which
	// is used to keep track of the current number of healthy distributor replicas.
	GlobalIngestionRateStrategy = "global"

	bytesInMB = 1048576

	defaultPerStreamRateLimit  = 3 << 20 // 3MB
	defaultPerStreamBurstLimit = 5 * defaultPerStreamRateLimit

	DefaultPerTenantQueryTimeout = "1m"
)

// Limits describe all the limits for users; can be used to describe global default
// limits via flags, or per-user limits via yaml config.
// NOTE: we use custom `model.Duration` instead of standard `time.Duration` because,
// to support user-friendly duration format (e.g: "1h30m45s") in JSON value.
type Limits struct {
	// Distributor enforced limits.
	IngestionRateMB        float64 `yaml:"ingestion_rate_mb" json:"ingestion_rate_mb"`
	IngestionBurstSizeMB   float64 `yaml:"ingestion_burst_size_mb" json:"ingestion_burst_size_mb"`
	MaxLabelNameLength     int     `yaml:"max_label_name_length" json:"max_label_name_length"`
	MaxLabelValueLength    int     `yaml:"max_label_value_length" json:"max_label_value_length"`
	MaxLabelNamesPerSeries int     `yaml:"max_label_names_per_series" json:"max_label_names_per_series"`

	// Ingester enforced limits.
	MaxLocalStreamsPerUser  int `yaml:"max_streams_per_user" json:"max_streams_per_user"`
	MaxGlobalStreamsPerUser int `yaml:"max_global_streams_per_user" json:"max_global_streams_per_user"`

	// Querier enforced limits.
	MaxQueryLookback    model.Duration `yaml:"max_query_lookback" json:"max_query_lookback"`
	MaxQueryLength      model.Duration `yaml:"max_query_length" json:"max_query_length"`
	MaxQueryParallelism int            `yaml:"max_query_parallelism" json:"max_query_parallelism"`

	// Config for overrides, convenient if it goes here.
	PerTenantOverrideConfig string         `yaml:"per_tenant_override_config" json:"per_tenant_override_config"`
	PerTenantOverridePeriod model.Duration `yaml:"per_tenant_override_period" json:"per_tenant_override_period"`
}

// LimitError are errors that do not comply with the limits specified.
type LimitError string

func (e LimitError) Error() string {
	return string(e)
}

// RegisterFlags adds the flags required to config this to the given FlagSet
func (l *Limits) RegisterFlags(f *flag.FlagSet) {
	f.Float64Var(&l.IngestionRateMB, "distributor.ingestion-rate-limit-mb", 4, "Per-user ingestion rate limit in sample size per second. Units in MB.")
	f.Float64Var(&l.IngestionBurstSizeMB, "distributor.ingestion-burst-size-mb", 6, "Per-user allowed ingestion burst size (in sample size). Units in MB. The burst size refers to the per-distributor local rate limiter even in the case of the 'global' strategy, and should be set at least to the maximum logs size expected in a single push request.")
	f.IntVar(&l.MaxLabelNameLength, "validation.max-length-label-name", 1024, "Maximum length accepted for label names.")
	f.IntVar(&l.MaxLabelValueLength, "validation.max-length-label-value", 2048, "Maximum length accepted for label value. This setting also applies to the metric name.")
	f.IntVar(&l.MaxLabelNamesPerSeries, "validation.max-label-names-per-series", 30, "Maximum number of label names per series.")

	f.IntVar(&l.MaxLocalStreamsPerUser, "ingester.max-streams-per-user", 0, "Maximum number of active streams per user, per ingester. 0 to disable.")
	f.IntVar(&l.MaxGlobalStreamsPerUser, "ingester.max-global-streams-per-user", 5000, "Maximum number of active streams per user, across the cluster. 0 to disable. When the global limit is enabled, each ingester is configured with a dynamic local limit based on the replication factor and the current number of healthy ingesters, and is kept updated whenever the number of ingesters change.")

	_ = l.MaxQueryLength.Set("721h")
	f.Var(&l.MaxQueryLength, "store.max-query-length", "The limit to length of chunk store queries. 0 to disable.")

	_ = l.MaxQueryLookback.Set("0s")
	f.Var(&l.MaxQueryLookback, "querier.max-query-lookback", "Limit how far back in time series data and metadata can be queried, up until lookback duration ago. This limit is enforced in the query frontend, the querier and the ruler. If the requested time range is outside the allowed range, the request will not fail, but will be modified to only query data within the allowed time range. The default value of 0 does not set a limit.")
	f.IntVar(&l.MaxQueryParallelism, "querier.max-query-parallelism", 32, "Maximum number of queries that will be scheduled in parallel by the frontend.")
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (l *Limits) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// We want to set c to the defaults and then overwrite it with the input.
	// To make unmarshal fill the plain data struct rather than calling UnmarshalYAML
	// again, we have to hide it using a type indirection.  See prometheus/config.
	type plain Limits

	// During startup we wont have a default value so we don't want to overwrite them
	if defaultLimits != nil {
		b, err := yaml.Marshal(defaultLimits)
		if err != nil {
			return errors.Wrap(err, "cloning limits (marshaling)")
		}
		if err := yaml.Unmarshal(b, (*plain)(l)); err != nil {
			return errors.Wrap(err, "cloning limits (unmarshaling)")
		}
	}
	return unmarshal((*plain)(l))
}

// Validate validates that this limits config is valid.
func (l *Limits) Validate() error {
	return nil
}

// When we load YAML from disk, we want the various per-customer limits
// to default to any values specified on the command line, not default
// command line values.  This global contains those values.  I (Tom) cannot
// find a nicer way I'm afraid.
var defaultLimits *Limits

// SetDefaultLimitsForYAMLUnmarshalling sets global default limits, used when loading
// Limits from YAML files. This is used to ensure per-tenant limits are defaulted to
// those values.
func SetDefaultLimitsForYAMLUnmarshalling(defaults Limits) {
	defaultLimits = &defaults
}

type TenantLimits interface {
	// TenantLimits is a function that returns limits for given tenant, or
	// nil, if there are no tenant-specific limits.
	TenantLimits(userID string) *Limits
	// AllByUserID gets a mapping of all tenant IDs and limits for that user
	AllByUserID() map[string]*Limits
}

// Overrides periodically fetch a set of per-user overrides, and provides convenience
// functions for fetching the correct value.
type Overrides struct {
	defaultLimits *Limits
	tenantLimits  TenantLimits
}

// NewOverrides makes a new Overrides.
func NewOverrides(defaults Limits, tenantLimits TenantLimits) (*Overrides, error) {
	return &Overrides{
		tenantLimits:  tenantLimits,
		defaultLimits: &defaults,
	}, nil
}

func (o *Overrides) AllByUserID() map[string]*Limits {
	if o.tenantLimits != nil {
		return o.tenantLimits.AllByUserID()
	}
	return nil
}

// IngestionRateBytes returns the limit on ingester rate (MBs per second).
func (o *Overrides) IngestionRateBytes(userID string) float64 {
	return o.getOverridesForUser(userID).IngestionRateMB * bytesInMB
}

// IngestionBurstSizeBytes returns the burst size for ingestion rate.
func (o *Overrides) IngestionBurstSizeBytes(userID string) int {
	return int(o.getOverridesForUser(userID).IngestionBurstSizeMB * bytesInMB)
}

// MaxLabelNameLength returns maximum length a label name can be.
func (o *Overrides) MaxLabelNameLength(userID string) int {
	return o.getOverridesForUser(userID).MaxLabelNameLength
}

// MaxLabelValueLength returns maximum length a label value can be. This also is
// the maximum length of a metric name.
func (o *Overrides) MaxLabelValueLength(userID string) int {
	return o.getOverridesForUser(userID).MaxLabelValueLength
}

// MaxLabelNamesPerSeries returns maximum number of label/value pairs timeseries.
func (o *Overrides) MaxLabelNamesPerSeries(userID string) int {
	return o.getOverridesForUser(userID).MaxLabelNamesPerSeries
}

// MaxLocalStreamsPerUser returns the maximum number of streams a user is allowed to store
// in a single ingester.
func (o *Overrides) MaxLocalStreamsPerUser(userID string) int {
	return o.getOverridesForUser(userID).MaxLocalStreamsPerUser
}

// MaxGlobalStreamsPerUser returns the maximum number of streams a user is allowed to store
// across the cluster.
func (o *Overrides) MaxGlobalStreamsPerUser(userID string) int {
	return o.getOverridesForUser(userID).MaxGlobalStreamsPerUser
}

// MaxQueryLength returns the limit of the length (in time) of a query.
func (o *Overrides) MaxQueryLength(userID string) time.Duration {
	return time.Duration(o.getOverridesForUser(userID).MaxQueryLength)
}

// MaxQueryParallelism returns the limit to the number of sub-queries the
// frontend will process in parallel.
func (o *Overrides) MaxQueryParallelism(userID string) int {
	return o.getOverridesForUser(userID).MaxQueryParallelism
}

// MaxQueryLookback returns the max lookback period of queries.
func (o *Overrides) MaxQueryLookback(userID string) time.Duration {
	return time.Duration(o.getOverridesForUser(userID).MaxQueryLookback)
}

func (o *Overrides) DefaultLimits() *Limits {
	return o.defaultLimits
}

func (o *Overrides) getOverridesForUser(userID string) *Limits {
	if o.tenantLimits != nil {
		l := o.tenantLimits.TenantLimits(userID)
		if l != nil {
			return l
		}
	}
	return o.defaultLimits
}

// OverwriteMarshalingStringMap will overwrite the src map when unmarshaling
// as opposed to merging.
type OverwriteMarshalingStringMap struct {
	m map[string]string
}

func NewOverwriteMarshalingStringMap(m map[string]string) OverwriteMarshalingStringMap {
	return OverwriteMarshalingStringMap{m: m}
}

func (sm *OverwriteMarshalingStringMap) Map() map[string]string {
	return sm.m
}

// MarshalJSON explicitly uses the the type receiver and not pointer receiver
// or it won't be called
func (sm OverwriteMarshalingStringMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(sm.m)
}

func (sm *OverwriteMarshalingStringMap) UnmarshalJSON(val []byte) error {
	var def map[string]string
	if err := json.Unmarshal(val, &def); err != nil {
		return err
	}
	sm.m = def

	return nil
}

// MarshalYAML explicitly uses the the type receiver and not pointer receiver
// or it won't be called
func (sm OverwriteMarshalingStringMap) MarshalYAML() (interface{}, error) {
	return sm.m, nil
}

func (sm *OverwriteMarshalingStringMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var def map[string]string

	err := unmarshal(&def)
	if err != nil {
		return err
	}
	sm.m = def

	return nil
}
