// SPDX-License-Identifier: AGPL-3.0-only

package validation

import (
	"net/http"

	"github.com/grafana/dskit/tenant"

	"github.com/grafana/phlare/pkg/util"
)

type UserLimitsResponse struct {
	// Write path limits
	IngestionRate          float64 `json:"ingestion_rate"`
	IngestionBurstSize     int     `json:"ingestion_burst_size"`
	MaxGlobalSeriesPerUser int     `json:"max_global_series_per_user"`

	// todo
}

// UserLimitsHandler handles user limits.
func UserLimitsHandler(defaultLimits Limits, tenantLimits TenantLimits) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := tenant.TenantID(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userLimits := tenantLimits.TenantLimits(userID)
		if userLimits == nil {
			userLimits = &defaultLimits
		}

		limits := UserLimitsResponse{
			// Write path limits
			IngestionRate:          userLimits.IngestionRateMB,
			IngestionBurstSize:     int(userLimits.IngestionBurstSizeMB),
			MaxGlobalSeriesPerUser: userLimits.MaxGlobalSeriesPerUser,
		}

		util.WriteJSONResponse(w, limits)
	}
}
