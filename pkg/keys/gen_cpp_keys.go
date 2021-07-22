// Copyright 2017 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// +build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func rocksdbSlice(key []byte) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `"`)
	for _, v := range key {
		fmt.Fprintf(&buf, "\\x%02x", v)
	}
	fmt.Fprintf(&buf, `", %d`, len(key))
	return buf.String()
}

func main() {
	f, err := os.Create("../../c-deps/libroach/keys.h")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file: ", err)
		os.Exit(1)
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "Error closing file: ", err)
			os.Exit(1)
		}
	}()

	// First comment for github/Go; second for reviewable.
	// https://github.com/golang/go/issues/13560#issuecomment-277804473
	// https://github.com/Reviewable/Reviewable/wiki/FAQ#how-do-i-tell-reviewable-that-a-file-is-generated-and-should-not-be-reviewed
	//
	// The funky string concatenation is to foil reviewable so that it doesn't
	// think this file is generated.
	fmt.Fprintf(f, `// Code generated by gen_cpp_keys.go; `+`DO `+`NOT `+`EDIT.
// `+`GENERATED `+`FILE `+`DO `+`NOT `+`EDIT

#pragma once

namespace cockroach {

`)

	genKey := func(key roachpb.Key, name string) {
		fmt.Fprintf(f, "const rocksdb::Slice k%s(%s);\n", name, rocksdbSlice(key))
	}
	genKey(keys.LocalMax, "LocalMax")
	genKey(keys.LocalRangeIDPrefix.AsRawKey(), "LocalRangeIDPrefix")
	genKey(keys.LocalRangeIDReplicatedInfix, "LocalRangeIDReplicatedInfix")
	genKey(keys.LocalRangeAppliedStateSuffix, "LocalRangeAppliedStateSuffix")
	genKey(keys.Meta2KeyMax, "Meta2KeyMax")
	genKey(keys.MaxKey, "MaxKey")
	fmt.Fprintf(f, "\n")

	genSortedSpans := func(spans []roachpb.Span, name string) {
		// Sort the spans by end key which reduces the number of comparisons on
		// libroach/db.cc:IsValidSplitKey().
		sortedSpans := spans
		sort.Slice(sortedSpans, func(i, j int) bool {
			return sortedSpans[i].EndKey.Compare(sortedSpans[j].EndKey) > 0
		})

		fmt.Fprintf(f, "const std::vector<std::pair<rocksdb::Slice, rocksdb::Slice> > kSorted%s = {\n", name)
		for _, span := range sortedSpans {
			fmt.Fprintf(f, "  std::make_pair(rocksdb::Slice(%s), rocksdb::Slice(%s)),\n",
				rocksdbSlice(span.Key), rocksdbSlice(span.EndKey))
		}
		fmt.Fprintf(f, "};\n")
	}
	genSortedSpans(keys.NoSplitSpans, "NoSplitSpans")

	fmt.Fprintf(f, `
}  // namespace cockroach
`)
}
