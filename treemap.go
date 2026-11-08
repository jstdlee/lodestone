package main

import "math"

// treemapCell is one rectangle of a squarified treemap. index is into the
// sizes slice passed to layoutTreemap.
type treemapCell struct {
	index      int
	x, y, w, h float32
}

// layoutTreemap packs sizes into width×height. Zero sizes are skipped.
// Areas sum to the rectangle. Larger values are placed first, and index
// points back at the input.
func layoutTreemap(sizes []int64, width, height float32) []treemapCell {
	if width < 1 || height < 1 {
		return nil
	}
	var total int64
	items := make([]sizedItem, 0, len(sizes))
	for i, n := range sizes {
		if n <= 0 {
			continue
		}
		total += n
		items = append(items, sizedItem{index: i, bytes: n})
	}
	if total == 0 || len(items) == 0 {
		return nil
	}
	for i := 1; i < len(items); i++ {
		j := i
		for j > 0 && items[j].bytes > items[j-1].bytes {
			items[j], items[j-1] = items[j-1], items[j]
			j--
		}
	}
	scale := float64(width) * float64(height) / float64(total)
	for i := range items {
		items[i].area = float64(items[i].bytes) * scale
	}
	var cells []treemapCell
	squarify(items, 0, 0, float64(width), float64(height), &cells)
	return cells
}

type sizedItem struct {
	index int
	bytes int64
	area  float64
}

func squarify(items []sizedItem, x, y, w, h float64, out *[]treemapCell) {
	if len(items) == 0 || w <= 0 || h <= 0 {
		return
	}
	if len(items) == 1 {
		*out = append(*out, treemapCell{
			index: items[0].index,
			x:     float32(x), y: float32(y), w: float32(w), h: float32(h),
		})
		return
	}
	side := h
	if w < h {
		side = w
	}
	row := items[:1]
	rest := items[1:]
	worst := worstRatio(row, side)
	for len(rest) > 0 {
		next := items[:len(row)+1]
		if nw := worstRatio(next, side); nw <= worst {
			row = next
			rest = items[len(row):]
			worst = nw
			continue
		}
		break
	}
	rowArea := 0.0
	for _, it := range row {
		rowArea += it.area
	}
	if w >= h {
		rw := rowArea / h
		if rw > w || rw < 0.5 {
			rw = w
		}
		placeColumn(row, x, y, rw, h, out)
		if w-rw < 0.5 {
			return
		}
		squarify(rest, x+rw, y, w-rw, h, out)
		return
	}
	rh := rowArea / w
	if rh > h || rh < 0.5 {
		rh = h
	}
	placeRow(row, x, y, w, rh, out)
	if h-rh < 0.5 {
		return
	}
	squarify(rest, x, y+rh, w, h-rh, out)
}

func placeColumn(row []sizedItem, x, y, w, h float64, out *[]treemapCell) {
	yy := y
	for i, it := range row {
		ih := 0.0
		if w > 0 {
			ih = it.area / w
		}
		if i == len(row)-1 || yy+ih > y+h {
			ih = y + h - yy
		}
		if ih < 0 {
			ih = 0
		}
		*out = append(*out, treemapCell{
			index: it.index,
			x:     float32(x), y: float32(yy), w: float32(w), h: float32(ih),
		})
		yy += ih
	}
}

func placeRow(row []sizedItem, x, y, w, h float64, out *[]treemapCell) {
	xx := x
	for i, it := range row {
		iw := 0.0
		if h > 0 {
			iw = it.area / h
		}
		if i == len(row)-1 || xx+iw > x+w {
			iw = x + w - xx
		}
		if iw < 0 {
			iw = 0
		}
		*out = append(*out, treemapCell{
			index: it.index,
			x:     float32(xx), y: float32(y), w: float32(iw), h: float32(h),
		})
		xx += iw
	}
}

func worstRatio(row []sizedItem, side float64) float64 {
	if side <= 0 {
		return math.Inf(1)
	}
	sum := 0.0
	minA := math.MaxFloat64
	maxA := 0.0
	for _, it := range row {
		sum += it.area
		if it.area < minA {
			minA = it.area
		}
		if it.area > maxA {
			maxA = it.area
		}
	}
	if sum <= 0 || minA <= 0 {
		return math.Inf(1)
	}
	side2 := side * side
	sum2 := sum * sum
	a := side2 * maxA / sum2
	b := sum2 / (side2 * minA)
	if a > b {
		return a
	}
	return b
}
