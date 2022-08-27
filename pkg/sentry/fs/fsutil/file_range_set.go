// Copyright 2018 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fsutil

import (
	"fmt"
	"io"
	"math"

	"gvisor.dev/gvisor/pkg/context"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/safemem"
	"gvisor.dev/gvisor/pkg/sentry/memmap"
	"gvisor.dev/gvisor/pkg/sentry/pgalloc"
	"gvisor.dev/gvisor/pkg/sentry/usage"
)

// FileRangeSet maps offsets into a memmap.Mappable to offsets into a
// memmap.File. It is used to implement Mappables that store data in
// sparsely-allocated memory.
//
// type FileRangeSet <generated by go_generics>

// FileRangeSetFunctions implements segment.Functions for FileRangeSet.
type FileRangeSetFunctions struct{}

// MinKey implements segment.Functions.MinKey.
func (FileRangeSetFunctions) MinKey() uint64 {
	return 0
}

// MaxKey implements segment.Functions.MaxKey.
func (FileRangeSetFunctions) MaxKey() uint64 {
	return math.MaxUint64
}

// ClearValue implements segment.Functions.ClearValue.
func (FileRangeSetFunctions) ClearValue(_ *uint64) {
}

// Merge implements segment.Functions.Merge.
func (FileRangeSetFunctions) Merge(mr1 memmap.MappableRange, frstart1 uint64, _ memmap.MappableRange, frstart2 uint64) (uint64, bool) {
	if frstart1+mr1.Length() != frstart2 {
		return 0, false
	}
	return frstart1, true
}

// Split implements segment.Functions.Split.
func (FileRangeSetFunctions) Split(mr memmap.MappableRange, frstart uint64, split uint64) (uint64, uint64) {
	return frstart, frstart + (split - mr.Start)
}

// FileRange returns the FileRange mapped by seg.
func (seg FileRangeIterator) FileRange() memmap.FileRange {
	return seg.FileRangeOf(seg.Range())
}

// FileRangeOf returns the FileRange mapped by mr.
//
// Preconditions:
//   - seg.Range().IsSupersetOf(mr).
//   - mr.Length() != 0.
func (seg FileRangeIterator) FileRangeOf(mr memmap.MappableRange) memmap.FileRange {
	frstart := seg.Value() + (mr.Start - seg.Start())
	return memmap.FileRange{frstart, frstart + mr.Length()}
}

// PagesToFill returns the number of pages that that Fill() will allocate
// for the given required and optional parameters.
func (frs *FileRangeSet) PagesToFill(required, optional memmap.MappableRange) uint64 {
	var numPages uint64
	gap := frs.LowerBoundGap(required.Start)
	for gap.Ok() && gap.Start() < required.End {
		gr := gap.Range().Intersect(optional)
		numPages += gr.Length() / hostarch.PageSize
		gap = gap.NextGap()
	}
	return numPages
}

// Fill attempts to ensure that all memmap.Mappable offsets in required are
// mapped to a memmap.File offset, by allocating from mf with the given
// memory usage kind and invoking readAt to store data into memory. (If readAt
// returns a successful partial read, Fill will call it repeatedly until all
// bytes have been read.) EOF is handled consistently with the requirements of
// mmap(2): bytes after EOF on the same page are zeroed; pages after EOF are
// invalid. fileSize is an upper bound on the file's size; bytes after fileSize
// will be zeroed without calling readAt.
//
// Fill may read offsets outside of required, but will never read offsets
// outside of optional. It returns a non-nil error if any error occurs, even
// if the error only affects offsets in optional, but not in required.
//
// Fill returns the number of pages that were allocated.
//
// Preconditions:
//   - required.Length() > 0.
//   - optional.IsSupersetOf(required).
//   - required and optional must be page-aligned.
func (frs *FileRangeSet) Fill(ctx context.Context, required, optional memmap.MappableRange, fileSize uint64, mf *pgalloc.MemoryFile, kind usage.MemoryKind, readAt func(ctx context.Context, dsts safemem.BlockSeq, offset uint64) (uint64, error)) (uint64, error) {
	gap := frs.LowerBoundGap(required.Start)
	var pagesAlloced uint64
	for gap.Ok() && gap.Start() < required.End {
		if gap.Range().Length() == 0 {
			gap = gap.NextGap()
			continue
		}
		gr := gap.Range().Intersect(optional)

		// Read data into the gap.
		fr, err := mf.AllocateAndFill(gr.Length(), kind, safemem.ReaderFunc(func(dsts safemem.BlockSeq) (uint64, error) {
			var done uint64
			for !dsts.IsEmpty() {
				n, err := func() (uint64, error) {
					off := gr.Start + done
					if off >= fileSize {
						return 0, io.EOF
					}
					if off+dsts.NumBytes() > fileSize {
						rd := fileSize - off
						n, err := readAt(ctx, dsts.TakeFirst64(rd), off)
						if n == rd && err == nil {
							return n, io.EOF
						}
						return n, err
					}
					return readAt(ctx, dsts, off)
				}()
				done += n
				dsts = dsts.DropFirst64(n)
				if err != nil {
					if err == io.EOF {
						// MemoryFile.AllocateAndFill truncates down to a page
						// boundary, but FileRangeSet.Fill is supposed to
						// zero-fill to the end of the page in this case.
						donepgaddr, ok := hostarch.Addr(done).RoundUp()
						if donepg := uint64(donepgaddr); ok && donepg != done {
							dsts.DropFirst64(donepg - done)
							done = donepg
							if dsts.IsEmpty() {
								return done, nil
							}
						}
					}
					return done, err
				}
			}
			return done, nil
		}))

		// Store anything we managed to read into the cache.
		if done := fr.Length(); done != 0 {
			gr.End = gr.Start + done
			pagesAlloced += gr.Length() / hostarch.PageSize
			gap = frs.Insert(gap, gr, fr.Start).NextGap()
		}

		if err != nil {
			return pagesAlloced, err
		}
	}
	return pagesAlloced, nil
}

// Drop removes segments for memmap.Mappable offsets in mr, freeing the
// corresponding memmap.FileRanges.
//
// Preconditions: mr must be page-aligned.
func (frs *FileRangeSet) Drop(mr memmap.MappableRange, mf *pgalloc.MemoryFile) {
	seg := frs.LowerBoundSegment(mr.Start)
	for seg.Ok() && seg.Start() < mr.End {
		seg = frs.Isolate(seg, mr)
		mf.DecRef(seg.FileRange())
		seg = frs.Remove(seg).NextSegment()
	}
}

// DropAll removes all segments in mr, freeing the corresponding
// memmap.FileRanges. It returns the number of pages freed.
func (frs *FileRangeSet) DropAll(mf *pgalloc.MemoryFile) uint64 {
	var pagesFreed uint64
	for seg := frs.FirstSegment(); seg.Ok(); seg = seg.NextSegment() {
		mf.DecRef(seg.FileRange())
		pagesFreed += seg.Range().Length() / hostarch.PageSize
	}
	frs.RemoveAll()
	return pagesFreed
}

// Truncate updates frs to reflect Mappable truncation to the given length:
// bytes after the new EOF on the same page are zeroed, and pages after the new
// EOF are freed. It returns the number of pages freed.
func (frs *FileRangeSet) Truncate(end uint64, mf *pgalloc.MemoryFile) uint64 {
	var pagesFreed uint64
	pgendaddr, ok := hostarch.Addr(end).RoundUp()
	if ok {
		pgend := uint64(pgendaddr)

		// Free truncated pages.
		frs.SplitAt(pgend)
		seg := frs.LowerBoundSegment(pgend)
		for seg.Ok() {
			mf.DecRef(seg.FileRange())
			pagesFreed += seg.Range().Length() / hostarch.PageSize
			seg = frs.Remove(seg).NextSegment()
		}

		if end == pgend {
			return pagesFreed
		}
	}

	// Here we know end < end.RoundUp(). If the new EOF lands in the
	// middle of a page that we have, zero out its contents beyond the new
	// length.
	seg := frs.FindSegment(end)
	if seg.Ok() {
		fr := seg.FileRange()
		fr.Start += end - seg.Start()
		ims, err := mf.MapInternal(fr, hostarch.Write)
		if err != nil {
			// There's no good recourse from here. This means
			// that we can't keep cached memory consistent with
			// the new end of file. The caller may have already
			// updated the file size on their backing file system.
			//
			// We don't want to risk blindly continuing onward,
			// so in the extremely rare cases this does happen,
			// we abandon ship.
			panic(fmt.Sprintf("Failed to map %v: %v", fr, err))
		}
		if _, err := safemem.ZeroSeq(ims); err != nil {
			panic(fmt.Sprintf("Zeroing %v failed: %v", fr, err))
		}
	}
	return pagesFreed
}
