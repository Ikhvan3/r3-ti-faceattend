# Duplicate Biometric Enrollment Protection

Dokumen ini menjelaskan pencegahan satu biometric wajah dipakai oleh lebih dari satu akun pegawai.

## Tujuan

Attendance tetap menggunakan verifikasi 1:1:

1. JWT menentukan user yang sedang login.
2. Backend mengambil satu face profile milik user tersebut.
3. Candidate embedding dibandingkan secara exact dengan template user tersebut.

Jumlah pegawai tidak membuat proses check-in/check-out berubah menjadi pencarian 1:N.

Pencarian 1:N hanya dilakukan pada saat enrollment untuk menjawab pertanyaan: apakah biometric candidate kemungkinan sudah digunakan akun lain?

## Arsitektur

Enrollment menggunakan dua tahap:

1. PostgreSQL `pgvector` + HNSW mencari kandidat terdekat dengan cepat.
2. Backend Go menghitung ulang cosine similarity secara exact terhadap kandidat Top-K.

Jika exact similarity mencapai `FACE_DUPLICATE_ENROLLMENT_THRESHOLD`, enrollment ditolak dengan HTTP 409 tanpa memberitahu identitas pemilik akun lain.

`FACE_VERIFICATION_THRESHOLD` dan `FACE_DUPLICATE_ENROLLMENT_THRESHOLD` sengaja dipisahkan karena 1:1 verification dan 1:N duplicate search mempunyai karakteristik false accept/false reject yang berbeda.

Threshold production wajib dikalibrasi dengan model, preprocessing, kamera, kondisi pencahayaan, dan populasi yang representatif. Jangan menganggap threshold development sebagai nilai produksi.

## Penyimpanan

`face_profiles.embedding` tetap dipertahankan sementara untuk exact cosine reranking di Go.

Migration `000008_add_face_vector_index` menambahkan:

- extension PostgreSQL `vector`;
- `embedding_vector vector(128)`;
- backfill dari embedding lama;
- HNSW index dengan `vector_cosine_ops`;
- index metadata model/status.

Raw image, crop wajah, base64 image, similarity score, dan threshold tidak disimpan sebagai bagian enrollment.

## Concurrency

Enrollment menggunakan transaction-scoped advisory lock. Enrollment relatif jarang dibanding attendance, sehingga serialisasi pendek pada proses enrollment dipilih untuk mencegah dua request concurrent dengan biometric yang sama sama-sama lolos duplicate check sebelum insert terlihat.

Lock ini tidak digunakan pada check-in/check-out.

## Instalasi pgvector pada PostgreSQL Windows

Migration `000008` membutuhkan extension `vector` tersedia pada instalasi PostgreSQL. Untuk PostgreSQL Windows, ikuti petunjuk resmi pgvector yang sesuai dengan versi PostgreSQL yang dipakai. Setelah pgvector terpasang, extension dibuat per-database oleh migration melalui:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

Verifikasi extension setelah migration:

```powershell
psql -h localhost -p 5432 -U postgres -d r3_ti_faceattend -c "SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';"
```

Verifikasi kolom/index:

```powershell
psql -h localhost -p 5432 -U postgres -d r3_ti_faceattend -c "\d face_profiles"
```

Harus terdapat `embedding_vector` dan index `face_profiles_embedding_hnsw_idx`.

## Upgrade environment lokal yang sudah ada

Jangan membuat ulang `.env` hanya untuk menambah konfigurasi ini karena membuat ulang JWT secret dapat menginvalidasi session development yang ada.

Dari root repository:

```powershell
.\scripts\upgrade-face-duplicate-config.ps1
```

Script menambahkan:

```env
FACE_DUPLICATE_ENROLLMENT_THRESHOLD=<nilai-development-yang-dievaluasi>
FACE_DUPLICATE_SEARCH_TOP_K=20
```

## Migration

Bangun `DatabaseUrl` sesuai konfigurasi lokal, lalu:

```powershell
cd backend
migrate -path migrations -database $DatabaseUrl up
```

Pastikan version migration telah mencapai `8` sebelum menjalankan backend baru.

## Existing duplicate data

Migration tidak menghapus atau memilih pemilik dari face profile yang sudah terlanjur duplikat. Keputusan kepemilikan biometric harus dilakukan secara eksplisit.

Untuk development test dengan dua akun yang saat ini memakai wajah sama:

1. tentukan satu akun yang menjadi enrollment canonical;
2. reset enrollment akun kedua;
3. jangan reset akun canonical;
4. pada akun kedua, coba enroll wajah canonical lagi;
5. backend harus menolak dengan pesan bahwa wajah telah terdaftar pada akun lain.

Jangan menghapus data biometric pegawai production secara otomatis berdasarkan nearest-neighbor result.

## Admin visibility

Admin hanya menerima metadata aman:

- enrolled / not enrolled;
- face status;
- enrolled_at;
- embedding model;
- embedding version.

Admin tidak menerima:

- embedding array/vector;
- raw image;
- face crop;
- similarity score;
- duplicate threshold;
- verification threshold;
- verification grant/token.

## Acceptance test

Minimum manual test:

```text
Akun A + wajah A -> enrollment berhasil
Akun B + wajah A -> enrollment ditolak
Akun B + wajah B -> enrollment berhasil

Akun A + wajah A -> attendance face verification berhasil
Akun A + wajah B -> ditolak
Akun B + wajah B -> berhasil
Akun B + wajah A -> ditolak
```

Setelah logout dari Akun A dan login Akun B yang belum memiliki face profile, `GET /face/status` harus mengembalikan `NOT_ENROLLED`; mobile juga harus membangun ulang state face controller untuk user baru.
