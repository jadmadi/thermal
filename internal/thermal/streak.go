package thermal

import (
	"math"
	"sort"
	"time"
)

func ComputeStreaks(days map[string]bool) (current int, longest int) {
	sorted := make([]string, 0, len(days))
	for d := range days {
		sorted = append(sorted, d)
	}
	sort.Strings(sorted)

	longest = 0
	run := 0
	var prev time.Time
	for _, d := range sorted {
		t, _ := time.Parse("2006-01-02", d)
		if !prev.IsZero() && prev.AddDate(0, 0, 1).Equal(t) {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
		prev = t
	}

	current = 0
	if len(sorted) > 0 {
		cursor, _ := time.Parse("2006-01-02", sorted[len(sorted)-1])
		for {
			if !days[LocalDay(cursor)] {
				break
			}
			current++
			cursor = cursor.AddDate(0, 0, -1)
		}
	}

	return current, longest
}

func ActivityThresholds(activity map[string]DayActivity) [3]int64 {
	var nonzero []int64
	for _, a := range activity {
		if a.Tokens > 0 {
			nonzero = append(nonzero, a.Tokens)
		}
	}
	if len(nonzero) == 0 {
		return [3]int64{0, 0, 0}
	}
	sort.Slice(nonzero, func(i, j int) bool { return nonzero[i] < nonzero[j] })
	at := func(q float64) int64 {
		idx := int(math.Floor(float64(len(nonzero)-1) * q))
		return nonzero[idx]
	}
	t1, t2, t3 := at(0.25), at(0.5), at(0.75)
	if t2 <= t1 && t1 < nonzero[len(nonzero)-1] {
		t2 = t1 + 1
	}
	if t3 <= t2 && t2 < nonzero[len(nonzero)-1] {
		t3 = t2 + 1
	}
	return [3]int64{t1, t2, t3}
}

func ActivityLevel(a DayActivity, thresholds [3]int64) int {
	if a.Tokens <= 0 {
		return 1
	}
	if a.Tokens <= thresholds[0] {
		return 1
	}
	if a.Tokens <= thresholds[1] {
		return 2
	}
	if a.Tokens <= thresholds[2] {
		return 3
	}
	return 4
}
