// Copyright 2023 The gVisor Authors.
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

// Package rand implements a cryptographically secure pseudorandom number
// generator.
package rand

import (
	"encoding/binary"
	"fmt"
	"io"
)

// RNG exposes convenience functions based on a cryptographically secure
// io.Reader.
type RNG struct {
	Reader io.Reader
}

// RNGFrom returns a new RNG. r must be a cryptographically secure io.Reader.
func RNGFrom(r io.Reader) RNG {
	return RNG{Reader: r}
}

// Uint32 is analogous to the standard library's math/rand.Uint32.
func (rg *RNG) Uint32() uint32 {
	var data [4]byte
	if _, err := rg.Reader.Read(data[:]); err != nil {
		panic(fmt.Sprintf("Read() failed: %v", err))
	}
	// The endianness doesn't matter here as it's random bytes either way.
	return binary.LittleEndian.Uint32(data[:])
}
