// Protocol Buffers - Google's data interchange format
// Copyright 2008 Google Inc.  All rights reserved.
// Copyright 2022 Teppei Fukuda.  All rights reserved.
// https://developers.google.com/protocol-buffers/

package fieldmaskpb_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/runtime-radar/go-plugin/types/known/fieldmaskpb"
)

func TestAppend(t *testing.T) {
	t.Run("appends in order", func(t *testing.T) {
		m := &fieldmaskpb.FieldMask{}
		m.Append("a", "b.c")
		m.Append("d")
		assert.Equal(t, []string{"a", "b.c", "d"}, m.Paths)
	})

	t.Run("no-op when given nothing", func(t *testing.T) {
		m := &fieldmaskpb.FieldMask{Paths: []string{"x"}}
		m.Append()
		assert.Equal(t, []string{"x"}, m.Paths)
	})
}

func TestUnion(t *testing.T) {
	tests := []struct {
		name string
		in   []*fieldmaskpb.FieldMask
		want []string
	}{
		{
			name: "empty input",
			in:   nil,
			want: nil,
		},
		{
			name: "deduplicates and sorts",
			in: []*fieldmaskpb.FieldMask{
				{Paths: []string{"a", "b"}},
				{Paths: []string{"b", "c"}},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "tolerates nil masks",
			in: []*fieldmaskpb.FieldMask{
				nil,
				{Paths: []string{"x"}},
				nil,
			},
			want: []string{"x"},
		},
		{
			name: "parent absorbs child",
			in: []*fieldmaskpb.FieldMask{
				{Paths: []string{"foo"}},
				{Paths: []string{"foo.bar"}},
			},
			want: []string{"foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldmaskpb.Union(tt.in...)
			assert.Equal(t, tt.want, got.Paths)
		})
	}
}

func TestIntersect(t *testing.T) {
	tests := []struct {
		name string
		in   []*fieldmaskpb.FieldMask
		want []string
	}{
		{
			name: "no input",
			in:   nil,
			want: nil,
		},
		{
			name: "single mask returns its normalized self",
			in: []*fieldmaskpb.FieldMask{
				{Paths: []string{"b", "a", "a"}},
			},
			want: []string{"a", "b"},
		},
		{
			name: "common paths only",
			in: []*fieldmaskpb.FieldMask{
				{Paths: []string{"a", "b", "c"}},
				{Paths: []string{"b", "c", "d"}},
			},
			want: []string{"b", "c"},
		},
		{
			name: "empty when one mask is empty",
			in: []*fieldmaskpb.FieldMask{
				{Paths: []string{"a"}},
				{Paths: []string{}},
			},
			want: nil,
		},
		{
			name: "child intersects parent — most-specific wins",
			in: []*fieldmaskpb.FieldMask{
				{Paths: []string{"foo.bar", "user"}},
				{Paths: []string{"foo", "user.email"}},
			},
			want: []string{"foo.bar", "user.email"},
		},
		{
			name: "disjoint masks",
			in: []*fieldmaskpb.FieldMask{
				{Paths: []string{"a"}},
				{Paths: []string{"b"}},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldmaskpb.Intersect(tt.in...)
			assert.Equal(t, tt.want, got.Paths)
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "empty", in: nil, want: nil},
		{name: "sorts", in: []string{"b", "a"}, want: []string{"a", "b"}},
		{name: "dedupes", in: []string{"a", "a", "b"}, want: []string{"a", "b"}},
		{
			name: "parent covers child",
			in:   []string{"foo.bar", "foo"},
			want: []string{"foo"},
		},
		{
			name: "parent only covers descendants at boundary",
			in:   []string{"foo", "foobar"},
			want: []string{"foo", "foobar"},
		},
		{
			name: "deeply nested redundancy",
			in:   []string{"a.b.c.d", "a.b", "a.b.c"},
			want: []string{"a.b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &fieldmaskpb.FieldMask{Paths: tt.in}
			m.Normalize()
			assert.Equal(t, tt.want, m.Paths)
		})
	}

	t.Run("nil receiver is safe", func(t *testing.T) {
		var m *fieldmaskpb.FieldMask
		m.Normalize()
	})
}

func TestIsValid(t *testing.T) {
	t.Run("valid paths", func(t *testing.T) {
		m := &fieldmaskpb.FieldMask{Paths: []string{
			"a",
			"foo",
			"foo.bar",
			"a.b.c",
			"name_with_underscore",
			"x1.y2",
		}}
		assert.True(t, m.IsValid())
	})

	cases := []struct {
		name string
		path string
	}{
		{"empty path", ""},
		{"capital letter", "Foo"},
		{"capital in segment", "foo.Bar"},
		{"double dot", "foo..bar"},
		{"trailing dot", "foo."},
		{"leading dot", ".foo"},
		{"leading digit", "9foo"},
		{"hyphen", "foo-bar"},
		{"space", "foo bar"},
		{"underscore-only segment is invalid", "_foo"},
	}
	for _, tt := range cases {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			m := &fieldmaskpb.FieldMask{Paths: []string{tt.path}}
			assert.False(t, m.IsValid(), "path: %q", tt.path)
		})
	}

	t.Run("nil receiver is invalid", func(t *testing.T) {
		var m *fieldmaskpb.FieldMask
		assert.False(t, m.IsValid())
	})

	t.Run("empty mask is valid", func(t *testing.T) {
		m := &fieldmaskpb.FieldMask{}
		assert.True(t, m.IsValid())
	})
}

func TestGetPaths(t *testing.T) {
	t.Run("nil receiver returns nil", func(t *testing.T) {
		var m *fieldmaskpb.FieldMask
		assert.Nil(t, m.GetPaths())
	})

	t.Run("returns underlying slice", func(t *testing.T) {
		m := &fieldmaskpb.FieldMask{Paths: []string{"a", "b"}}
		assert.Equal(t, []string{"a", "b"}, m.GetPaths())
	})
}
