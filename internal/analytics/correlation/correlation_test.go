package correlation_test

import (
	"math"
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/analytics/correlation"
)

func TestPearsonCorrelation(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}
	corr, err := correlation.PearsonCorrelation(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(corr-1.0) > 1e-9 {
		t.Errorf("expected correlation 1.0, got %f", corr)
	}
}

func TestSpearmanCorrelation(t *testing.T) {
	x := []float64{10, 20, 30, 40, 50}
	y := []float64{100, 200, 300, 400, 500}
	corr, err := correlation.SpearmanCorrelation(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(corr-1.0) > 1e-9 {
		t.Errorf("expected correlation 1.0, got %f", corr)
	}
}

func TestCorrelationMatrix(t *testing.T) {
	data := [][]float64{
		{1, 2, 3},
		{2, 4, 6},
		{3, 6, 9},
	}
	cols := []int{0, 1, 2}
	matrix, err := correlation.CorrelationMatrix(data, cols, "pearson")
	if err != nil {
		t.Fatal(err)
	}
	if len(matrix) != 3 || len(matrix[0]) != 3 {
		t.Errorf("unexpected matrix dimensions")
	}
	if math.Abs(matrix[0][1]-1.0) > 1e-9 {
		t.Errorf("expected high correlation, got %f", matrix[0][1])
	}
}

func TestExtractFloatColumns(t *testing.T) {
	data := [][]float64{
		{1, 2},
		{3, 4},
	}
	cols, err := correlation.ExtractFloatColumns(data, []int{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 2 || len(cols[0]) != 2 {
		t.Errorf("unexpected column lengths")
	}
	if cols[0][0] != 1.0 || cols[1][1] != 4.0 {
		t.Errorf("unexpected values: %+v", cols)
	}
}

func TestRank(t *testing.T) {
	data := []float64{10, 20, 20, 30}
	ranked, err := correlation.Rank(data)
	if err != nil {
		t.Fatal(err)
	}
	expected := []float64{0, 1.5, 1.5, 3}
	for i := range ranked {
		if math.Abs(ranked[i]-expected[i]) > 1e-9 {
			t.Errorf("expected rank %.1f, got %.1f at index %d", expected[i], ranked[i], i)
		}
	}
}

func TestGenerateCorrelationLabels(t *testing.T) {
	headers := []string{"A", "B", "C"}
	indices := []int{0, 2}
	labels := correlation.GenerateCorrelationLabels(indices, headers)
	if labels[0][1] != "A" || labels[0][2] != "C" {
		t.Errorf("unexpected header labels: %+v", labels[0])
	}
	if labels[1][0] != "A" || labels[2][0] != "C" {
		t.Errorf("unexpected row labels: %+v", labels)
	}
}

func TestFormatCorrelationResult(t *testing.T) {
	matrix := [][]float64{{1.0, 0.5}, {0.5, 1.0}}
	headers := []string{"A", "B", "C"}
	indices := []int{0, 2}
	result := correlation.FormatCorrelationResult(matrix, indices, headers)
	if len(result.Labels) != 2 || result.Labels[1] != "C" {
		t.Errorf("unexpected labels: %+v", result.Labels)
	}
	if len(result.Matrix) != 2 || len(result.Matrix[0]) != 2 {
		t.Errorf("unexpected matrix size")
	}
}
