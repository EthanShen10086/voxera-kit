export interface FixtureUser {
  id: string;
  email: string;
  name: string;
  role: "admin" | "member" | "guest";
}

export interface FixtureSession {
  token: string;
  userId: string;
  expiresAt: string;
}

export interface ApiListResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ApiErrorResponse {
  code: string;
  message: string;
}
