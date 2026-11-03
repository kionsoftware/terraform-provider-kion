package paging_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/paging"
)

// server serves a collection of `total` records in pages of `pageSize`,
// optionally reporting a total larger than it will actually serve.
type server struct {
	total     int  // what it reports
	serve     int  // what it actually has; 0 means "same as total"
	pageSize  int  // records per page
	notPaged  bool // report Total < 0, the non-envelope case
	pagesSeen []int
}

func (s *server) fetch(_ context.Context, page int) (paging.Page[int], error) {
	s.pagesSeen = append(s.pagesSeen, page)
	if s.notPaged {
		return paging.Page[int]{Items: []int{1, 2, 3}, Total: -1}, nil
	}
	have := s.serve
	if have == 0 {
		have = s.total
	}
	start := (page - 1) * s.pageSize
	items := []int{}
	for i := start; i < start+s.pageSize && i < have; i++ {
		items = append(items, i+1)
	}
	return paging.Page[int]{Items: items, Total: s.total}, nil
}

// A collection larger than one page must come back whole.
func TestAll_walksEveryPage(t *testing.T) {
	t.Parallel()

	s := &server{total: 25, pageSize: 10}
	got, err := paging.All(t.Context(), "test", s.fetch)

	require.NoError(t, err)
	require.Len(t, got, 25)
	require.Equal(t, 1, got[0])
	require.Equal(t, 25, got[24])
	require.Equal(t, []int{1, 2, 3}, s.pagesSeen)
}

func TestAll_singlePage(t *testing.T) {
	t.Parallel()

	s := &server{total: 4, pageSize: 10}
	got, err := paging.All(t.Context(), "test", s.fetch)

	require.NoError(t, err)
	require.Len(t, got, 4)
	require.Equal(t, []int{1}, s.pagesSeen, "one page suffices, so only one request")
}

func TestAll_empty(t *testing.T) {
	t.Parallel()

	s := &server{total: 0, pageSize: 10}
	got, err := paging.All(t.Context(), "test", s.fetch)

	require.NoError(t, err)
	require.Empty(t, got)
	require.Equal(t, []int{1}, s.pagesSeen)
}

// A negative Total means the response was not a paginated envelope, so the
// first page is the whole answer and nothing more is requested.
func TestAll_notPaginated(t *testing.T) {
	t.Parallel()

	s := &server{notPaged: true}
	got, err := paging.All(t.Context(), "test", s.fetch)

	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, []int{1}, s.pagesSeen)
}

// An empty page ends the walk, so a total the server never fulfills does not
// run it to MaxPages.
func TestAll_stopsOnEmptyPageDespiteWrongTotal(t *testing.T) {
	t.Parallel()

	s := &server{total: 10_000, serve: 5, pageSize: 10}
	got, err := paging.All(t.Context(), "test", s.fetch)

	require.NoError(t, err)
	require.Len(t, got, 5)
	require.Equal(t, []int{1, 2}, s.pagesSeen, "stops at the first empty page")
}

// A full page that never advances past the total is bounded by the cap.
func TestAll_capsRunawayWalk(t *testing.T) {
	t.Parallel()

	calls := 0
	_, err := paging.All(t.Context(), "listing widgets", func(_ context.Context, _ int) (paging.Page[int], error) {
		calls++
		// Always one item, always claims more remain.
		return paging.Page[int]{Items: []int{1}, Total: 1 << 30}, nil
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "listing widgets")
	require.Contains(t, err.Error(), "paging exceeded max pages")
	require.LessOrEqual(t, calls, paging.MaxPages+1)
}

// A first-page failure yields no records: there is nothing partial to report.
func TestAll_firstPageErrorReturnsNothing(t *testing.T) {
	t.Parallel()

	want := errors.New("boom")
	got, err := paging.All(t.Context(), "test", func(_ context.Context, _ int) (paging.Page[int], error) {
		return paging.Page[int]{}, want
	})

	require.ErrorIs(t, err, want)
	require.Nil(t, got)
}

// A later failure returns what was gathered alongside the error.
func TestAll_laterPageErrorReturnsPartial(t *testing.T) {
	t.Parallel()

	want := errors.New("page 2 exploded")
	got, err := paging.All(t.Context(), "test", func(_ context.Context, page int) (paging.Page[int], error) {
		if page == 1 {
			return paging.Page[int]{Items: []int{1, 2}, Total: 10}, nil
		}
		return paging.Page[int]{}, want
	})

	require.ErrorIs(t, err, want)
	require.Equal(t, []int{1, 2}, got, "the partial collection is returned with the error")
}
