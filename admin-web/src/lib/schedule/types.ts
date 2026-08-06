import type { Employee } from "@/lib/employee/types";

export type ScheduleStatus = "ACTIVE" | "INACTIVE";
export type AssignmentStatus = "CURRENT" | "UPCOMING" | "ENDED";

export type WorkSchedule = {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  grace_minutes: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type WorkScheduleListResponse = {
  items: WorkSchedule[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type WorkScheduleListQuery = {
  page?: number;
  page_size?: number;
  search?: string;
  status?: ScheduleStatus;
};

export type CreateWorkScheduleRequest = {
  name: string;
  start_time: string;
  end_time: string;
  grace_minutes: number;
};

export type UpdateWorkScheduleRequest = CreateWorkScheduleRequest;

export type UpdateWorkScheduleStatusRequest = {
  is_active: boolean;
};

export type ScheduleAssignment = {
  id: string;
  user: Employee;
  schedule: WorkSchedule;
  effective_from: string;
  effective_to: string | null;
  status: AssignmentStatus;
  created_at: string;
  updated_at: string;
};

export type ScheduleAssignmentListResponse = {
  items: ScheduleAssignment[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type ScheduleAssignmentListQuery = {
  page?: number;
  page_size?: number;
  search?: string;
  user_id?: string;
  schedule_id?: string;
  status?: AssignmentStatus;
};

export type CreateScheduleAssignmentRequest = {
  user_id: string;
  schedule_id: string;
  effective_from: string;
  effective_to: string | null;
};

export type EndScheduleAssignmentRequest = {
  effective_to: string;
};
