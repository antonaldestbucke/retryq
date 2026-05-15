package tagging_test

import (
	"testing"

	"github.com/example/retryq/job"
	"github.com/example/retryq/tagging"
)

func newJob(t *testing.T) *job.Job {
	t.Helper()
	j, err := job.New("test", []byte(`{}`))
	if err != nil {
		t.Fatalf("job.New: %v", err)
	}
	return j
}

func TestSet_And_Get(t *testing.T) {
	s := tagging.New()
	j := newJob(t)
	s.Set(j, "alpha", "beta")
	tags := s.Get(j)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
}

func TestAdd_AppendsTags(t *testing.T) {
	s := tagging.New()
	j := newJob(t)
	s.Set(j, "alpha")
	s.Add(j, "beta", "gamma")
	if !s.Has(j, "alpha", "beta", "gamma") {
		t.Fatal("expected all three tags to be present")
	}
}

func TestHas_ReturnsFalseForMissingTag(t *testing.T) {
	s := tagging.New()
	j := newJob(t)
	s.Set(j, "alpha")
	if s.Has(j, "alpha", "missing") {
		t.Fatal("expected Has to return false when a tag is absent")
	}
}

func TestRemove_DeletesSpecificTags(t *testing.T) {
	s := tagging.New()
	j := newJob(t)
	s.Set(j, "alpha", "beta", "gamma")
	s.Remove(j, "beta")
	if s.Has(j, "beta") {
		t.Fatal("removed tag should no longer be present")
	}
	if !s.Has(j, "alpha", "gamma") {
		t.Fatal("remaining tags should still be present")
	}
}

func TestDelete_RemovesAllTags(t *testing.T) {
	s := tagging.New()
	j := newJob(t)
	s.Set(j, "alpha", "beta")
	s.Delete(j)
	if len(s.Get(j)) != 0 {
		t.Fatal("expected no tags after Delete")
	}
}

func TestGet_ReturnsEmptyForUnknownJob(t *testing.T) {
	s := tagging.New()
	j := newJob(t)
	if tags := s.Get(j); len(tags) != 0 {
		t.Fatalf("expected empty slice, got %v", tags)
	}
}

func TestSet_ReplacesExistingTags(t *testing.T) {
	s := tagging.New()
	j := newJob(t)
	s.Set(j, "old")
	s.Set(j, "new")
	if s.Has(j, "old") {
		t.Fatal("Set should have replaced old tags")
	}
	if !s.Has(j, "new") {
		t.Fatal("Set should have applied new tag")
	}
}
