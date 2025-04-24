package descriptives

type SummaryStats struct {
	Count    int       `json:"count"`
	Mean     float64   `json:"mean"`
	Median   float64   `json:"median"`
	Mode     []float64 `json:"mode"`
	StdDev   float64   `json:"std_dev"`
	Variance float64   `json:"variance"`
	Min      float64   `json:"min"`
	Max      float64   `json:"max"`
	Range    float64   `json:"range"`
	Sum      float64   `json:"sum"`
}
