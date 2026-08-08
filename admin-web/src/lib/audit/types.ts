export type AuditAction =
  | "ATTENDANCE_CORRECTED"
  | "FACE_ENROLLMENT_RESET";

export type AuditEntityType = "ATTENDANCE_RECORD" | "FACE_PROFILE";

export type AuditLogItem = {
  id: string;
  actor_user_id: string | null;
  actor_email: string;
  actor_role: string;
  action: AuditAction;
  entity_type: AuditEntityType;
  entity_id: string | null;
  target_user_id: string | null;
  target_label: string | null;
  reason: string;
  before_data: Record<string, unknown>;
  after_data: Record<string, unknown>;
  created_at: string;
};

export type AuditLogListResponse = {
  items: AuditLogItem[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type AuditLogQuery = {
  action?: AuditAction;
  entity_type?: AuditEntityType;
  entity_id?: string;
  date_from?: string;
  date_to?: string;
  page?: number;
  page_size?: number;
};
