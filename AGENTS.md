# AGENTS.md

## Project overview

Project name: R3 TI FaceAttend

R3 TI FaceAttend is a local-first attendance prototype for employees of
the Information Technology Division at PTPN I Regional 3 Semarang.

The application consists of:

- Flutter mobile application for TI employees.
- Next.js admin website for administrators.
- Golang REST API as the only backend.
- PostgreSQL as the primary database.

The project currently runs locally and does not use a VPS.

## Current scope

The current scope is limited to:

- One division: Teknologi Informasi.
- One office: Kantor PTPN I Regional 3 Semarang.
- Two roles: USER and ADMIN.
- USER accesses the Flutter mobile application.
- ADMIN accesses the Next.js website.
- One primary work schedule.
- One work location.

Do not add multi-division, payroll, leave management, shift management,
or production deployment unless explicitly requested.

## Architecture rules

- Flutter and Next.js must never connect directly to PostgreSQL.
- All application data must pass through the Golang REST API.
- Use REST endpoints under `/api/v1`.
- Golang owns authentication, authorization, attendance rules,
  geofence validation, reporting, and audit logging.
- Next.js is only the admin frontend and session-facing BFF where needed.
- PostgreSQL must only be accessed by the Golang backend.
- Secrets must never be committed.
- Provide `.env.example` for every application that needs environment variables.
- Use database migrations for schema changes.
- Use server time as the authoritative attendance time.

## Development strategy

Work incrementally.

Current implementation order:

1. Environment verification.
2. Repository scaffolding.
3. Golang health endpoint and PostgreSQL connection.
4. Database migrations.
5. Admin authentication.
6. Employee management.
7. Flutter authentication.
8. Basic check-in and check-out without face recognition.
9. Admin attendance dashboard.
10. Geolocation.
11. Face enrollment and verification.
12. Basic liveness detection.
13. Attendance corrections and reports.

Do not implement face recognition before the basic attendance flow works.

## Coding conventions

- Use English for source code identifiers.
- Use Indonesian for user-facing text and project documentation.
- Keep functions small and focused.
- Avoid unnecessary abstractions.
- Prefer explicit code over premature generic solutions.
- Validate all request payloads.
- Return consistent JSON API responses.
- Handle errors explicitly.
- Do not expose internal errors, database errors, or stack traces to clients.
- Do not add a production dependency without explaining why it is needed.
- Avoid editing generated files unless required.

## Golang rules

- Use a modular structure under `internal/`.
- Separate HTTP handler, service, and repository responsibilities.
- Prefer PostgreSQL access through pgx.
- Use context for database and HTTP operations.
- Format code with `gofmt`.
- Run `go test ./...` after backend changes.
- Run `go vet ./...` when relevant.
- Do not use an ORM unless explicitly approved.

## Next.js rules

- Use TypeScript.
- Use App Router.
- Use a `src/` directory.
- Use Tailwind CSS.
- Keep server and client components clearly separated.
- Validate forms and API responses.
- Do not expose backend secrets through `NEXT_PUBLIC_` variables.
- Run lint and type checking after changes.

## Flutter rules

- Use feature-first organization.
- Separate presentation, data, and domain concerns where useful.
- Keep API base URLs configurable through `--dart-define`.
- Store tokens only in secure storage.
- Run `dart format`.
- Run `flutter analyze`.
- Run relevant Flutter tests after changes.
- Do not add face recognition packages until the face phase begins.

## Working agreement for Codex

Before editing:

1. Read this file.
2. Inspect the relevant existing files.
3. State the proposed plan.
4. Avoid unrelated changes.

After editing:

1. Format the code.
2. Run relevant linting and tests.
3. Review the resulting diff.
4. Report files changed.
5. Report commands executed.
6. Report tests that passed or failed.
7. Mention remaining risks or TODO items.

Never claim a command passed unless it was actually run successfully.
