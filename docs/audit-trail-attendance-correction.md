# Audit Trail dan Koreksi Presensi

Tahap ini menambahkan audit trail append-only untuk tindakan Admin yang sensitif serta koreksi waktu presensi yang tetap menjaga integritas bukti lokasi.

## Prinsip

- Hanya role `ADMIN` yang dapat melakukan koreksi presensi dan reset enrollment wajah.
- Alasan wajib diisi minimal 5 karakter dan maksimum 1000 karakter.
- Perubahan data utama dan penulisan audit dilakukan dalam satu transaksi PostgreSQL.
- Audit log tidak memiliki endpoint update/delete.
- Password, JWT, refresh token, verification grant, face embedding, foto/crop wajah, similarity score, dan threshold biometrik tidak boleh disimpan di audit log.
- Koreksi waktu presensi tidak membuat atau memalsukan data GPS.

## Migration 000009

Migration `000009_create_audit_logs` membuat tabel `audit_logs` dengan metadata:

- actor user id, email, dan role;
- action;
- entity type dan entity id;
- target user dan label pegawai;
- reason;
- before_data dan after_data JSONB;
- created_at.

Jalankan:

```powershell
cd backend
migrate -path migrations -database $DatabaseUrl up
migrate -path migrations -database $DatabaseUrl version
```

Target version:

```text
9
```

## Action yang Dicatat

### ATTENDANCE_CORRECTED

Dibuat ketika Admin mengubah jam masuk dan/atau jam pulang melalui:

```http
PATCH /api/v1/admin/attendance/{attendance_id}/correction
```

Request:

```json
{
  "check_in_time": "08:05",
  "check_out_time": "17:10",
  "reason": "Koreksi berdasarkan konfirmasi kehadiran pegawai."
}
```

`check_out_time` boleh `null` jika pegawai memang belum check-out.

Backend menggabungkan nilai `HH:mm` dengan `attendance_date` menggunakan `BUSINESS_TIMEZONE`. `check_out_time` tidak boleh lebih awal daripada `check_in_time`.

Jika pegawai lupa check-out dan Admin menambahkan jam pulang secara manual, field bukti lokasi check-out tetap `NULL`. Sistem tidak membuat koordinat, distance, accuracy, atau office location palsu.

### FACE_ENROLLMENT_RESET

Reset enrollment wajah Admin sekarang membutuhkan body:

```json
{
  "reason": "Enrollment perlu diulang setelah verifikasi identitas pegawai."
}
```

Reset dan audit insert dilakukan dalam satu transaksi. Audit hanya menyimpan perubahan status `ENROLLED -> NOT_ENROLLED`; template embedding tidak pernah disalin ke audit log.

## Endpoint Audit Log

```http
GET /api/v1/admin/audit-logs
```

Filter opsional:

- `action=ATTENDANCE_CORRECTED|FACE_ENROLLMENT_RESET`
- `entity_type=ATTENDANCE_RECORD|FACE_PROFILE`
- `date_from=YYYY-MM-DD`
- `date_to=YYYY-MM-DD`
- `page`
- `page_size`, maksimum 100

Rentang tanggal maksimum 90 hari per request. Halaman Admin `/audit-logs` bersifat read-only.

## Acceptance Test

### Koreksi presensi

1. Pilih attendance record existing melalui Admin -> Presensi -> Detail.
2. Klik `Koreksi Presensi`.
3. Ubah jam masuk/jam pulang dan isi alasan.
4. Simpan.
5. Detail presensi harus langsung menampilkan waktu baru.
6. Status `is_late` harus dihitung ulang dari jam masuk baru, jadwal, dan grace period.
7. Buka `Audit Log`.
8. Harus terdapat `ATTENDANCE_CORRECTED` dengan actor Admin, target pegawai, reason, before, dan after.
9. Jika sebelumnya `check_out_at = null` lalu Admin mengisi jam pulang, `check_out_location` tetap kosong.

### Reset enrollment

1. Buka Admin -> Pegawai TI -> Detail pegawai dengan wajah terdaftar.
2. Klik `Reset Enrollment`.
3. Tombol konfirmasi harus disabled sampai alasan minimal 5 karakter.
4. Setelah berhasil, metadata wajah menjadi `NOT_ENROLLED`.
5. Buka `Audit Log`.
6. Harus terdapat `FACE_ENROLLMENT_RESET` dengan reason dan target pegawai.
7. Audit response tidak boleh memuat embedding, similarity, threshold, foto, atau token.

## Regression

Sebelum merge jalankan:

```powershell
cd backend
go fmt ./...
go test ./...
go vet ./...

cd ../mobile
dart format lib test
flutter analyze
flutter test

cd ../admin-web
npm run lint
npm run build
```

Physical regression minimum:

- login Admin tetap bekerja;
- daftar/detail Presensi tetap bekerja;
- mobile check-in/check-out tetap bekerja;
- riwayat mobile tetap ter-update;
- reset enrollment Admin masih bekerja dan sekarang membutuhkan alasan;
- duplicate biometric prevention tetap bekerja setelah enrollment ulang.
