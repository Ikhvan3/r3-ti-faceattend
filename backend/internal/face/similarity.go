package face

import "math"

func CosineSimilarity(left []float64, right []float64) (float64, error) {
	if len(left) == 0 || len(right) == 0 {
		return 0, ErrInvalidInput
	}
	if len(left) != len(right) {
		return 0, ErrInvalidDimension
	}

	var dot float64
	var leftNormSquared float64
	var rightNormSquared float64
	for i := range left {
		l := left[i]
		r := right[i]
		if !isFinite(l) || !isFinite(r) {
			return 0, ErrInvalidInput
		}
		dot += l * r
		leftNormSquared += l * l
		rightNormSquared += r * r
	}
	if leftNormSquared == 0 || rightNormSquared == 0 {
		return 0, ErrInvalidInput
	}

	similarity := dot / (math.Sqrt(leftNormSquared) * math.Sqrt(rightNormSquared))
	if !isFinite(similarity) {
		return 0, ErrInvalidInput
	}
	if similarity > 1 && similarity < 1+1e-12 {
		return 1, nil
	}
	if similarity < -1 && similarity > -1-1e-12 {
		return -1, nil
	}
	return similarity, nil
}

func L2Normalize(embedding []float64) ([]float64, error) {
	if len(embedding) == 0 {
		return nil, ErrInvalidInput
	}
	var normSquared float64
	for _, value := range embedding {
		if !isFinite(value) {
			return nil, ErrInvalidInput
		}
		normSquared += value * value
	}
	if normSquared == 0 {
		return nil, ErrInvalidInput
	}
	norm := math.Sqrt(normSquared)
	if !isFinite(norm) {
		return nil, ErrInvalidInput
	}

	normalized := make([]float64, len(embedding))
	for i, value := range embedding {
		normalized[i] = value / norm
	}
	return normalized, nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
