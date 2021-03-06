// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package time

import "container/list"

// RangeIter iterates over a collection of time ranges.
type RangeIter interface {
	// Next moves to the next item.
	Next() bool

	// Value returns the current time range.
	Value() Range
}

type rangeIter struct {
	ranges *list.List
	cur    *list.Element
}

func newRangeIter(ranges *list.List) RangeIter {
	return &rangeIter{ranges: ranges}
}

// Next moves to the next item.
func (it *rangeIter) Next() bool {
	if it.ranges == nil {
		return false
	}
	if it.cur == nil {
		it.cur = it.ranges.Front()
	} else {
		it.cur = it.cur.Next()
	}
	return it.cur != nil
}

// Value returns the current time range.
func (it *rangeIter) Value() Range {
	return it.cur.Value.(Range)
}
