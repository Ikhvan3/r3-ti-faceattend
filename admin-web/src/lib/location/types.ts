import type { Employee } from "@/lib/employee/types";
import type { AssignmentStatus } from "@/lib/schedule/types";

export type OfficeLocationStatus = "ACTIVE" | "INACTIVE";

export type OfficeLocation = {
  id: string;
  name: string;
  address: string | null;
  latitude: number;
  longitude: number;
  radius_meters: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type OfficeLocationListResponse = {
  items: OfficeLocation[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type OfficeLocationListQuery = {
  page?: number;
  page_size?: number;
  search?: string;
  status?: OfficeLocationStatus;
};

export type CreateOfficeLocationRequest = {
  name: string;
  address: string | null;
  latitude: number;
  longitude: number;
  radius_meters: number;
};

export type UpdateOfficeLocationRequest = CreateOfficeLocationRequest;

export type UpdateOfficeLocationStatusRequest = {
  is_active: boolean;
};

export type LocationAssignment = {
  id: string;
  user: Employee;
  office_location: OfficeLocation;
  effective_from: string;
  effective_to: string | null;
  status: AssignmentStatus;
  created_at: string;
  updated_at: string;
};

export type LocationAssignmentListResponse = {
  items: LocationAssignment[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type LocationAssignmentListQuery = {
  page?: number;
  page_size?: number;
  search?: string;
  user_id?: string;
  office_location_id?: string;
  status?: AssignmentStatus;
};

export type CreateLocationAssignmentRequest = {
  user_id: string;
  office_location_id: string;
  effective_from: string;
  effective_to: string | null;
};

export type EndLocationAssignmentRequest = {
  effective_to: string;
};
