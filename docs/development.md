# Development

Dokumen ini mencatat cara kerja lokal untuk repository R3 TI FaceAttend.

## Prasyarat

- Git.
- Flutter dan Android toolchain.
- Node.js dan npm.
- Golang.
- PostgreSQL lokal.

## Mobile

```powershell
cd mobile
flutter pub get
dart format .
flutter analyze
flutter test
```

Konfigurasi awal dicatat di `mobile/.env.example`. Pada tahap berikutnya nilai
runtime dapat diberikan melalui `--dart-define`.

## Admin Web

```powershell
cd admin-web
npm install
npm run lint
npx tsc --noEmit
npm run dev
```

Konfigurasi awal dicatat di `admin-web/.env.example`.

## Backend

```powershell
cd backend
go test ./...
go vet ./...
go run ./cmd/api
```

Konfigurasi awal dicatat di `backend/.env.example`.
