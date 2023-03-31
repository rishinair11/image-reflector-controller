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
	"fmt"
	"sort"

	"github.com/fluxcd/image-reflector-controller/internal/database"
)

const (
	// AlphabeticalOrderAsc ascending order
	AlphabeticalOrderAsc = "ASC"
	// AlphabeticalOrderDesc descending order
	AlphabeticalOrderDesc = "DESC"
)

// Alphabetical representes a alphabetical ordering policy
type Alphabetical struct {
	Order string
}

var _ Policer = &Alphabetical{}

// NewAlphabetical constructs a Alphabetical object validating the provided
// order argument
func NewAlphabetical(order string) (*Alphabetical, error) {
	switch order {
	case "":
		order = AlphabeticalOrderAsc
	case AlphabeticalOrderAsc, AlphabeticalOrderDesc:
		break
	default:
		return nil, fmt.Errorf("invalid order argument provided: '%s', must be one of: %s, %s", order, AlphabeticalOrderAsc, AlphabeticalOrderDesc)
	}

	return &Alphabetical{
		Order: order,
	}, nil
}

// Latest returns latest version from a provided list of strings
func (p *Alphabetical) Latest(versions []database.Tag) (*database.Tag, error) {
	if len(versions) == 0 {
		return nil, fmt.Errorf("version list argument cannot be empty")
	}

	tagNames := make([]string, len(versions))
	tagsByName := make(map[string]database.Tag, len(versions))
	for idx, name := range versions {
		tagNames[idx] = name.Name
		tagsByName[name.Name] = name
	}

	var sorted sort.StringSlice = tagNames
	if p.Order == AlphabeticalOrderDesc {
		sort.Sort(sorted)
	} else {
		sort.Sort(sort.Reverse(sorted))
	}
	selected := tagsByName[sorted[0]]

	return &selected, nil
}
