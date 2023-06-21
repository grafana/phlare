package symtab

import (
	"github.com/grafana/phlare/ebpf/util"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testCacheOptions = GCacheOptions{32, 3}
)

func TestElfCacheStrippedEmpty(t *testing.T) {
	logger := util.TestLogger(t)
	elfCache, _ := NewElfCache(testCacheOptions, testCacheOptions)
	fs := "." // make it unable to find debug file by buildID
	stripped := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, fs, "elf/testdata/elfs/elf.stripped",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	syms := []struct {
		name string
		pc   uint64
	}{
		{"iter", 0x1149},
		{"main", 0x115e},
	}
	for _, sym := range syms {
		res := stripped.Resolve(sym.pc)
		require.Error(t, stripped.err)
		require.Equal(t, "", res)
	}
}

func TestElfCacheBuildID(t *testing.T) {
	elfCache, _ := NewElfCache(testCacheOptions, testCacheOptions)
	logger := util.TestLogger(t)
	debug := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	stripped := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf.stripped",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	syms := []struct {
		name string
		pc   uint64
	}{
		{"iter", 0x1149},
		{"main", 0x115e},
	}
	for _, sym := range syms {
		res := debug.Resolve(sym.pc)
		require.NoError(t, debug.err)
		require.Equal(t, sym.name, res)
		res = stripped.Resolve(sym.pc)
		require.NoError(t, stripped.err)
		require.Equal(t, sym.name, res)
	}
	require.Equal(t, 1, elfCache.BuildIDCache.lruCache.Len())
	require.Equal(t, 1, elfCache.SameFileCache.lruCache.Len())
}

func TestElfCacheStat(t *testing.T) {
	elfCache, _ := NewElfCache(testCacheOptions, testCacheOptions)
	logger := util.TestLogger(t)
	f1 := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf.nobuildid",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	f2 := NewElfTable(logger, &ProcMap{StartAddr: 0x1000, Offset: 0x1000}, ".", "elf/testdata/elfs/elf.nobuildid",
		ElfTableOptions{
			ElfCache: elfCache,
		})

	syms := []struct {
		name string
		pc   uint64
	}{
		{"iter", 0x1149},
		{"main", 0x115e},
	}
	for _, sym := range syms {
		res := f1.Resolve(sym.pc)
		require.NoError(t, f1.err)
		require.Equal(t, sym.name, res)
		res = f2.Resolve(sym.pc)
		require.NoError(t, f2.err)
		require.Equal(t, sym.name, res)
	}
	require.Equal(t, 0, elfCache.BuildIDCache.lruCache.Len())
	require.Equal(t, 1, elfCache.SameFileCache.lruCache.Len())
}
