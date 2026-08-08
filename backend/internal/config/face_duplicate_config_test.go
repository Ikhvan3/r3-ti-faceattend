package config

import (
	"math"
	"testing"
)

func TestFaceDuplicateConfigValidate(t *testing.T) {
	if err := (FaceDuplicateConfig{Threshold: 0.8, SearchTopK: 20}).Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestFaceDuplicateConfigValidateRejectsInvalidThreshold(t *testing.T) {
	for _, value := range []float64{math.NaN(), math.Inf(1), -1.1, 1.1} {
		if err := (FaceDuplicateConfig{Threshold: value, SearchTopK: 20}).Validate(); err == nil {
			t.Fatalf("Validate() error = nil for threshold %v", value)
		}
	}
}

func TestFaceDuplicateConfigValidateRejectsInvalidTopK(t *testing.T) {
	for _, value := range []int{0, -1, 101} {
		if err := (FaceDuplicateConfig{Threshold: 0.8, SearchTopK: value}).Validate(); err == nil {
			t.Fatalf("Validate() error = nil for topK %d", value)
		}
	}
}
