# Admin Attendance Monitoring

Tahap ini menambahkan monitoring presensi read-only untuk ADMIN. Fitur ini tidak menyediakan koreksi, penghapusan, manual check-in, manual check-out, atau mutation attendance lain.

## Endpoint

Semua endpoint berada di bawah `/api/v1`, membutuhkan access token valid dengan role `ADMIN`, dan menggunakan `BUSINESS_TIMEZONE` sebagai acuan tanggal bisnis.

### Ringkasan

`GET /api/v1/admin/attendance/summary`

Query opsional:

- `date=YYYY-MM-DD`
- jika `date` tidak diberikan, backend menggunakan tanggal bisnis saat ini.

Response `data` memuat:

- `date`: tanggal bisnis.
- `active_employees`: jumlah pegawai USER ACTIVE yang memiliki assignment jadwal pada tanggal tersebut.
- `checked_in`: jumlah pegawai terjadwal yang sudah memiliki record check-in.
- `checked_out`: jumlah pegawai terjadwal yang sudah memiliki check-out.
- `not_checked_in`: jumlah pegawai terjadwal yang belum memiliki record attendance.
- `late`: jumlah attendance dengan waktu check-in setelah `start_time + grace_minutes` menurut business timezone.

Keterlambatan merupakan flag terpisah dari lifecycle attendance. Pegawai dapat memiliki `attendance_state = CHECKED_OUT` dan `is_late = true` sekaligus.

### Daftar presensi

`GET /api/v1/admin/attendance`

Query yang didukung:

- `date_from=YYYY-MM-DD`
- `date_to=YYYY-MM-DD`
- `employee_id=<uuid>`
- `search=<nama|nomor pegawai|email>`
- `attendance_state=NOT_CHECKED_IN|CHECKED_IN|CHECKED_OUT`
- `is_late=true|false`
- `page`, default 1
- `page_size`, default mengikuti pagination admin dan maksimum 100

Jika rentang tanggal tidak diberikan, daftar menggunakan tanggal bisnis hari ini. Jika rentang diberikan, `date_from` dan `date_to` harus keduanya tersedia, `date_from <= date_to`, dan panjang rentang maksimum 31 hari.

Daftar juga dapat memuat pegawai terjadwal yang belum check-in. Record tersebut memiliki `id = null`, `check_in_at = null`, dan `attendance_state = NOT_CHECKED_IN`, sehingga tidak memiliki halaman detail attendance.

### Detail presensi

`GET /api/v1/admin/attendance/{id}`

Detail hanya tersedia untuk attendance record yang benar-benar sudah dibuat. Response memuat:

- profil pegawai aman;
- tanggal bisnis;
- jadwal dan grace period;
- waktu check-in/check-out;
- `attendance_state`;
- `is_late`;
- bukti lokasi check-in/check-out jika tersedia: nama office, radius, latitude, longitude, GPS accuracy, dan distance.

## Data yang sengaja tidak diekspos

Endpoint monitoring tidak mengembalikan:

- raw face image atau crop;
- face embedding/template;
- candidate embedding;
- similarity score atau verification threshold;
- face verification grant atau grant hash;
- JWT/access token/refresh token;
- password hash.

## State presensi

Lifecycle state dipisahkan dari keterlambatan:

- `NOT_CHECKED_IN`: pegawai memiliki assignment jadwal untuk tanggal bisnis tetapi belum memiliki attendance record.
- `CHECKED_IN`: attendance record ada dan `check_out_at` masih null.
- `CHECKED_OUT`: attendance record ada dan `check_out_at` sudah terisi.

`is_late` dihitung backend dengan membandingkan waktu check-in pada `BUSINESS_TIMEZONE` terhadap waktu mulai jadwal ditambah `grace_minutes`.

## Admin web

Navigasi admin menambahkan menu `Presensi` dengan route `/attendance`.

Dashboard menampilkan ringkasan hari ini dan beberapa baris presensi terbaru. Halaman `/attendance` menyediakan filter tanggal, pencarian pegawai, lifecycle state, keterlambatan, dan pagination. Detail record tersedia pada `/attendance/{id}` dan menampilkan bukti GPS secara read-only.

## Batasan tahap ini

- Monitoring bersifat read-only.
- Tidak ada koreksi attendance atau approval.
- Tidak ada export CSV/XLSX pada tahap ini.
- Tidak ada peta eksternal.
- Tidak ada audit mutation karena belum ada mutation attendance admin.
- Range query dibatasi agar prototype tidak melakukan scan laporan yang tidak terkontrol.

Koreksi attendance, audit trail, dan reporting/export dikerjakan pada tahap terpisah setelah monitoring ini stabil.
