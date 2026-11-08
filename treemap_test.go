package main

import "testing"

func TestTreemapCoversRect(t *testing.T) {
	sizes := []int64{1, 1, 2, 4}
	const w, h float32 = 200, 100
	cells := layoutTreemap(sizes, w, h)
	if len(cells) != 4 {
		t.Fatalf("cells %d", len(cells))
	}
	seen := map[int]bool{}
	var area float32
	for _, c := range cells {
		if c.w <= 0 || c.h <= 0 {
			t.Fatalf("empty %+v", c)
		}
		if c.x < -0.2 || c.y < -0.2 || c.x+c.w > w+0.2 || c.y+c.h > h+0.2 {
			t.Fatalf("outside %+v", c)
		}
		if seen[c.index] {
			t.Fatalf("duplicate %d", c.index)
		}
		seen[c.index] = true
		area += c.w * c.h
	}
	if area < w*h*0.98 || area > w*h*1.02 {
		t.Fatalf("area %v", area)
	}
}

func TestTreemapSkipsEmpty(t *testing.T) {
	cells := layoutTreemap([]int64{0, 0}, 100, 50)
	if cells != nil {
		t.Fatalf("cells %+v", cells)
	}
	one := layoutTreemap([]int64{0, 5}, 80, 40)
	if len(one) != 1 || one[0].index != 1 || one[0].w < 79 || one[0].h < 39 {
		t.Fatalf("one %+v", one)
	}
}
