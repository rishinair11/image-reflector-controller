/*
Copyright 2020, 2021 The Flux authors

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

	. "github.com/onsi/gomega"

	"github.com/fluxcd/image-reflector-controller/internal/database"
)

func TestNewSemVer(t *testing.T) {
	cases := []struct {
		label        string
		semverRanges []string
		expectErr    bool
	}{
		{
			label:        "With valid range",
			semverRanges: []string{"1.0.x", "^1.0", "=1.0.0", "~1.0", ">=1.0", ">0,<2.0"},
		},
		{
			label:        "With invalid range",
			semverRanges: []string{"1.0.0p", "1x", "x1", "-1", "a", ""},
			expectErr:    true,
		},
	}

	for _, tt := range cases {
		for _, r := range tt.semverRanges {
			t.Run(tt.label, func(t *testing.T) {
				_, err := NewSemVer(r)
				if tt.expectErr && err == nil {
					t.Fatalf("expecting error, got nil for range value: '%s'", r)
				}
				if !tt.expectErr && err != nil {
					t.Fatalf("returned unexpected error: %s", err)
				}
			})
		}
	}
}

func TestSemVer_Latest(t *testing.T) {
	cases := []struct {
		label           string
		semverRange     string
		versions        []database.Tag
		expectedVersion database.Tag
		expectErr       bool
	}{
		{
			label: "With valid format",
			versions: []database.Tag{
				{Name: "1.0.0", Digest: "foo"},
				{Name: "1.0.0.1", Digest: "bar"},
				{Name: "1.0.0p", Digest: "baz"},
				{Name: "1.0.1", Digest: "qux"},
				{Name: "1.2.0", Digest: "faa"},
				{Name: "0.1.0", Digest: "quux"},
			},
			semverRange:     "1.0.x",
			expectedVersion: database.Tag{Name: "1.0.1", Digest: "qux"},
		},
		{
			label: "With valid format prefix",
			versions: []database.Tag{
				{Name: "v1.2.3", Digest: "v1.2.3-digest"},
				{Name: "v1.0.0", Digest: "v1.0.0-digest"},
				{Name: "v0.1.0", Digest: "v0.1.0-dig"},
			},
			semverRange:     "1.0.x",
			expectedVersion: database.Tag{Name: "v1.0.0", Digest: "v1.0.0-digest"},
		},
		{
			label: "With invalid format prefix",
			versions: []database.Tag{
				{Name: "b1.2.3"},
				{Name: "b1.0.0"},
				{Name: "b0.1.0"},
			},
			semverRange: "1.0.x",
			expectErr:   true,
		},
		{
			label:       "With empty list",
			versions:    []database.Tag{},
			semverRange: "1.0.x",
			expectErr:   true,
		},
		{
			label:       "With non-matching version list",
			versions:    []database.Tag{{Name: "1.2.0"}},
			semverRange: "1.0.x",
			expectErr:   true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			g := NewWithT(t)

			policy, err := NewSemVer(tt.semverRange)
			g.Expect(err).NotTo(HaveOccurred())

			latest, err := policy.Latest(tt.versions)
			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(latest).To(Equal(&tt.expectedVersion), "incorrect computed version returned")
		})
	}
}
