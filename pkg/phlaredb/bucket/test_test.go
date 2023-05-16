package bucket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_blockPrefixesFromTo(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2023-02-17T15:04:05Z")
	require.NoError(t, err)

	for _, tc := range []struct {
		from         time.Time
		to           time.Time
		orderOfSplit uint8

		want    []string
		wantErr bool
	}{
		{
			from:         now.Add(-time.Hour * 24 * 7),
			to:           now,
			orderOfSplit: 0,
			want:         []string{"0"},
		},
		{
			from:         now.Add(-time.Hour * 24 * 7),
			to:           now,
			orderOfSplit: 3,
			want:         []string{"01GR", "01GS"},
		},
		{
			from:         now.Add(-time.Hour * 24),
			to:           now,
			orderOfSplit: 4,
			want:         []string{"01GSD", "01GSE", "01GSF"},
		},
		{
			from:         now.Add(-time.Hour * 24),
			to:           now,
			orderOfSplit: 10,
			wantErr:      true,
		},
	} {

		result, err := blockPrefixesFromTo(tc.from, tc.to, tc.orderOfSplit)
		if tc.wantErr {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
		require.Equal(t, tc.want, result)
	}

}
