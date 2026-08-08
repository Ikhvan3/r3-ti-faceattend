export type AdminAttendanceState =
  | "NOT_CHECKED_IN"
  | "CHECKED_IN"
  | "CHECKED_OUT";

export type AdminAttendanceSummary = {
  date: string;
  active_employees: number;
  checked_in: number;
  checked_out: number;
  not_checked_in: number;
  late: number;
};

export type AdminAttendanceEmployee = {
  id: string;
  employee_number: string;
  name: string;
  email: string;
  position: string | null;
};

export type AdminAttendanceSchedule = {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  grace_minutes: number;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
};

export type AdminAttendanceOfficeLocation = {
  id: string;
  name: string;
  radius_meters: number;
};

export type AdminAttendanceLocationEvidence = {
  office_location_id: string;
  office_location_name: string;
  radius_meters: number;
  latitude: number;
  longitude: number;
  accuracy_meters: number;
  distance_meters: number;
  inside_geofence: boolean;
};

export type AdminAttendanceListItem = {
  id: string | null;
  attendance_date: string;
  employee: AdminAttendanceEmployee;
  schedule: AdminAttendanceSchedule;
  check_in_at: string | null;
  check_out_at: string | null;
  attendance_state: AdminAttendanceState;
  is_late: boolean;
  office_location?: AdminAttendanceOfficeLocation | null;
};

export type AdminAttendanceDetail = {
  id: string;
  attendance_date: string;
  employee: AdminAttendanceEmployee;
  schedule: AdminAttendanceSchedule;
  check_in_at: string;
  check_out_at: string | null;
  attendance_state: AdminAttendanceState;
  is_late: boolean;
  check_in_location: AdminAttendanceLocationEvidence | null;
  check_out_location: AdminAttendanceLocationEvidence | null;
};

export type AdminAttendanceCorrectionInput = {
  check_in_time: string;
  check_out_time: string | null;
  reason: string;
};

export type AdminAttendanceListResponse = {
  items: AdminAttendanceListItem[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type AdminAttendanceListQuery = {
  date_from?: string;
  date_to?: string;
  employee_id?: string;
  search?: string;
  attendance_state?: AdminAttendanceState;
  is_late?: boolean;
  page?: number;
  page_size?: number;
};
