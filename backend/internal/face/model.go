package face

import "time"

type FaceStatus string

const (
	FaceStatusNotEnrolled FaceStatus = "NOT_ENROLLED"
	FaceStatusEnrolled    FaceStatus = "ENROLLED"
)

type FaceProfile struct {
	ID               string
	UserID           string
	Embedding        []float64
	EmbeddingModel   string
	EmbeddingVersion string
	Status           FaceStatus
	EnrolledAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type EnrollmentInput struct {
	Embedding        []float64
	EmbeddingModel   string
	EmbeddingVersion string
}

type StatusResponse struct {
	Enrolled         bool       `json:"enrolled"`
	FaceStatus       FaceStatus `json:"face_status"`
	EmbeddingModel   string     `json:"embedding_model,omitempty"`
	EmbeddingVersion string     `json:"embedding_version,omitempty"`
	EnrolledAt       *time.Time `json:"enrolled_at,omitempty"`
}

type SupportedModel struct {
	Name      string
	Version   string
	Dimension int
}

const (
	FaceNetModelName      = "facenet"
	FaceNetModelVersion   = "shubham0204-facenet-2020-fp32"
	FaceNetModelDimension = 128
)

func ProductionModelRegistry() ModelRegistry {
	return NewModelRegistry([]SupportedModel{{
		Name:      FaceNetModelName,
		Version:   FaceNetModelVersion,
		Dimension: FaceNetModelDimension,
	}})
}

type ModelRegistry struct {
	models map[string]SupportedModel
}

func NewModelRegistry(models []SupportedModel) ModelRegistry {
	registry := ModelRegistry{models: make(map[string]SupportedModel, len(models))}
	for _, model := range models {
		registry.models[modelKey(model.Name, model.Version)] = model
	}
	return registry
}

func EmptyModelRegistry() ModelRegistry {
	return NewModelRegistry(nil)
}

func (r ModelRegistry) Find(name string, version string) (SupportedModel, bool) {
	model, ok := r.models[modelKey(name, version)]
	return model, ok
}

func modelKey(name string, version string) string {
	return name + "\x00" + version
}
