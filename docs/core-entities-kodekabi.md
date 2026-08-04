# Core Entity KODEKABI: Jejak Algoritma

Dokumen ini mendefinisikan core entity yang dibutuhkan untuk mengimplementasikan MVP KODEKABI: Jejak Algoritma.

## Konvensi Database

- Seluruh primary key dan foreign key menggunakan `varchar(36)` untuk menyimpan UUID.
- Nama tabel menggunakan bentuk plural dan `snake_case`.
- Nama foreign key menggunakan pola `<entity>_id`.
- Kolom waktu menggunakan `created_at`, `updated_at`, dan `deleted_at`.
- Soft delete digunakan pada master data yang masih mungkin direferensikan oleh histori.
- Nilai status disimpan sebagai `varchar` dan divalidasi pada backend atau melalui database constraint.

---

# Daftar Core Entity

1. `roles`
2. `users`
3. `user_profiles`
4. `avatars`
5. `cities`
6. `city_statistics`
7. `cases`
8. `case_versions`
9. `evidences`
10. `questions`
11. `question_options`
12. `case_sessions`
13. `session_evidence_logs`
14. `answers`
15. `score_results`
16. `city_impact_logs`

---

# 1. Roles

Entity `roles` menyimpan jenis peran pengguna yang digunakan untuk Role-Based Access Control.

Contoh role:

- `player`
- `content_author`
- `admin`

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID role |
| `name` | `varchar(100)` | NOT NULL, UNIQUE | Nama role untuk tampilan |
| `code` | `varchar(50)` | NOT NULL, UNIQUE | Kode role untuk otorisasi |
| `description` | `text` | NULL | Penjelasan fungsi role |
| `status` | `varchar(20)` | NOT NULL | `active` atau `inactive` |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |
| `deleted_at` | `timestamp` | NULL | Waktu soft delete |

## Relasi

- Satu `role` dapat dimiliki banyak `users`.
- Satu `user` memiliki satu role utama.

---

# 2. Users

Entity `users` menyimpan identitas akun, kredensial autentikasi, dan status akun.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID user |
| `role_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `roles.id` |
| `username` | `varchar(50)` | NOT NULL, UNIQUE | Identitas publik pengguna |
| `email` | `varchar(150)` | NOT NULL, UNIQUE | Email untuk autentikasi |
| `password_hash` | `varchar(255)` | NOT NULL | Password yang sudah di-hash |
| `account_status` | `varchar(20)` | NOT NULL | `pending`, `active`, `suspended`, atau `banned` |
| `email_verified_at` | `timestamp` | NULL | Waktu verifikasi email |
| `last_login_at` | `timestamp` | NULL | Waktu login terakhir |
| `created_at` | `timestamp` | NOT NULL | Waktu registrasi |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |
| `deleted_at` | `timestamp` | NULL | Waktu soft delete |

## Relasi

- Banyak `users` dimiliki oleh satu `role`.
- Satu `user` memiliki satu `user_profile`.
- Satu `user` dapat membuat banyak `cases`.
- Satu `user` dapat menjalankan banyak `case_sessions`.

---

# 3. User Profiles

Entity `user_profiles` menyimpan profil gameplay dan progres terbaru pemain.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID profile |
| `user_id` | `varchar(36)` | FK, NOT NULL, UNIQUE | Referensi ke `users.id` |
| `avatar_id` | `varchar(36)` | FK, NULL | Referensi ke `avatars.id` |
| `age_range` | `varchar(20)` | NULL | Contoh `15-17`, `18-21`, atau `22-24` |
| `preferred_language` | `varchar(10)` | NOT NULL | Bahasa utama, default `id` |
| `consent_status` | `boolean` | NOT NULL | Status persetujuan kebijakan data |
| `consented_at` | `timestamp` | NULL | Waktu persetujuan diberikan |
| `current_level` | `int` | NOT NULL | Level pemain saat ini |
| `current_xp` | `bigint` | NOT NULL | Total XP pemain |
| `auditor_reputation` | `decimal(8,2)` | NOT NULL | Reputasi Auditor Digital |
| `evidence_evaluation_score` | `decimal(5,2)` | NOT NULL | Kompetensi evaluasi bukti |
| `claim_analysis_score` | `decimal(5,2)` | NOT NULL | Kompetensi analisis klaim |
| `confidence_calibration_score` | `decimal(5,2)` | NOT NULL | Kompetensi kalibrasi keyakinan |
| `reasoning_score` | `decimal(5,2)` | NOT NULL | Kompetensi penalaran |
| `safety_judgment_score` | `decimal(5,2)` | NOT NULL | Kompetensi penilaian risiko |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Relasi

- Satu `user_profile` dimiliki oleh satu `user`.
- Banyak `user_profiles` dapat menggunakan satu `avatar`.

---

# 4. Avatars

Entity `avatars` menyimpan master data avatar yang dapat dipilih atau dibuka pemain.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID avatar |
| `name` | `varchar(100)` | NOT NULL | Nama avatar |
| `image_url` | `varchar(500)` | NOT NULL | URL aset utama |
| `thumbnail_url` | `varchar(500)` | NULL | URL thumbnail |
| `unlock_level` | `int` | NOT NULL | Level minimum untuk memakai avatar |
| `is_default` | `boolean` | NOT NULL | Tersedia sejak onboarding |
| `status` | `varchar(20)` | NOT NULL | `active` atau `inactive` |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |
| `deleted_at` | `timestamp` | NULL | Waktu soft delete |

## Relasi

- Satu `avatar` dapat digunakan oleh banyak `user_profiles`.

---

# 5. Cities

Entity `cities` menyimpan identitas kota virtual. Pada MVP, data utama dapat berupa Kota Nusa.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID kota |
| `name` | `varchar(100)` | NOT NULL | Nama kota |
| `slug` | `varchar(120)` | NOT NULL, UNIQUE | Identitas URL |
| `description` | `text` | NULL | Deskripsi kota |
| `background_url` | `varchar(500)` | NULL | URL visual kota |
| `status` | `varchar(20)` | NOT NULL | Status kota |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |
| `deleted_at` | `timestamp` | NULL | Waktu soft delete |

## Relasi

- Satu `city` memiliki satu `city_statistics`.
- Satu `city` dapat memiliki banyak `cases`.
- Satu `city` dapat menerima banyak `city_impact_logs`.

---

# 6. City Statistics

Entity `city_statistics` menyimpan kondisi terbaru kota.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID statistik |
| `city_id` | `varchar(36)` | FK, NOT NULL, UNIQUE | Referensi ke `cities.id` |
| `information_health` | `decimal(5,2)` | NOT NULL | Nilai 0 sampai 100 |
| `public_trust` | `decimal(5,2)` | NOT NULL | Nilai 0 sampai 100 |
| `social_stability` | `decimal(5,2)` | NOT NULL | Nilai 0 sampai 100 |
| `public_wellbeing` | `decimal(5,2)` | NOT NULL | Nilai 0 sampai 100 |
| `state_version` | `bigint` | NOT NULL | Versi state untuk concurrency control |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Relasi

- Satu `city_statistics` dimiliki oleh satu `city`.

---

# 7. Cases

Entity `cases` menyimpan identitas dan metadata utama case investigasi.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID case |
| `city_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `cities.id` |
| `created_by` | `varchar(36)` | FK, NOT NULL | Referensi ke `users.id` |
| `title` | `varchar(200)` | NOT NULL | Judul case |
| `slug` | `varchar(220)` | NOT NULL, UNIQUE | Identitas URL |
| `short_description` | `text` | NOT NULL | Ringkasan case |
| `thumbnail_url` | `varchar(500)` | NULL | URL thumbnail |
| `source_type` | `varchar(50)` | NOT NULL | Jenis sumber utama |
| `difficulty_level` | `varchar(20)` | NOT NULL | Tingkat kesulitan |
| `risk_level` | `varchar(20)` | NOT NULL | Tingkat risiko |
| `estimated_duration_minutes` | `int` | NOT NULL | Estimasi waktu pengerjaan |
| `minimum_level` | `int` | NOT NULL | Level minimum |
| `minimum_reputation` | `decimal(8,2)` | NOT NULL | Reputasi minimum |
| `status` | `varchar(20)` | NOT NULL | `draft`, `published`, atau `archived` |
| `published_at` | `timestamp` | NULL | Waktu publikasi |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |
| `deleted_at` | `timestamp` | NULL | Waktu soft delete |

## Nilai `source_type`

- `social_post`
- `article`
- `blog`
- `forum_thread`
- `chatbot_conversation`
- `public_announcement`

## Relasi

- Banyak `cases` dimiliki oleh satu `city`.
- Banyak `cases` dapat dibuat oleh satu `user`.
- Satu `case` memiliki banyak `case_versions`.
- Satu `case` dapat dimainkan dalam banyak `case_sessions`.

---

# 8. Case Versions

Entity `case_versions` menyimpan versi konten case agar perubahan konten tidak memengaruhi sesi yang sedang berjalan.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID versi case |
| `case_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `cases.id` |
| `version_number` | `int` | NOT NULL | Nomor versi |
| `briefing_title` | `varchar(200)` | NOT NULL | Judul briefing |
| `briefing_content` | `text` | NOT NULL | Isi briefing |
| `learning_focus` | `varchar(255)` | NULL | Fokus kompetensi |
| `chatbot_persona` | `text` | NULL | Persona chatbot |
| `chatbot_knowledge_boundary` | `text` | NULL | Batas pengetahuan chatbot |
| `chatbot_prohibited_behavior` | `text` | NULL | Perilaku yang dilarang |
| `scoring_config` | `json` | NOT NULL | Konfigurasi scoring |
| `outcome_config` | `json` | NOT NULL | Konfigurasi outcome dan dampak |
| `status` | `varchar(20)` | NOT NULL | `draft`, `active`, atau `retired` |
| `created_by` | `varchar(36)` | FK, NOT NULL | Referensi ke pembuat versi |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `published_at` | `timestamp` | NULL | Waktu versi dipublikasikan |

## Constraint Tambahan

```sql
UNIQUE (case_id, version_number)
```

## Relasi

- Banyak `case_versions` dimiliki oleh satu `case`.
- Satu `case_version` memiliki banyak `evidences`.
- Satu `case_version` memiliki banyak `questions`.
- Satu `case_version` digunakan oleh banyak `case_sessions`.

---

# 9. Evidences

Entity `evidences` menyimpan potongan informasi yang dapat diperiksa pemain.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID evidence |
| `case_version_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `case_versions.id` |
| `title` | `varchar(200)` | NOT NULL | Judul evidence |
| `evidence_type` | `varchar(50)` | NOT NULL | Jenis evidence |
| `source_name` | `varchar(150)` | NULL | Nama sumber simulatif |
| `source_profile` | `json` | NULL | Profil sumber |
| `content` | `longtext` | NOT NULL | Isi evidence |
| `media_url` | `varchar(500)` | NULL | URL media |
| `metadata` | `json` | NULL | Metadata tambahan |
| `display_order` | `int` | NOT NULL | Urutan default |
| `is_required` | `boolean` | NOT NULL | Wajib dibuka atau tidak |
| `reliability_weight` | `decimal(5,2)` | NULL | Bobot kualitas evidence |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Nilai `evidence_type`

- `social_post`
- `article`
- `chat_transcript`
- `statistic`
- `source_profile`
- `forum_comment`
- `public_statement`

## Relasi

- Banyak `evidences` dimiliki oleh satu `case_version`.
- Satu `evidence` dapat dibuka dalam banyak `session_evidence_logs`.

---

# 10. Questions

Entity `questions` menyimpan pertanyaan yang harus dijawab pemain selama investigasi.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID pertanyaan |
| `case_version_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `case_versions.id` |
| `question_text` | `text` | NOT NULL | Isi pertanyaan |
| `question_type` | `varchar(50)` | NOT NULL | Jenis pertanyaan |
| `question_stage` | `varchar(30)` | NOT NULL | Tahap pertanyaan |
| `competency_dimension` | `varchar(50)` | NOT NULL | Kompetensi yang diukur |
| `instruction_text` | `text` | NULL | Instruksi tambahan |
| `is_required` | `boolean` | NOT NULL | Pertanyaan wajib |
| `min_value` | `decimal(8,2)` | NULL | Nilai minimum |
| `max_value` | `decimal(8,2)` | NULL | Nilai maksimum |
| `max_length` | `int` | NULL | Maksimum panjang jawaban |
| `display_order` | `int` | NOT NULL | Urutan tampilan |
| `scoring_rule` | `json` | NULL | Rule scoring pertanyaan |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Nilai `question_type`

- `single_choice`
- `multiple_choice`
- `confidence_input`
- `claim_classification`
- `open_question`
- `final_decision`

## Nilai `question_stage`

- `initial`
- `investigation`
- `final`

## Relasi

- Banyak `questions` dimiliki oleh satu `case_version`.
- Satu `question` dapat memiliki banyak `question_options`.
- Satu `question` dapat dijawab dalam banyak `answers`.

---

# 11. Question Options

Entity `question_options` menyimpan pilihan jawaban dari pertanyaan terstruktur.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID option |
| `question_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `questions.id` |
| `option_code` | `varchar(50)` | NOT NULL | Kode pilihan |
| `option_text` | `text` | NOT NULL | Teks pilihan |
| `option_value` | `varchar(100)` | NULL | Nilai internal |
| `score_weight` | `decimal(8,2)` | NULL | Bobot scoring |
| `feedback_text` | `text` | NULL | Feedback khusus |
| `display_order` | `int` | NOT NULL | Urutan pilihan |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Relasi

- Banyak `question_options` dimiliki oleh satu `question`.
- Satu `question_option` dapat dipilih pada banyak `answers`.

---

# 12. Case Sessions

Entity `case_sessions` merepresentasikan satu sesi pemain dalam menyelesaikan sebuah case.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID sesi |
| `user_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `users.id` |
| `case_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `cases.id` |
| `case_version_id` | `varchar(36)` | FK, NOT NULL | Snapshot versi yang digunakan |
| `status` | `varchar(30)` | NOT NULL | Status sesi |
| `session_version` | `bigint` | NOT NULL | Optimistic locking version |
| `started_at` | `timestamp` | NOT NULL | Waktu mulai |
| `last_activity_at` | `timestamp` | NULL | Aktivitas terakhir |
| `submitted_at` | `timestamp` | NULL | Waktu final submit |
| `completed_at` | `timestamp` | NULL | Waktu proses selesai |
| `initial_confidence` | `decimal(5,2)` | NULL | Confidence awal |
| `final_confidence` | `decimal(5,2)` | NULL | Confidence akhir |
| `final_decision` | `varchar(100)` | NULL | Keputusan akhir |
| `final_reasoning` | `text` | NULL | Alasan keputusan |
| `idempotency_key` | `varchar(100)` | NULL, UNIQUE | Pencegah submit ganda |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Nilai `status`

- `active`
- `submitted`
- `scoring`
- `completed`
- `abandoned`
- `expired`

## Relasi

- Banyak `case_sessions` dimiliki oleh satu `user`.
- Banyak `case_sessions` merujuk ke satu `case`.
- Banyak `case_sessions` menggunakan satu `case_version`.
- Satu `case_session` memiliki banyak `answers`.
- Satu `case_session` memiliki banyak `session_evidence_logs`.
- Satu `case_session` menghasilkan satu `score_result`.

---

# 13. Session Evidence Logs

Entity `session_evidence_logs` mencatat evidence yang dibuka pemain beserta pola interaksinya.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID log |
| `session_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `case_sessions.id` |
| `evidence_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `evidences.id` |
| `open_sequence` | `int` | NOT NULL | Urutan pembukaan evidence |
| `first_opened_at` | `timestamp` | NOT NULL | Pertama kali dibuka |
| `last_opened_at` | `timestamp` | NOT NULL | Terakhir kali dibuka |
| `open_count` | `int` | NOT NULL | Jumlah pembukaan |
| `view_duration_seconds` | `int` | NOT NULL | Estimasi durasi dilihat |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Constraint Tambahan

```sql
UNIQUE (session_id, evidence_id)
```

## Relasi

- Banyak `session_evidence_logs` dimiliki oleh satu `case_session`.
- Banyak `session_evidence_logs` merujuk ke satu `evidence`.

---

# 14. Answers

Entity `answers` menyimpan jawaban draft dan final pemain.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID jawaban |
| `session_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `case_sessions.id` |
| `question_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `questions.id` |
| `selected_option_id` | `varchar(36)` | FK, NULL | Pilihan tunggal |
| `answer_text` | `longtext` | NULL | Jawaban teks |
| `numeric_value` | `decimal(8,2)` | NULL | Confidence atau nilai numerik |
| `answer_payload` | `json` | NULL | Multiple choice atau data kompleks |
| `is_final` | `boolean` | NOT NULL | Draft atau final |
| `answered_at` | `timestamp` | NULL | Waktu pertama menjawab |
| `finalized_at` | `timestamp` | NULL | Waktu jawaban difinalisasi |
| `idempotency_key` | `varchar(100)` | NULL | Pencegah duplikasi mutasi |
| `created_at` | `timestamp` | NOT NULL | Waktu data dibuat |
| `updated_at` | `timestamp` | NOT NULL | Waktu terakhir diperbarui |

## Constraint Tambahan

```sql
UNIQUE (session_id, question_id)
```

## Relasi

- Banyak `answers` dimiliki oleh satu `case_session`.
- Banyak `answers` merujuk ke satu `question`.
- Banyak `answers` dapat memilih satu `question_option`.

---

# 15. Score Results

Entity `score_results` menyimpan hasil scoring final dari satu sesi.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID hasil scoring |
| `session_id` | `varchar(36)` | FK, NOT NULL, UNIQUE | Referensi ke `case_sessions.id` |
| `evidence_evaluation_score` | `decimal(5,2)` | NOT NULL | Skor evaluasi evidence |
| `claim_analysis_score` | `decimal(5,2)` | NOT NULL | Skor analisis klaim |
| `confidence_calibration_score` | `decimal(5,2)` | NOT NULL | Skor kalibrasi confidence |
| `reasoning_score` | `decimal(5,2)` | NOT NULL | Skor penalaran |
| `safety_judgment_score` | `decimal(5,2)` | NOT NULL | Skor keputusan aman |
| `total_score` | `decimal(6,2)` | NOT NULL | Total skor |
| `semantic_signals` | `json` | NULL | Sinyal hasil analisis AI |
| `scoring_breakdown` | `json` | NOT NULL | Penjelasan rule scoring |
| `outcome_code` | `varchar(100)` | NOT NULL | Kode outcome |
| `outcome_title` | `varchar(200)` | NOT NULL | Judul outcome |
| `outcome_description` | `text` | NOT NULL | Narasi konsekuensi |
| `strength_feedback` | `text` | NULL | Kekuatan utama |
| `improvement_feedback` | `text` | NULL | Area perbaikan |
| `xp_earned` | `int` | NOT NULL | XP yang diperoleh |
| `reputation_change` | `decimal(8,2)` | NOT NULL | Perubahan reputasi |
| `scoring_rule_version` | `varchar(50)` | NOT NULL | Versi scoring rule |
| `created_at` | `timestamp` | NOT NULL | Waktu scoring selesai |

## Relasi

- Satu `score_result` dihasilkan oleh satu `case_session`.
- Satu `score_result` menghasilkan satu `city_impact_log`.

---

# 16. City Impact Logs

Entity `city_impact_logs` menyimpan perubahan kondisi kota akibat hasil sebuah sesi.

## Atribut

| Atribut | Tipe Data | Constraint | Keterangan |
|---|---|---|---|
| `id` | `varchar(36)` | PK | UUID impact |
| `city_id` | `varchar(36)` | FK, NOT NULL | Referensi ke `cities.id` |
| `session_id` | `varchar(36)` | FK, NOT NULL, UNIQUE | Sesi penyebab dampak |
| `score_result_id` | `varchar(36)` | FK, NOT NULL, UNIQUE | Referensi hasil scoring |
| `information_health_before` | `decimal(5,2)` | NOT NULL | Nilai sebelum |
| `information_health_change` | `decimal(5,2)` | NOT NULL | Perubahan nilai |
| `information_health_after` | `decimal(5,2)` | NOT NULL | Nilai sesudah |
| `public_trust_before` | `decimal(5,2)` | NOT NULL | Nilai sebelum |
| `public_trust_change` | `decimal(5,2)` | NOT NULL | Perubahan nilai |
| `public_trust_after` | `decimal(5,2)` | NOT NULL | Nilai sesudah |
| `social_stability_before` | `decimal(5,2)` | NOT NULL | Nilai sebelum |
| `social_stability_change` | `decimal(5,2)` | NOT NULL | Perubahan nilai |
| `social_stability_after` | `decimal(5,2)` | NOT NULL | Nilai sesudah |
| `public_wellbeing_before` | `decimal(5,2)` | NOT NULL | Nilai sebelum |
| `public_wellbeing_change` | `decimal(5,2)` | NOT NULL | Perubahan nilai |
| `public_wellbeing_after` | `decimal(5,2)` | NOT NULL | Nilai sesudah |
| `impact_summary` | `text` | NULL | Ringkasan dampak |
| `created_at` | `timestamp` | NOT NULL | Waktu perubahan terjadi |

## Relasi

- Banyak `city_impact_logs` dimiliki oleh satu `city`.
- Satu `city_impact_log` berasal dari satu `case_session`.
- Satu `city_impact_log` berasal dari satu `score_result`.

---

# Mermaid ERD Lengkap

```mermaid
erDiagram
    ROLES {
        varchar(36) id PK
        varchar(100) name UK
        varchar(50) code UK
        text description
        varchar(20) status
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    USERS {
        varchar(36) id PK
        varchar(36) role_id FK
        varchar(50) username UK
        varchar(150) email UK
        varchar(255) password_hash
        varchar(20) account_status
        timestamp email_verified_at
        timestamp last_login_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    USER_PROFILES {
        varchar(36) id PK
        varchar(36) user_id FK,UK
        varchar(36) avatar_id FK
        varchar(20) age_range
        varchar(10) preferred_language
        boolean consent_status
        timestamp consented_at
        int current_level
        bigint current_xp
        decimal(8,2) auditor_reputation
        decimal(5,2) evidence_evaluation_score
        decimal(5,2) claim_analysis_score
        decimal(5,2) confidence_calibration_score
        decimal(5,2) reasoning_score
        decimal(5,2) safety_judgment_score
        timestamp created_at
        timestamp updated_at
    }

    AVATARS {
        varchar(36) id PK
        varchar(100) name
        varchar(500) image_url
        varchar(500) thumbnail_url
        int unlock_level
        boolean is_default
        varchar(20) status
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    CITIES {
        varchar(36) id PK
        varchar(100) name
        varchar(120) slug UK
        text description
        varchar(500) background_url
        varchar(20) status
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    CITY_STATISTICS {
        varchar(36) id PK
        varchar(36) city_id FK,UK
        decimal(5,2) information_health
        decimal(5,2) public_trust
        decimal(5,2) social_stability
        decimal(5,2) public_wellbeing
        bigint state_version
        timestamp updated_at
    }

    CASES {
        varchar(36) id PK
        varchar(36) city_id FK
        varchar(36) created_by FK
        varchar(200) title
        varchar(220) slug UK
        text short_description
        varchar(500) thumbnail_url
        varchar(50) source_type
        varchar(20) difficulty_level
        varchar(20) risk_level
        int estimated_duration_minutes
        int minimum_level
        decimal(8,2) minimum_reputation
        varchar(20) status
        timestamp published_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    CASE_VERSIONS {
        varchar(36) id PK
        varchar(36) case_id FK
        int version_number
        varchar(200) briefing_title
        text briefing_content
        varchar(255) learning_focus
        text chatbot_persona
        text chatbot_knowledge_boundary
        text chatbot_prohibited_behavior
        json scoring_config
        json outcome_config
        varchar(20) status
        varchar(36) created_by FK
        timestamp created_at
        timestamp published_at
    }

    EVIDENCES {
        varchar(36) id PK
        varchar(36) case_version_id FK
        varchar(200) title
        varchar(50) evidence_type
        varchar(150) source_name
        json source_profile
        longtext content
        varchar(500) media_url
        json metadata
        int display_order
        boolean is_required
        decimal(5,2) reliability_weight
        timestamp created_at
        timestamp updated_at
    }

    QUESTIONS {
        varchar(36) id PK
        varchar(36) case_version_id FK
        text question_text
        varchar(50) question_type
        varchar(30) question_stage
        varchar(50) competency_dimension
        text instruction_text
        boolean is_required
        decimal(8,2) min_value
        decimal(8,2) max_value
        int max_length
        int display_order
        json scoring_rule
        timestamp created_at
        timestamp updated_at
    }

    QUESTION_OPTIONS {
        varchar(36) id PK
        varchar(36) question_id FK
        varchar(50) option_code
        text option_text
        varchar(100) option_value
        decimal(8,2) score_weight
        text feedback_text
        int display_order
        timestamp created_at
        timestamp updated_at
    }

    CASE_SESSIONS {
        varchar(36) id PK
        varchar(36) user_id FK
        varchar(36) case_id FK
        varchar(36) case_version_id FK
        varchar(30) status
        bigint session_version
        timestamp started_at
        timestamp last_activity_at
        timestamp submitted_at
        timestamp completed_at
        decimal(5,2) initial_confidence
        decimal(5,2) final_confidence
        varchar(100) final_decision
        text final_reasoning
        varchar(100) idempotency_key UK
        timestamp created_at
        timestamp updated_at
    }

    SESSION_EVIDENCE_LOGS {
        varchar(36) id PK
        varchar(36) session_id FK
        varchar(36) evidence_id FK
        int open_sequence
        timestamp first_opened_at
        timestamp last_opened_at
        int open_count
        int view_duration_seconds
        timestamp created_at
        timestamp updated_at
    }

    ANSWERS {
        varchar(36) id PK
        varchar(36) session_id FK
        varchar(36) question_id FK
        varchar(36) selected_option_id FK
        longtext answer_text
        decimal(8,2) numeric_value
        json answer_payload
        boolean is_final
        timestamp answered_at
        timestamp finalized_at
        varchar(100) idempotency_key
        timestamp created_at
        timestamp updated_at
    }

    SCORE_RESULTS {
        varchar(36) id PK
        varchar(36) session_id FK,UK
        decimal(5,2) evidence_evaluation_score
        decimal(5,2) claim_analysis_score
        decimal(5,2) confidence_calibration_score
        decimal(5,2) reasoning_score
        decimal(5,2) safety_judgment_score
        decimal(6,2) total_score
        json semantic_signals
        json scoring_breakdown
        varchar(100) outcome_code
        varchar(200) outcome_title
        text outcome_description
        text strength_feedback
        text improvement_feedback
        int xp_earned
        decimal(8,2) reputation_change
        varchar(50) scoring_rule_version
        timestamp created_at
    }

    CITY_IMPACT_LOGS {
        varchar(36) id PK
        varchar(36) city_id FK
        varchar(36) session_id FK,UK
        varchar(36) score_result_id FK,UK
        decimal(5,2) information_health_before
        decimal(5,2) information_health_change
        decimal(5,2) information_health_after
        decimal(5,2) public_trust_before
        decimal(5,2) public_trust_change
        decimal(5,2) public_trust_after
        decimal(5,2) social_stability_before
        decimal(5,2) social_stability_change
        decimal(5,2) social_stability_after
        decimal(5,2) public_wellbeing_before
        decimal(5,2) public_wellbeing_change
        decimal(5,2) public_wellbeing_after
        text impact_summary
        timestamp created_at
    }

    ROLES ||--o{ USERS : has
    USERS ||--|| USER_PROFILES : owns
    AVATARS ||--o{ USER_PROFILES : selected_by

    CITIES ||--|| CITY_STATISTICS : has
    CITIES ||--o{ CASES : contains
    CITIES ||--o{ CITY_IMPACT_LOGS : receives

    USERS ||--o{ CASES : creates
    USERS ||--o{ CASE_VERSIONS : creates
    USERS ||--o{ CASE_SESSIONS : plays

    CASES ||--o{ CASE_VERSIONS : has
    CASES ||--o{ CASE_SESSIONS : investigated_in

    CASE_VERSIONS ||--o{ EVIDENCES : contains
    CASE_VERSIONS ||--o{ QUESTIONS : contains
    CASE_VERSIONS ||--o{ CASE_SESSIONS : snapshotted_as

    QUESTIONS ||--o{ QUESTION_OPTIONS : provides
    QUESTIONS ||--o{ ANSWERS : answered_in
    QUESTION_OPTIONS ||--o{ ANSWERS : selected_in

    CASE_SESSIONS ||--o{ SESSION_EVIDENCE_LOGS : records
    EVIDENCES ||--o{ SESSION_EVIDENCE_LOGS : opened_in

    CASE_SESSIONS ||--o{ ANSWERS : contains
    CASE_SESSIONS ||--|| SCORE_RESULTS : produces
    CASE_SESSIONS ||--|| CITY_IMPACT_LOGS : causes

    SCORE_RESULTS ||--|| CITY_IMPACT_LOGS : generates
```

---

# Urutan Pembuatan Migration

Urutan migration yang disarankan agar dependency foreign key tidak bertabrakan:

1. `roles`
2. `avatars`
3. `users`
4. `user_profiles`
5. `cities`
6. `city_statistics`
7. `cases`
8. `case_versions`
9. `evidences`
10. `questions`
11. `question_options`
12. `case_sessions`
13. `session_evidence_logs`
14. `answers`
15. `score_results`
16. `city_impact_logs`

---

# Entity Lanjutan

Entity berikut belum dimasukkan ke core awal, tetapi dapat ditambahkan setelah core gameplay stabil:

- `chat_conversations`
- `chat_messages`
- `case_prerequisites`
- `user_competency_histories`
- `user_progress_logs`
- `level_configs`
- `leaderboard_records`
- `achievements`
- `user_achievements`
- `analytics_events`
- `refresh_tokens`
- `case_reviews`
