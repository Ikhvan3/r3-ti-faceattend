package audit

import "time"

type Action string

type EntityType string

const (
	ActionAttendanceCorrected Action = "ATTENDANCE_CORRECTED"
	ActionFaceEnrollmentReset Action = "FACE_ENROLLMENT_RESET"
)

const (
	EntityAttendanceRecord EntityType = "ATTENDANCE_RECORD"
	EntityFaceProfile      EntityType = "FACE_PROFILE"
)

type Event struct {
	ActorUserID  string
	ActorEmail   string
	ActorRole    string
	Action       Action
	EntityType   EntityType
	EntityID     string
	TargetUserID string
	TargetLabel  string
	Reason       string
	BeforeData   map[string]any
	AfterData    map[string]any
}

type Log struct {
	ID           string         `json:"id"`
	ActorUserID  *string        `json:"actor_user_id"`
	ActorEmail   string         `json:"actor_email"`
	ActorRole    string         `json:"actor_role"`
	Action       Action         `json:"action"`
	EntityType   EntityType     `json:"entity_type"`
	EntityID     *string        `json:"entity_id"`
	TargetUserID *string        `json:"target_user_id"`
	TargetLabel  *string        `json:"target_label"`
	Reason       string         `json:"reason"`
	BeforeData   map[string]any `json:"before_data"`
	AfterData    map[string]any `json:"after_data"`
	CreatedAt    time.Time      `json:"created_at"`
}

type ListFilter struct {
	Action     Action
	EntityType EntityType
	EntityID   string
	DateFrom   string
	DateTo     string
	Page       int
	PageSize   int
}

type List struct {
	Items      []Log `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int   `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type listQuery struct {
	Action     Action
	EntityType EntityType
	EntityID   string
	DateFrom   *time.Time
	DateTo     *time.Time
	Page       int
	PageSize   int
}
