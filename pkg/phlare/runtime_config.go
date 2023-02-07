package phlare

import (
	"github.com/cortexproject/cortex/pkg/util/validation"
)


type runtimeConfigValues struct {
	TenantLimits map[string]*validation.Limits `yaml:"overrides"`
}
