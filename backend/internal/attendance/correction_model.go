package attendance

import "r3-ti-faceattend/backend/internal/auth"

type AdminAttendanceCorrectionInput struct {
	CheckInTime  string  `json:"check_in_time"`
	CheckOutTime *string `json:"check_out_time"`
	Reason       string  `json:"reason"`
}

type adminAttendanceCorrectionCommand struct {
	AttendanceID string
	CheckInTime  string
	CheckOutTime *string
	Reason       string
	Timezone     string
	Actor        auth.Claims
}
