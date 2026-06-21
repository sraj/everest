package dbx

import (
	"testing"
)

func TestPage_Offset_Default(t *testing.T) {
	p := Page{}
	offset := p.offset()
	if offset != 0 {
		t.Errorf("expected offset 0 for default page, got %d", offset)
	}
}

func TestPage_Offset_FirstPage(t *testing.T) {
	p := Page{Number: 1, Size: 20}
	offset := p.offset()
	if offset != 0 {
		t.Errorf("expected offset 0 for page 1, got %d", offset)
	}
}

func TestPage_Offset_SecondPage(t *testing.T) {
	p := Page{Number: 2, Size: 20}
	offset := p.offset()
	if offset != 20 {
		t.Errorf("expected offset 20 for page 2 with size 20, got %d", offset)
	}
}

func TestPage_Offset_ThirdPage(t *testing.T) {
	p := Page{Number: 3, Size: 50}
	offset := p.offset()
	if offset != 100 {
		t.Errorf("expected offset 100 for page 3 with size 50, got %d", offset)
	}
}

func TestPage_Offset_ZeroPage(t *testing.T) {
	p := Page{Number: 0, Size: 20}
	offset := p.offset()
	if offset != 0 {
		t.Errorf("expected offset 0 for page 0 (treated as 1), got %d", offset)
	}
}

func TestPage_Offset_NegativePage(t *testing.T) {
	p := Page{Number: -5, Size: 20}
	offset := p.offset()
	if offset != 0 {
		t.Errorf("expected offset 0 for negative page (treated as 1), got %d", offset)
	}
}

func TestPage_Size_Default(t *testing.T) {
	p := Page{}
	size := p.size()
	if size != 20 {
		t.Errorf("expected default size 20, got %d", size)
	}
}

func TestPage_Size_Custom(t *testing.T) {
	p := Page{Size: 50}
	size := p.size()
	if size != 50 {
		t.Errorf("expected size 50, got %d", size)
	}
}

func TestPage_Size_Zero(t *testing.T) {
	p := Page{Size: 0}
	size := p.size()
	if size != 20 {
		t.Errorf("expected default size 20 when zero is provided, got %d", size)
	}
}

func TestPage_Size_Negative(t *testing.T) {
	p := Page{Size: -10}
	size := p.size()
	if size != 20 {
		t.Errorf("expected default size 20 when negative is provided, got %d", size)
	}
}

func TestPageResult_Structure(t *testing.T) {
	result := &PageResult[string]{
		Items:      []string{"a", "b", "c"},
		Total:      100,
		Page:       1,
		PageSize:   20,
		TotalPages: 5,
	}

	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
	if result.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Total)
	}
	if result.TotalPages != 5 {
		t.Errorf("expected 5 total pages, got %d", result.TotalPages)
	}
}

func TestPageResult_EmptyItems(t *testing.T) {
	result := &PageResult[int]{
		Items:      []int{},
		Total:      0,
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
	}

	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
	if result.TotalPages != 0 {
		t.Errorf("expected 0 total pages, got %d", result.TotalPages)
	}
}

func TestPage_LargePage(t *testing.T) {
	p := Page{Number: 1000, Size: 100}
	offset := p.offset()
	expected := 99900 // (1000-1) * 100
	if offset != expected {
		t.Errorf("expected offset %d, got %d", expected, offset)
	}
}

func TestPage_OneItemPerPage(t *testing.T) {
	p := Page{Number: 5, Size: 1}
	offset := p.offset()
	if offset != 4 {
		t.Errorf("expected offset 4 for page 5 with size 1, got %d", offset)
	}
}

func TestPageResult_TotalPagesCalculation(t *testing.T) {
	tests := []struct {
		name           string
		total          int
		pageSize       int
		expectedPages  int
	}{
		{"exact pages", 100, 20, 5},
		{"remainder", 101, 20, 6},
		{"one page", 10, 20, 1},
		{"zero total", 0, 20, 0},
		{"large remainder", 99, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &PageResult[string]{
				Total:      tt.total,
				PageSize:   tt.pageSize,
				TotalPages: (tt.total + tt.pageSize - 1) / tt.pageSize,
			}

			if result.TotalPages != tt.expectedPages {
				t.Errorf("expected %d pages, got %d", tt.expectedPages, result.TotalPages)
			}
		})
	}
}

func TestPage_Calculation_Examples(t *testing.T) {
	tests := []struct {
		page     int
		size     int
		offset   int
		name     string
	}{
		{1, 10, 0, "first page"},
		{2, 10, 10, "second page"},
		{5, 20, 80, "fifth page with size 20"},
		{10, 50, 450, "tenth page with size 50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Page{Number: tt.page, Size: tt.size}
			offset := p.offset()
			if offset != tt.offset {
				t.Errorf("expected offset %d, got %d", tt.offset, offset)
			}
		})
	}
}
