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

import (
	"errors"
	"fmt"
	"time"
)

var errDurationEmpty = errors.New("duration empty")

var durationUnits = map[string]time.Duration{
	"s":   time.Second,
	"min": time.Minute,
	"m":   time.Minute,
	"h":   time.Hour,
	"d":   time.Hour * 24,
	"w":   time.Hour * 24 * 7,
	"mon": time.Hour * 24 * 30,
	"y":   time.Hour * 24 * 365,
	"us":  time.Microsecond,
	"ms":  time.Millisecond,
	"ns":  time.Nanosecond,
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// ParseExtendedDuration parses a duration, with the ability to specify time
// units in days, weeks, months, and years.
func ParseExtendedDuration(s string) (time.Duration, error) {
	if len(s) == 0 {
		return 0, errDurationEmpty
	}

	var d time.Duration
	i := 0

	for i < len(s) {
		if !isDigit(s[i]) {
			return 0, fmt.Errorf("invalid duration %s, no value specified", s)
		}

		// Consume [0-9]+
		n := 0
		for i < len(s) && isDigit(s[i]) {
			n *= 10
			n += int(s[i]) - int('0')
			i++
		}

		// Consume [^0-9]+ and convert into a unit
		if i == len(s) {
			return 0, fmt.Errorf("invalid duration %s, no unit", s)
		}

		unitStart := i
		for i < len(s) && !isDigit(s[i]) {
			i++
		}

		unitText := s[unitStart:i]
		unit, unitExists := durationUnits[unitText]
		if !unitExists {
			return 0, fmt.Errorf("invalid duration %s, invalid unit %s", s, unitText)
		}

		d += time.Duration(n) * unit

	}

	return d, nil
}
