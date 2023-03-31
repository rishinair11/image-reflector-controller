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
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/fluxcd/image-reflector-controller/internal/database"
)

func TestNewNumerical(t *testing.T) {
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
			order: NumericalOrderAsc,
		},
		{
			label: "With valid desc order",
			order: NumericalOrderDesc,
		},
		{
			label:     "With invalid order",
			order:     "invalid",
			expectErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.label, func(t *testing.T) {
			_, err := NewNumerical(tt.order)
			if tt.expectErr && err == nil {
				t.Fatalf("expecting error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("returned unexpected error: %s", err)
			}
		})
	}
}

func TestNumerical_Latest(t *testing.T) {
	cases := []struct {
		label           string
		order           string
		versions        []database.Tag
		expectedVersion database.Tag
		expectErr       bool
	}{
		{
			label: "With unordered list of integers ascending",
			versions: shuffle([]database.Tag{
				{Name: "-62"},
				{Name: "-88"},
				{Name: "73", Digest: "foodigest"},
				{Name: "72"},
				{Name: "15"},
				{Name: "16"},
				{Name: "15"},
				{Name: "29"},
				{Name: "-33"},
				{Name: "-91"},
			}),
			expectedVersion: database.Tag{Name: "73", Digest: "foodigest"},
		},
		{
			label: "With unordered list of integers descending",
			versions: shuffle([]database.Tag{
				{Name: "5"},
				{Name: "-8"},
				{Name: "-78", Digest: "somedig"},
				{Name: "25"},
				{Name: "70"},
				{Name: "-4"},
				{Name: "80"},
				{Name: "92"},
				{Name: "-20"},
				{Name: "-24"},
			}),
			order:           NumericalOrderDesc,
			expectedVersion: database.Tag{Name: "-78", Digest: "somedig"},
		},
		{
			label: "With unordered list of floats ascending",
			versions: shuffle([]database.Tag{
				{Name: "47.40896403322944"},
				{Name: "-27.8520927455902"},
				{Name: "-27.930666514224427"},
				{Name: "-31.352485948094568"},
				{Name: "-50.41072694704882"},
				{Name: "-21.962849842263736"},
				{Name: "24.71884721436865"},
				{Name: "-39.99177354004344"},
				{Name: "53.47333823144817", Digest: "47333823144817"},
				{Name: "3.2008658570411086"},
			}),
			expectedVersion: database.Tag{Name: "53.47333823144817", Digest: "47333823144817"},
		},
		{
			label: "With unordered list of floats descending",
			versions: shuffle([]database.Tag{
				{Name: "-65.27202780220686"},
				{Name: "57.82948329142309"},
				{Name: "22.40184684363291"},
				{Name: "-86.36934305697784"},
				{Name: "-90.29082099756083", Digest: "-90"},
				{Name: "-12.041712603564264"},
				{Name: "77.70488240399305"},
				{Name: "-38.98425003883552"},
				{Name: "16.06867070412028"},
				{Name: "53.735674335181216"},
			}),
			order:           NumericalOrderDesc,
			expectedVersion: database.Tag{Name: "-90.29082099756083", Digest: "-90"},
		},
		{
			label: "With Unix Timestamps ascending",
			versions: shuffle([]database.Tag{
				{Name: "1606234201"},
				{Name: "1606364286", Digest: "find-me"},
				{Name: "1606334092"},
				{Name: "1606334284"},
				{Name: "1606334201"},
			}),
			expectedVersion: database.Tag{Name: "1606364286", Digest: "find-me"},
		},
		{
			label: "With Unix Timestamps descending",
			versions: shuffle([]database.Tag{
				{Name: "1606234201", Digest: "foobar"},
				{Name: "1606364286"},
				{Name: "1606334092"},
				{Name: "1606334284"},
				{Name: "1606334201"},
			}),
			order:           NumericalOrderDesc,
			expectedVersion: database.Tag{Name: "1606234201", Digest: "foobar"},
		},
		{
			label:           "With single value ascending",
			versions:        []database.Tag{{Name: "1"}},
			expectedVersion: database.Tag{Name: "1"},
		},
		{
			label:           "With single value descending",
			versions:        []database.Tag{{Name: "1"}},
			order:           NumericalOrderDesc,
			expectedVersion: database.Tag{Name: "1"},
		},
		{
			label: "With invalid numerical value",
			versions: []database.Tag{{Name: "0"},
				{Name: "1a"},
				{Name: "b"},
			},
			expectErr: true,
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

			policy, err := NewNumerical(tt.order)
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

func shuffle(list []database.Tag) []database.Tag {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	return list
}
