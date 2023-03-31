/*
Copyright 2021 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package policy

import (
	"testing"

	"github.com/fluxcd/image-reflector-controller/internal/database"
	. "github.com/onsi/gomega"
)

func TestRegexFilter(t *testing.T) {
	cases := []struct {
		label    string
		tags     []database.Tag
		pattern  string
		extract  string
		expected []database.Tag
		origTags map[string]int
	}{
		{
			label:    "none",
			tags:     []database.Tag{{Name: "a", Digest: "aa"}},
			expected: []database.Tag{{Name: "a", Digest: "aa"}},
			origTags: map[string]int{
				"a": 0,
			},
		},
		{
			label: "valid pattern",
			tags: []database.Tag{
				{Name: "ver1", Digest: "1rev"},
				{Name: "ver2", Digest: "2rev"},
				{Name: "ver3", Digest: "3rev"},
				{Name: "rel1", Digest: "1ler"},
			},
			pattern: "^ver",
			expected: []database.Tag{
				{Name: "ver1", Digest: "1rev"},
				{Name: "ver2", Digest: "2rev"},
				{Name: "ver3", Digest: "3rev"},
			},
			origTags: map[string]int{
				"ver1": 0,
			},
		},
		{
			label: "valid pattern with capture group",
			tags: []database.Tag{
				{Name: "ver1", Digest: "foo"},
				{Name: "ver2", Digest: "bar"},
				{Name: "rel1", Digest: "qux"},
				{Name: "ver3", Digest: "baz"},
			},
			pattern: `ver(\d+)`,
			extract: `$1`,
			expected: []database.Tag{
				{Name: "1", Digest: "foo"},
				{Name: "2", Digest: "bar"},
				{Name: "3", Digest: "baz"},
			},
			origTags: map[string]int{
				"1": 0,
				"2": 1,
				"3": 3,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			g := NewWithT(t)
			filter := newRegexFilter(tt.pattern, tt.extract)
			filter.Apply(tt.tags)
			g.Expect(filter.Items()).To(HaveLen(len(tt.expected)))
			g.Expect(filter.Items()).To(ContainElements(tt.expected))
			for tagKey, idx := range tt.origTags {
				g.Expect(filter.GetOriginalTag(tagKey)).To(Equal(tt.tags[idx]))
			}
		})
	}
}

func newRegexFilter(pattern string, extract string) *RegexFilter {
	f, _ := NewRegexFilter(pattern, extract)
	return f
}
