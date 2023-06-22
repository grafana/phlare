package v1

import (
	"fmt"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/segmentio/parquet-go"
	"github.com/stretchr/testify/require"
)

func TestInMemoryProfilesRowReader(t *testing.T) {
	r := NewProfilesRowReader(
		generateProfiles(10),
	)

	batch := make([]parquet.Row, 3)
	count := 0
	for {
		n, err := r.ReadRows(batch)
		if err != nil && err != io.EOF {
			t.Fatal(err)
		}
		count += n
		if n == 0 || err == io.EOF {
			break
		}
	}
	require.Equal(t, 10, count)
}

const samplesPerProfile = 3

func TestRoundtripProfile(t *testing.T) {
	profiles := generateProfiles(1000)
	iprofiles := generateMemoryProfiles(1000)
	actual, err := readAll(NewInMemoryProfilesRowReader(iprofiles))
	require.NoError(t, err)
	expected, err := readAll(NewProfilesRowReader(profiles))
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func BenchmarkRowReader(b *testing.B) {
	profiles := generateProfiles(1000)
	iprofiles := generateMemoryProfiles(1000)
	b.Run("in-memory", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := readAll(NewInMemoryProfilesRowReader(iprofiles))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("schema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := readAll(NewProfilesRowReader(profiles))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func readAll(r parquet.RowReader) ([]parquet.Row, error) {
	var rows []parquet.Row
	batch := make([]parquet.Row, 1000)
	for {
		n, err := r.ReadRows(batch)
		if err != nil && err != io.EOF {
			return rows, err
		}
		if n != 0 {
			rows = append(rows, batch[:n]...)
		}
		if n == 0 || err == io.EOF {
			break
		}
	}
	return rows, nil
}

func generateMemoryProfiles(n int) []InMemoryProfile {
	profiles := make([]InMemoryProfile, n)
	for i := 0; i < n; i++ {
		stacktraceID := make([]uint32, samplesPerProfile)
		value := make([]uint64, samplesPerProfile)
		for j := 0; j < samplesPerProfile; j++ {
			stacktraceID[j] = uint32(j)
			value[j] = uint64(j)
		}
		profiles[i] = InMemoryProfile{
			ID:                uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i)),
			SeriesIndex:       uint32(i),
			DropFrames:        1,
			KeepFrames:        3,
			TimeNanos:         int64(i),
			Period:            100000,
			DurationNanos:     1000000000,
			Comments:          []int64{1, 2, 3},
			DefaultSampleType: 2,
			Samples: Samples{
				StacktraceIDs: stacktraceID,
				Values:        value,
			},
		}
	}
	return profiles
}

func generateProfiles(n int) []*Profile {
	profiles := make([]*Profile, n)
	for i := 0; i < n; i++ {
		profiles[i] = &Profile{
			ID:                uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i)),
			SeriesIndex:       uint32(i),
			DropFrames:        1,
			KeepFrames:        3,
			TimeNanos:         int64(i),
			Period:            100000,
			DurationNanos:     1000000000,
			Comments:          []int64{1, 2, 3},
			DefaultSampleType: 2,
			Samples:           generateSamples(samplesPerProfile),
		}
	}

	return profiles
}

func generateSamples(n int) []*Sample {
	samples := make([]*Sample, n)
	for i := 0; i < n; i++ {
		samples[i] = &Sample{
			StacktraceID: uint64(i),
			Value:        int64(i),
		}
	}
	return samples
}
