package user

import "time"

type EmployeeFaceEnrollment struct {
	Enrolled         bool       `json:"enrolled"`
	FaceStatus       string     `json:"face_status"`
	EmbeddingModel   string     `json:"embedding_model,omitempty"`
	EmbeddingVersion string     `json:"embedding_version,omitempty"`
	EnrolledAt       *time.Time `json:"enrolled_at,omitempty"`
}

type EmployeeProfile struct {
	ID             string                  `json:"id"`
	EmployeeNumber string                  `json:"employee_number"`
	Name           string                  `json:"name"`
	Email          string                  `json:"email"`
	Phone          *string                 `json:"phone"`
	Position       *string                 `json:"position"`
	Role           Role                    `json:"role"`
	AccountStatus  AccountStatus           `json:"account_status"`
	FaceEnrollment *EmployeeFaceEnrollment `json:"face_enrollment,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type EmployeeListFilter struct {
	Page     int
	PageSize int
	Search   string
	Status   AccountStatus
}

type EmployeeList struct {
	Items      []EmployeeProfile `json:"items"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalItems int               `json:"total_items"`
	TotalPages int               `json:"total_pages"`
}

func safeEmployee(u User) EmployeeProfile {
	return EmployeeProfile{
		ID:             u.ID,
		EmployeeNumber: u.EmployeeNumber,
		Name:           u.Name,
		Email:          u.Email,
		Phone:          u.Phone,
		Position:       u.Position,
		Role:           u.Role,
		AccountStatus:  u.AccountStatus,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}
