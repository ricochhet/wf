package bmap_test

import (
	"sort"
	"sync"
	"testing"

	"github.com/ricochhet/pkg/bmap"
)

func TestSet(t *testing.T) {
	t.Parallel()

	bmap := bmap.Bmap[int, int]{}

	for i := range 10 {
		go func() {
			for j := range 10 {
				want := j
				bmap.Set(i, want)

				got, _ := bmap.Get(i)
				if want != got {
					t.Errorf("got %d, wanted %d", got, want)
				}

				got, _ = bmap.Get(i)
				if want != got {
					t.Errorf("got %d, wanted %d", got, want)
				}
			}
		}()
	}
}

func TestAsync(t *testing.T) {
	t.Parallel()

	bmap := bmap.Bmap[int, int]{}

	for i := range 100000 {
		want := i * 13
		bmap.Set(i, want)

		got, _ := bmap.Get(i)
		if want != got {
			t.Errorf("got %d, wanted %d", got, want)
		}
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	bmap := bmap.Bmap[int, int]{}
	want := 847392
	bmap.Set(1001, want)

	for i := range 1000 {
		go bmap.Set(i, i)
		go bmap.Set(i*2, i*3)
		go bmap.Delete(i)
		go bmap.Delete(i * 2)
	}

	got, _ := bmap.Get(1001)
	if want != got {
		t.Errorf("got %d, wanted %d", got, want)
	}
}

func TestSwap(t *testing.T) {
	t.Parallel()

	bmap := bmap.Bmap[int, int]{}
	want := 107834
	bmap.Set(-1, want)

	for i := range 1000 {
		bmap.Set(i, i)

		err := bmap.Swap(i-1, i)
		if err != nil {
			t.Errorf("swap err: %v", err)
		}
	}

	got, _ := bmap.Get(999)
	if want != got {
		t.Errorf("got %d, wanted %d", got, want)
	}
}

func BenchmarkSetBmap(b *testing.B) {
	bmap := bmap.Bmap[int, int]{}
	for i := range b.N {
		bmap.Set(i, i)
	}
}

func BenchmarkInitSyncmap(b *testing.B) {
	smap := sync.Map{}
	for i := range b.N {
		smap.Store(i, i)
	}
}

func BenchmarkGetBmap(b *testing.B) {
	bmap := bmap.Bmap[int, int]{}
	bmap.Set(0, 0)

	for b.Loop() {
		bmap.Get(0)
	}
}

func BenchmarkGetSyncmap(b *testing.B) {
	smap := sync.Map{}
	smap.Store(0, 0)

	for b.Loop() {
		smap.Load(0)
	}
}

func BenchmarkMultiBmap(b *testing.B) {
	bmaps := bmap.Bmap[int, *bmap.Bmap[int, int]]{}

	for i := range b.N {
		bmap := bmap.Bmap[int, int]{}
		bmap.Set(i, i)
		bmaps.Set(i, &bmap)
	}

	for i := range b.N {
		bmap, _ := bmaps.Get(i)
		bmap.Get(i)
	}

	for i := range b.N {
		bmap, _ := bmaps.Get(i)
		bmap.Delete(i)
	}
}

func BenchmarkMultiSyncmap(b *testing.B) {
	smaps := sync.Map{}

	for i := range b.N {
		smap := sync.Map{}
		smap.Store(i, i)
		smaps.Store(i, &smap)
	}

	for i := range b.N {
		smap, _ := smaps.Load(i)

		m, ok := smap.(*sync.Map)
		if !ok {
			b.Fatalf("expected *sync.Map, got %T", smap)
		}

		m.Load(i)
	}

	for i := range b.N {
		smap, _ := smaps.Load(i)

		m, ok := smap.(*sync.Map)
		if !ok {
			b.Fatalf("expected *sync.Map, got %T", smap)
		}

		m.Delete(i)
	}
}

func BenchmarkSortMultiBmap(b *testing.B) {
	bmap1 := bmap.Bmap[int, int]{}
	bmap2 := bmap.Bmap[int, int]{}
	bmap3 := bmap.Bmap[int, int]{}

	for i := range b.N {
		want := i
		bmap1.Set(i, want)
		bmap2.Set(i, want)
		bmap3.Set(i, want)

		got, _ := bmap1.Get(i)
		if want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		got, _ = bmap2.Get(i)
		if want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		got, _ = bmap3.Get(i)
		if want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		bmap1.Sort(func(i, j int) bool {
			return i > j
		})
		bmap2.Sort(func(i, j int) bool {
			return i > j
		})
		bmap3.Sort(func(i, j int) bool {
			return i > j
		})

		got, _ = bmap1.Get(i)
		if want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		got, _ = bmap2.Get(i)
		if want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		got, _ = bmap3.Get(i)
		if want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}
	}

	for i := range 1000 {
		bmap1.Delete(i)
		bmap2.Delete(i)
		bmap3.Delete(i)
	}
}

func BenchmarkSortMultiSyncmap(b *testing.B) {
	smap1 := sync.Map{}
	smap2 := sync.Map{}
	smap3 := sync.Map{}

	for i := range b.N {
		want := i
		smap1.Store(i, want)
		smap2.Store(i, want)
		smap3.Store(i, want)

		got, _ := smap1.Load(i)
		if got != nil && want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		got, _ = smap2.Load(i)
		if got != nil && want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		got, _ = smap3.Load(i)
		if got != nil && want != got {
			b.Errorf("got %d, wanted %d", got, want)
		}

		// Manually sort the sync maps.
		keys := []int{}

		smap1.Range(func(k, _ any) bool {
			if ki, ok := k.(int); ok {
				keys = append(keys, ki)
			} else {
				b.Fatalf("expected int, got %T", k)
			}

			return true
		})
		sort.Ints(keys)

		for _, k := range keys {
			_, _ = smap1.Load(k)
		}

		keys = []int{}

		smap2.Range(func(k, _ any) bool {
			if ki, ok := k.(int); ok {
				keys = append(keys, ki)
			} else {
				b.Fatalf("expected int, got %T", k)
			}

			return true
		})
		sort.Ints(keys)

		for _, k := range keys {
			_, _ = smap2.Load(k)
		}

		keys = []int{}

		smap3.Range(func(k, _ any) bool {
			if ki, ok := k.(int); ok {
				keys = append(keys, ki)
			} else {
				b.Fatalf("expected int, got %T", k)
			}

			return true
		})
		sort.Ints(keys)

		for _, k := range keys {
			_, _ = smap3.Load(k)
		}
	}

	for i := range 1000 {
		smap1.Delete(i)
		smap2.Delete(i)
		smap3.Delete(i)
	}
}
