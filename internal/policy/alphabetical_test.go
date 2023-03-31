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

func TestNewAlphabetical(t *testing.T) {
	cases := []struct {
		label     string
		order     string
		expectErr bool
	}{
		{
			label: "With valid empty order",
			order: "",
		},
		{
			label: "With valid asc order",
			order: AlphabeticalOrderAsc,
		},
		{
			label: "With valid desc order",
			order: AlphabeticalOrderDesc,
		},
		{
			label:     "With invalid order",
			order:     "invalid",
			expectErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			_, err := NewAlphabetical(tt.order)
			if tt.expectErr && err == nil {
				t.Fatalf("expecting error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}
		})
	}
}

func TestAlphabetical_Latest(t *testing.T) {
	cases := []struct {
		label           string
		order           string
		versions        []database.Tag
		expectedVersion database.Tag
		expectErr       bool
	}{
		{
			label: "With Ubuntu CalVer",
			versions: []database.Tag{
				{Name: "16.04"},
				{Name: "16.04.1"},
				{Name: "16.10"},
				{Name: "20.04"},
				{Name: "20.10", Digest: "20.10-dig"},
			},
			expectedVersion: database.Tag{Name: "20.10", Digest: "20.10-dig"},
		},
		{
			label: "With Ubuntu CalVer descending",
			versions: []database.Tag{
				{Name: "16.04", Digest: "16.04-dig"},
				{Name: "16.04.1"},
				{Name: "16.10"},
				{Name: "20.04"},
				{Name: "20.10"},
			},
			order:           AlphabeticalOrderDesc,
			expectedVersion: database.Tag{Name: "16.04", Digest: "16.04-dig"},
		},
		{
			label: "With Ubuntu code names",
			versions: []database.Tag{
				{Name: "xenial"},
				{Name: "yakkety"},
				{Name: "zesty", Digest: "dig"},
				{Name: "artful"},
				{Name: "bionic"},
			},
			expectedVersion: database.Tag{Name: "zesty", Digest: "dig"},
		},
		{
			label: "With Ubuntu code names descending",
			versions: []database.Tag{
				{Name: "xenial"},
				{Name: "yakkety"},
				{Name: "zesty"},
				{Name: "artful", Digest: "aec070645fe53ee3b3763059376134f058cc337"},
				{Name: "bionic"},
			},
			order:           AlphabeticalOrderDesc,
			expectedVersion: database.Tag{Name: "artful", Digest: "aec070645fe53ee3b3763059376134f058cc337"},
		},
		{
			label: "With Timestamps",
			versions: []database.Tag{
				{Name: "1606234201"},
				{Name: "1606364286", Digest: "1606364286-33383"},
				{Name: "1606334092"},
				{Name: "1606334284"},
				{Name: "1606334201"},
			},
			expectedVersion: database.Tag{Name: "1606364286", Digest: "1606364286-33383"},
		},
		{
			label: "With Unix Timestamps desc",
			versions: []database.Tag{
				{Name: "1606234201", Digest: "1606234201@494781"},
				{Name: "1606364286"},
				{Name: "1606334092"},
				{Name: "1606334284"},
				{Name: "1606334201"},
			},
			order:           AlphabeticalOrderDesc,
			expectedVersion: database.Tag{Name: "1606234201", Digest: "1606234201@494781"},
		},
		{
			label: "With Unix Timestamps prefix",
			versions: []database.Tag{
				{Name: "rel-1606234201"},
				{Name: "rel-1606364286", Digest: "80f95d07201fd766c027279813220d6fc6826038e45cdd4f2b78e00297beb337"},
				{Name: "rel-1606334092"},
				{Name: "rel-1606334284"},
				{Name: "rel-1606334201"},
			},
			expectedVersion: database.Tag{Name: "rel-1606364286", Digest: "80f95d07201fd766c027279813220d6fc6826038e45cdd4f2b78e00297beb337"},
		},
		{
			label: "With RFC3339",
			versions: []database.Tag{
				{Name: "2021-01-08T21-18-21Z"},
				{Name: "2020-05-08T21-18-21Z"},
				{Name: "2021-01-08T19-20-00Z"},
				{Name: "1990-01-08T00-20-00Z"},
				{Name: "2023-05-08T00-20-00Z", Digest: "MjAyMy0wNS0wOFQwMC0yMC0wMFo="},
			},
			expectedVersion: database.Tag{Name: "2023-05-08T00-20-00Z", Digest: "MjAyMy0wNS0wOFQwMC0yMC0wMFo="},
		},
		{
			label: "With RFC3339 desc",
			versions: []database.Tag{
				{Name: "2021-01-08T21-18-21Z", Digest: "0"},
				{Name: "2020-05-08T21-18-21Z", Digest: "1"},
				{Name: "2021-01-08T19-20-00Z", Digest: "2"},
				{Name: "1990-01-08T00-20-00Z", Digest: "3"},
				{Name: "2023-05-08T00-20-00Z", Digest: "4"},
			},
			order:           AlphabeticalOrderDesc,
			expectedVersion: database.Tag{Name: "1990-01-08T00-20-00Z", Digest: "3"},
		},
		{
			label:     "Empty version list",
			versions:  []database.Tag{},
			expectErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			g := NewWithT(t)

			policy, err := NewAlphabetical(tt.order)
			if err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}
			latest, err := policy.Latest(tt.versions)
			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(latest).To(Equal(&tt.expectedVersion))
		})
	}
}
