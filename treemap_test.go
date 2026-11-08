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

func TestTreemapUsesLogAreas(t *testing.T) {
	const w, h float32 = 900, 420
	cells := layoutTreemap([]int64{1 << 20, 1 << 30}, w, h)
	if len(cells) != 2 {
		t.Fatalf("cells %d", len(cells))
	}
	var smallA, bigA float32
	var small treemapCell
	for _, c := range cells {
		if c.index == 0 {
			smallA = c.w * c.h
			small = c
			continue
		}
		bigA = c.w * c.h
	}
	ratio := bigA / smallA
	if ratio < 9 || ratio > 13 {
		t.Fatalf("area ratio %v, want about log2(1024)+1 = 11", ratio)
	}
	if small.w < 48 || small.h < 40 {
		t.Fatalf("small block too thin %+v", small)
	}
}

func TestTreemapRelativeLogSpreadsSameOrder(t *testing.T) {
	cells := layoutTreemap([]int64{3 << 30, 30 << 30}, 800, 400)
	if len(cells) != 2 {
		t.Fatalf("cells %d", len(cells))
	}
	var smallA, bigA float32
	for _, c := range cells {
		if c.index == 0 {
			smallA = c.w * c.h
			continue
		}
		bigA = c.w * c.h
	}
	ratio := bigA / smallA
	if ratio < 3.4 || ratio > 5.2 {
		t.Fatalf("area ratio %v, want about log2(10)+1 = 4.3", ratio)
	}
}

func TestTreemapLogKeepsLabels(t *testing.T) {
	sizes := []int64{
		16 << 30, 16 << 30, 8 << 30, 7 << 30, 6 << 30,
		4 << 30, 3 << 30, 2300 << 20, 2 << 30, 1500 << 20,
		100 << 20,
	}
	cells := layoutTreemap(sizes, 1100, 420)
	if len(cells) != len(sizes) {
		t.Fatalf("cells %d", len(cells))
	}
	for _, c := range cells {
		if c.w < 36 || c.h < 28 {
			t.Fatalf("label will not fit %+v", c)
		}
	}
}
