import type { AccountStatus, UserRole } from "@/lib/auth/types";

export type Employee = {
  id: string;
  employee_number: string;
  name: string;
  email: string;
  phone: string | null;
  position: string | null;
  role: UserRole;
  account_status: AccountStatus;
  created_at: string;
  updated_at: string;
};

export type EmployeeListItem = Employee;

export type EmployeePagination = {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export type EmployeeListResponse = EmployeePagination & {
  items: EmployeeListItem[];
};

export type EmployeeListQuery = {
  page?: number;
  page_size?: number;
  search?: string;
  status?: AccountStatus;
};

export type CreateEmployeeRequest = {
  employee_number: string;
  name: string;
  email: string;
  initial_password: string;
  phone: string | null;
  position: string | null;
};

export type UpdateEmployeeRequest = {
  employee_number: string;
  name: string;
  email: string;
  phone: string | null;
  position: string | null;
};

export type UpdateEmployeeStatusRequest = {
  account_status: AccountStatus;
};

export type ApiSuccessResponse<T> = {
  status: "ok";
  message: string;
  data: T;
};

export type ApiErrorResponse = {
  status: "error";
  message: string;
};
