package experiment

type NestedStacktrace struct {
	ID         uint64
	LocationID uint64
	Children   []NestedStacktrace
}

type NestedStacktraces struct {
	stacktraces NestedStacktrace
	m           map[string]*NestedStacktrace
}

func newNestedStacktraces(l int) NestedStacktraces {
	return NestedStacktraces{
		stacktraces: NestedStacktrace{},
		m:           make(map[string]*NestedStacktrace),
	}
}

/*

// add a stacktrace
func (s *NestedStacktraces) add(id uint64, v []uint64) uint64 {
	if len(v) == 0 {
		return 0
	}

	k := key(v)
	pos, ok := s.m[k]
	if ok {
		return pos
	}

	parent := uint64(0)
	if len(v) > 1 {
		parent = s.add(v[1:])
	}

	s.stacktraces = append(s.stacktraces, SelfRefStacktrace{
		Parent:     parent,
		LocationID: v[0],
	})
	pos = uint64(len(s.stacktraces) - 1)
	s.m[k] = pos
	return pos
}

func TestSelfRefStacktraces(t *testing.T) {
	beforePath := "/home/christian/parquet-notebook/stacktraces.parquet"
	afterPath := "/home/christian/parquet-notebook/stacktraces2.parquet"
	f, err := os.Open(beforePath)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	reader := parquet.NewGenericReader[schemav1.Stacktrace](f)

	t.Logf("original has %d rows", reader.NumRows())

	s := newStacktraces(int(reader.NumRows()))

	var slice = make([]schemav1.Stacktrace, 100_000)
	for {
		n, err := reader.Read(slice)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		t.Logf("read %d rows", n)

		for _, c := range slice[:n] {
			s.add(c.LocationIDs)
		}
	}

	// read all rows and insert into new structure

	require.NoError(t, parquet.WriteFile[SelfRefStacktrace](afterPath, s.stacktraces))

}
*/
