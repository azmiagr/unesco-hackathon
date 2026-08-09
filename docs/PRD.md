# Product Requirements Document (PRD)
## KODEKABI: Jejak Algoritma

**Game Simulasi Investigasi Digital untuk Penguatan Agensi Epistemik Generasi Muda di Tengah Ekosistem AI dan Informasi Digital**

> *"Bongkar informasinya. Baca polanya. Selamatkan kotanya."*

| | |
|---|---|
| **Konteks** | UNESCO Youth Hackathon |
| **Versi Dokumen** | 2.0 (Konsolidasi) |
| **Status** | Draft untuk Review |
| **Tanggal** | 24 Juli 2026 (revisi konsolidasi terakhir mengikuti pembahasan Shop & Redeem serta Admin Console) |

> **Catatan Konsolidasi:** Versi ini menggabungkan `PRD_KODEKABI_Jejak_Algoritma.md` (v1.0), `PRD_KODEKABI_Shop_Redeem_Addendum.md`, dan `PRD_KODEKABI_Admin_Console_Addendum.md` menjadi satu dokumen, dengan seluruh revisi terakhir sudah diterapkan: role `content_author` digabung ke `admin` (AUTH-05 direvisi), 2FA wajib untuk login admin, CMS Case tanpa tahap review terpisah, skema evidence tanpa field avatar, dan thumbnail case wajib per-case. Modul baru (Wallet, Shop, Redeem, Admin Console) ada di Bagian 19.

---

## Kontrol Dokumen

### Riwayat Versi

| Versi | Tanggal | Disusun Oleh | Deskripsi Perubahan |
|---|---|---|---|
| 0.1 | 24 Jul 2026 | Tim Produk KODEKABI | Draf awal disusun berdasarkan Dokumen Konsep Produk KODEKABI: Jejak Algoritma |
| 1.0 | 24 Jul 2026 | Tim Produk KODEKABI | PRD lengkap: goals, scope, functional & non-functional requirements, arsitektur, roadmap |

### Distribusi & Persetujuan

| Peran | Tanggung Jawab Review | Status |
|---|---|---|
| Product Owner / Ketua Tim | Kelengkapan lingkup, prioritas fitur, keselarasan tujuan bisnis | Menunggu review |
| Engineering Lead (Backend Go) | Kelayakan teknis modul backend, skema data, scoring engine | Menunggu review |
| Engineering Lead (Frontend Next.js/React) | Kelayakan UI tiga kolom, PWA, state management | Menunggu review |
| UX/UI Designer | Kelengkapan user journey, wireframe, aksesibilitas | Menunggu review |
| AI/ML Engineer | Kelayakan contextual chatbot & semantic evaluation layer | Menunggu review |
| Mentor / Juri UNESCO Youth Hackathon | Keselarasan dengan tema literasi digital & agensi epistemik | Menunggu review |

> **Catatan:** Dokumen ini merupakan turunan (derivative) dari Dokumen Konsep Produk KODEKABI: Jejak Algoritma yang diunggah pengguna, disusun ulang ke dalam format PRD standar industri. Bagian yang ditandai **"Diusulkan"** merupakan tambahan dari kerangka PRD yang belum eksplisit disebutkan pada dokumen konsep dan perlu divalidasi bersama stakeholder sebelum difinalisasi.

---

## Daftar Isi

1. [Ringkasan Eksekutif](#1-ringkasan-eksekutif)
2. [Latar Belakang dan Rumusan Masalah](#2-latar-belakang-dan-rumusan-masalah)
3. [Tujuan Produk dan Metrik Keberhasilan](#3-tujuan-produk-dan-metrik-keberhasilan)
4. [Target Pengguna](#4-target-pengguna)
5. [Ruang Lingkup Produk](#5-ruang-lingkup-produk)
6. [Proposisi Nilai dan Positioning](#6-proposisi-nilai-dan-positioning)
7. [Peta Perjalanan Pengguna (User Journey)](#7-peta-perjalanan-pengguna-user-journey)
8. [Mekanik Inti Permainan dan Requirement Layar](#8-mekanik-inti-permainan-dan-requirement-layar)
9. [Functional Requirements per Modul](#9-functional-requirements-per-modul)
10. [Non-Functional Requirements](#10-non-functional-requirements)
11. [Arsitektur dan Requirement Teknis](#11-arsitektur-dan-requirement-teknis)
12. [Data dan API Requirements](#12-data-dan-api-requirements)
13. [Requirement AI dan Guardrail](#13-requirement-ai-dan-guardrail)
14. [Analitik dan Instrumentasi](#14-analitik-dan-instrumentasi)
15. [Risiko, Asumsi, dan Dependensi (RAID)](#15-risiko-asumsi-dan-dependensi-raid)
16. [Roadmap dan Rencana Rilis](#16-roadmap-dan-rencana-rilis)
17. [Referensi Alur Sistem](#17-referensi-alur-sistem)
18. [Pertanyaan Terbuka dan Keputusan yang Diperlukan](#18-pertanyaan-terbuka-dan-keputusan-yang-diperlukan)
19. [Modul Tambahan: Shop, Redeem, dan Admin Console](#19-modul-tambahan-shop-redeem-dan-admin-console)
20. [Lampiran](#20-lampiran)

---

## 1. Ringkasan Eksekutif

KODEKABI: Jejak Algoritma adalah game simulasi investigasi berbasis web yang menempatkan pemain sebagai **Auditor Digital** di kota virtual bernama **Kota Nusa**. Produk ini dirancang untuk menjawab masalah rendahnya agensi epistemik generasi muda dalam menghadapi AI generatif, algoritma rekomendasi, dan ekosistem informasi digital yang kompleks — bukan dengan ceramah literasi digital formal, melainkan dengan pengalaman investigasi yang membuat pemain mengalami langsung proses pengambilan keputusan, melakukan kesalahan secara aman, dan melihat konsekuensinya terhadap kondisi kota.

### Masalah yang Diselesaikan
Pengguna muda usia 15–24 tahun cenderung menganggap bahasa AI yang meyakinkan sebagai bukti kebenaran, menggunakan keluaran AI tanpa verifikasi, memercayai informasi karena populer di kelompoknya, sulit membedakan fakta/opini/pengalaman pribadi, dan tidak memahami bagaimana interaksi mereka membentuk ruang gema algoritmik.

### Solusi Produk
Pemain memilih case investigasi dari City Dashboard, membuka evidence di layar investigasi tiga kolom, menguji chatbot kontekstual yang terlibat dalam kasus, menjawab pertanyaan terstruktur maupun terbuka, lalu mengirim keputusan akhir. Sistem backend menilai proses investigasi secara deterministik (bukan hanya jawaban akhir), menampilkan dampak terhadap kondisi Kota Nusa, dan memperbarui reputasi serta level Auditor.

### Target Pengguna
- **Primer:** remaja akhir dan dewasa muda usia 15–24 tahun — siswa SMA, mahasiswa, pengguna aktif AI generatif, media sosial, dan forum digital.
- **Sekunder:** guru, dosen, fasilitator literasi digital, komunitas pemuda, organisasi pendidikan, dan program literasi media/informasi yang memanfaatkan hasil permainan sebagai indikator kompetensi.

### Proposisi Nilai Utama
Satu gameplay shell yang konsisten dapat menjalankan berbagai case investigasi yang dibuat melalui schema terstruktur, dinilai menggunakan behavioral scoring yang dapat dijelaskan (explainable), diperkaya chatbot kontekstual, dan dihubungkan langsung dengan perubahan kondisi kota — sehingga penambahan case baru tidak memerlukan pembuatan halaman atau logika unik.

### Ringkasan Tujuan Bisnis (Diusulkan)
- Memvalidasi model literasi digital berbasis game-based learning yang dapat diadopsi program literasi media/informasi berskala nasional maupun regional.
- Mencapai engagement dan retensi yang cukup untuk membuktikan efektivitas pembelajaran (peningkatan skor kompetensi antar sesi).
- Menjadi kandidat kuat pada UNESCO Youth Hackathon dengan MVP yang dapat didemonstrasikan end-to-end: pemilihan case → investigasi → keputusan → dampak kota.

> **Catatan:** Metrik dan target angka pada dokumen ini bersifat usulan awal (baseline hipotesis) karena dokumen konsep sumber belum menyertakan data historis atau target kuantitatif. Target perlu dikalibrasi ulang setelah uji coba pengguna (playtesting) pertama.

---

## 2. Latar Belakang dan Rumusan Masalah

### 2.1 Konteks Masalah
Adopsi AI generatif dan sistem rekomendasi algoritmik oleh generasi muda tumbuh jauh lebih cepat dibandingkan kemampuan verifikasi dan berpikir kritis mereka terhadap keluaran sistem tersebut. Dokumen konsep produk mengidentifikasi tujuh pola masalah informasi digital yang dialami Kota Nusa (representasi masalah dunia nyata):

- Saran kesehatan yang menyesatkan.
- Halusinasi chatbot.
- Artikel dengan judul manipulatif (clickbait).
- Statistik yang digunakan di luar konteks.
- Validasi informasi keliru di forum.
- Konten viral yang memperkuat konflik.
- Sistem rekomendasi yang membentuk ruang gema (echo chamber).

### 2.2 Akar Masalah pada Sisi Pengguna
Berdasarkan pemetaan customer jobs dan pains pada dokumen konsep, akar masalah terletak pada perilaku epistemik berikut:

- Menganggap bahasa AI yang meyakinkan sebagai bukti kebenaran, alih-alih memeriksa dasar/sumber jawaban.
- Menggunakan keluaran AI tanpa memahami atau memverifikasi isinya karena tekanan waktu tugas.
- Memercayai informasi karena popularitas sosial, bukan karena kualitas sumber.
- Tidak membedakan fakta, opini, pengalaman pribadi, dan diagnosis/klaim yang belum terverifikasi.
- Terlalu yakin (overconfident) meskipun bukti masih terbatas, atau sebaliknya enggan mengubah pendapat karena dianggap "kalah" secara sosial.
- Menolak sumber hanya karena berasal dari kelompok berbeda (bias in-group).
- Tidak memahami bagaimana interaksi mereka (like, share, watch time) memengaruhi rekomendasi algoritmik yang mereka terima berikutnya.

### 2.3 Mengapa Pendekatan yang Ada Belum Memadai
Materi literasi digital konvensional umumnya berbentuk ceramah atau modul teks panjang dengan contoh yang terlalu mudah ditebak, tanpa konsekuensi personal maupun rasa progres — sehingga tidak menimbulkan perubahan perilaku yang bertahan (retensi rendah, transfer pengetahuan lemah). Target pengguna primer juga secara eksplisit menyatakan tidak tertarik pada "materi pembelajaran formal" (lihat Bagian 4 — Persona Nadia).

### 2.4 Mengapa Sekarang (Why Now)
- Paparan AI generatif pada aktivitas belajar dan mencari informasi sehari-hari meningkat pesat pada kelompok usia 15–24 tahun.
- Kebutuhan program literasi media/informasi (pengguna sekunder: sekolah, kampus, lembaga pelatihan) akan alat ukur kompetensi yang lebih engaging dan terukur.
- Momentum UNESCO Youth Hackathon sebagai panggung validasi konsep sekaligus jalur distribusi awal ke komunitas pemuda dan institusi pendidikan.

---

## 3. Tujuan Produk dan Metrik Keberhasilan

### 3.1 Tujuan Bisnis (Business Goals)
- Memvalidasi KODEKABI sebagai model literasi digital berbasis simulasi yang terbukti meningkatkan kompetensi epistemik pemain secara terukur.
- Membangun arsitektur produk (satu gameplay shell + schema case generik) yang memungkinkan penambahan case baru tanpa pengembangan halaman/logika baru — menekan biaya pengembangan konten jangka panjang.
- Membuka jalur adopsi oleh pengguna sekunder (sekolah, kampus, lembaga literasi) sebagai alat asesmen kompetensi digital yang lebih menarik dibanding tes konvensional.
- Memenangkan/lolos kurasi UNESCO Youth Hackathon dengan MVP yang credible dan dapat didemonstrasikan end-to-end.

### 3.2 Tujuan Pengguna (User Goals)
- Merasa cakap menghadapi informasi digital dan keluaran AI tanpa merasa dihakimi ketika salah.
- Mempelajari cara memeriksa sumber, mengklasifikasikan klaim, dan mengelola tingkat keyakinan tanpa proses yang melelahkan.
- Mendapatkan identitas dan pengakuan (Auditor Digital) yang dapat dibandingkan secara sehat dengan pengguna lain.
- Melihat dampak nyata (simulatif) dari keputusannya terhadap kondisi Kota Nusa.

### 3.3 Non-Tujuan (Eksplisit di Luar Tujuan Produk)
- KODEKABI bukan city-builder — kota berfungsi sebagai dashboard progres tervisualisasikan, bukan sistem pembangunan kota yang kompleks.
- AI tidak dirancang menjadi sumber kebenaran tunggal maupun penentu skor akhir; keputusan skor selalu deterministik melalui rule di backend Go.
- Produk tidak dirancang untuk memberi tahu pemain secara langsung bahwa "misinformasi itu berbahaya" secara didaktik — fokus pada pengalaman mengambil keputusan dan mengalami konsekuensinya.

### 3.4 Metrik Keberhasilan (Diusulkan)

> Dokumen konsep sumber tidak menyertakan target metrik kuantitatif. Tabel berikut adalah usulan kerangka metrik standar produk game-based learning yang perlu disepakati bersama stakeholder dan dikalibrasi dari data playtesting awal.

| Kategori | Metrik | Definisi Pengukuran | Target Awal (Hipotesis) |
|---|---|---|---|
| North Star Metric | Weekly Verified Cases | Jumlah case yang diselesaikan pemain dengan final submission per minggu | Ditetapkan setelah playtesting Sprint 1 |
| Aktivasi | Tutorial Completion Rate | % pengguna baru yang menyelesaikan tutorial case dalam 24 jam pertama | ≥ 70% |
| Aktivasi | Time to First Case | Median waktu dari registrasi hingga submit case pertama | ≤ 10 menit |
| Engagement | Case Completed / Active User / Minggu | Rata-rata case yang diselesaikan per pengguna aktif mingguan | ≥ 3 case/minggu |
| Retensi | D7 Retention | % pengguna yang kembali membuka aplikasi pada hari ke-7 | ≥ 30% |
| Retensi | D30 Retention | % pengguna yang kembali membuka aplikasi pada hari ke-30 | ≥ 15% |
| Learning Outcome | Delta Confidence Calibration Score | Perubahan rata-rata Confidence Calibration Score antara 3 case pertama dan 3 case terakhir per pengguna | Kenaikan positif signifikan |
| Learning Outcome | Revision Willingness Rate | % keputusan akhir yang confidence-nya berubah dari initial judgment setelah membuka evidence | Meningkat seiring level |
| Kepercayaan Produk | AI Fallback Rate | % interaksi chatbot/semantic evaluation yang jatuh ke fallback karena kegagalan AI | ≤ 5% |
| Kepercayaan Produk | Distrust Signal Rate | % pemain yang melewati (skip) pembahasan feedback tanpa membaca (proxy: dwell time < 2 detik) | ≤ 10% |
| Adopsi Sekunder | Institutional Sign-up | Jumlah sekolah/komunitas/lembaga yang menggunakan hasil permainan untuk asesmen | Ditetapkan pasca-MVP |
| Teknis | API p95 Latency | Waktu respons endpoint inti (dashboard, session, submit) persentil ke-95 | ≤ 300 ms (non-AI endpoint) |

---

## 4. Target Pengguna

### 4.1 Pengguna Primer
Remaja akhir dan dewasa muda usia sekitar 15–24 tahun, khususnya:
- Siswa sekolah menengah.
- Mahasiswa.
- Pengguna aktif AI generatif.
- Pengguna media sosial dan video pendek.
- Pengguna forum serta komunitas digital.
- Pemuda yang terbiasa mencari informasi secara cepat melalui internet.
- Pengguna dengan kemampuan digital operasional, tetapi belum memiliki kebiasaan verifikasi yang konsisten.

### 4.2 Pengguna Sekunder
Guru, dosen, fasilitator literasi digital, komunitas pemuda, organisasi pendidikan, lembaga pelatihan, serta program literasi media dan informasi. Pengguna sekunder tidak wajib terlibat dalam setiap sesi permainan; mereka memanfaatkan hasil permainan untuk memahami perkembangan kompetensi pengguna primer.

### 4.3 Persona Utama: Nadia

| Atribut | Detail |
|---|---|
| Usia | 18 tahun |
| Status | Mahasiswa tahun pertama |
| Perangkat | Laptop dan smartphone |
| Kebiasaan Digital | Menggunakan chatbot untuk tugas, mencari informasi lewat media sosial, menonton video pendek, membaca diskusi forum |
| Motivasi | Menyelesaikan tugas secara cepat, memperoleh hasil baik, tetap terlihat kompeten di depan teman/dosen |
| Hambatan (Barriers) | Sulit memeriksa sumber, mudah percaya pada jawaban yang terdengar profesional, tidak menikmati materi literasi berbentuk ceramah |
| Harapan (Aspirations) | Mampu menggunakan AI tanpa mudah tertipu, melalui pengalaman yang terasa seperti game, bukan pelatihan formal |

### 4.4 Jobs to Be Done

**Functional Jobs**
- Menentukan apakah suatu informasi dapat dipercaya.
- Memahami isi jawaban AI, bukan sekadar menerimanya.
- Memeriksa sumber tanpa proses yang rumit.
- Membedakan fakta, opini, pengalaman, dan klaim belum terverifikasi.
- Mengenali informasi yang kehilangan konteks.
- Mengambil keputusan ketika bukti tidak sempurna.
- Menggunakan AI secara aman dan produktif.
- Menghindari penyebaran informasi keliru dan memahami dampak tindakannya terhadap orang lain.

**Emotional Jobs**
- Merasa cakap menghadapi informasi digital.
- Tidak merasa bodoh ketika tertipu.
- Berani mengubah jawaban tanpa merasa "kalah".
- Merasa aman melakukan kesalahan.
- Merasakan kepuasan saat membongkar pola tersembunyi.

**Social Jobs**
- Terlihat kompeten dan tidak dikenal sebagai penyebar misinformasi.
- Tetap diterima dalam kelompok meski berbeda pendapat.
- Menyampaikan koreksi tanpa mempermalukan orang lain.
- Mendapatkan pengakuan sebagai auditor yang andal (status, ranking sehat).

---

## 5. Ruang Lingkup Produk

### 5.1 Dalam Lingkup MVP
- Registrasi/login, profil dasar, dan onboarding/tutorial interaktif.
- City Dashboard dengan minimal 4–5 indikator kondisi kota (Information Health, Public Trust, Social Stability, Public Wellbeing, Auditor Reputation).
- Case catalog dengan minimal template: social post, artikel, dan percakapan chatbot (prioritas MVP; template lain menyusul).
- Investigation Screen tiga kolom (evidence list, konten utama, panel pertanyaan) yang generik untuk seluruh case.
- Structured Question Engine: structured choice, confidence input, claim classification, open question.
- Contextual chatbot dasar dengan guardrail dan fallback response.
- Behavioral Epistemic Scoring Engine berbasis rule di backend Go (deterministik).
- Result screen dengan score breakdown, feedback, dan city impact.
- Progression system (XP, level, reputasi) dan minimal 1 mekanisme unlocking case.
- Penyimpanan draft jawaban lokal (IndexedDB) dan sinkronisasi ulang saat koneksi pulih.

### 5.2 Di Luar Lingkup MVP (Fase Berikutnya)
- Leaderboard penuh (mingguan, cohort/kelompok) — MVP dapat menyediakan versi minimal atau ditunda.
- Dashboard khusus pengguna sekunder (guru/fasilitator) untuk memantau kompetensi kelompok.
- Content authoring tool berbasis AI untuk mempercepat pembuatan draft case oleh content author eksternal.
- Elemen kosmetik/reward tambahan, audio, dan aset visual PixiJS lanjutan (animasi kota kompleks).
- Multi-bahasa (di luar Bahasa Indonesia).
- Mode kolaboratif/multiplayer.

### 5.3 Asumsi
- Pengguna primer memiliki akses perangkat (laptop/smartphone) dan koneksi internet meski tidak selalu stabil — ditangani melalui pendekatan PWA dan draft lokal.
- Model AI eksternal atau open-source endpoint tersedia dengan biaya dan kebijakan data yang dapat diterima untuk chatbot kontekstual dan semantic evaluation.
- Tim pengembangan awal berukuran kecil sehingga arsitektur modular monolith lebih sesuai dibanding microservices.

### 5.4 Batasan (Constraints)
- Timeline mengikuti jadwal UNESCO Youth Hackathon — MVP harus dapat didemonstrasikan end-to-end dalam periode kompetisi.
- AI tidak boleh menjadi penentu skor final atau sumber kebenaran tunggal (prinsip arsitektur non-negotiable — lihat Bagian 11).
- Data sensitif (riwayat media sosial asli, percakapan pribadi, preferensi politik nyata, diagnosis kesehatan) tidak boleh dikumpulkan (lihat Bagian 12.2).

---
## 6. Proposisi Nilai dan Positioning

### 6.1 Makna Nama Produk

| Elemen Nama | Makna |
|---|---|
| KODE | Merepresentasikan kode, teka-teki, pesan tersembunyi, dan aktivitas pemecahan misteri. |
| KABI | Maskot Kapibara — identitas Nusantara sekaligus representasi kota/dunia digital tempat pemain bertugas. |
| Jejak Algoritma | Fokus permainan pada penelusuran pola, sumber informasi, keputusan AI, dan dampak sistem rekomendasi. |

Nama produk sengaja bernuansa detektif dan riddle, modern, dan tidak terdengar seperti platform pembelajaran formal — selaras dengan hambatan persona Nadia yang tidak menikmati materi literasi berbentuk ceramah.

### 6.2 Tagline
- **Tagline Utama:** "Bongkar informasinya. Baca polanya. Selamatkan kotanya."
- **Tagline Alternatif:** "Setiap informasi meninggalkan jejak."

### 6.3 Unique Value Proposition (UVP)
KODEKABI memberikan pengalaman investigasi digital yang membuat pemain tidak hanya menilai apakah suatu informasi benar atau salah, tetapi memahami bagaimana AI, sumber, popularitas, dan tingkat keyakinannya memengaruhi keputusan serta kondisi sebuah kota.

### 6.4 Unique Selling Proposition (USP)
Satu gameplay shell dapat menjalankan berbagai case investigasi yang dibuat melalui schema terstruktur, dinilai menggunakan behavioral scoring, diperkaya chatbot kontekstual, dan dihubungkan langsung dengan perubahan kondisi kota.

### 6.5 Value Proposition Fit
Bagi pengguna muda yang ingin lebih sulit diperdaya oleh AI dan informasi digital tetapi tidak tertarik pada pelatihan formal, KODEKABI menyediakan simulasi investigasi berbasis case yang singkat, visual, dan konsekuensial, sehingga pengguna dapat melatih cara berpikir sambil membangun reputasi serta menjaga kondisi kota virtualnya.

---

## 7. Peta Perjalanan Pengguna (User Journey)

| Tahap | Aksi Pengguna Kunci | Pain Point Kritis | Implikasi Requirement Produk |
|---|---|---|---|
| 1. Awareness & Discovery | Melihat teaser, membuka landing page, membaca contoh case | Khawatir game sebenarnya hanya kuis; tidak ingin materi sekolah tambahan | Tampilkan city dashboard & case screenshot; gunakan tagline misteri; sediakan 1 demo case singkat |
| 2. Registration & Onboarding | Membuat akun, memilih avatar, menjalani tutorial | Bingung terhadap statistik kota & confidence slider; khawatir dinilai sekolah | Batasi data registrasi minimum; ajarkan satu mekanik per langkah; sediakan skip untuk penjelasan non-esensial |
| 3. City Observation | Masuk ke home, memeriksa statistik & level | Bingung memilih case; merasa terlalu banyak informasi | Batasi statistik utama 4–5 indikator; beri rekomendasi case; tandai statistik yang terancam |
| 4. Case Selection | Membaca ringkasan, melihat risk level & dampak | Khawatir case terlalu sulit; keluar karena loading lama | Label difficulty; tampilkan learning focus sederhana; prefetch data case setelah preview dibuka |
| 5. Case Briefing | Membaca situasi, mengisi confidence awal | Takut penilaian awal dianggap salah; tidak paham arti confidence | Brief singkat; jelaskan bahwa jawaban awal dapat berubah; beri contoh arti confidence |
| 6. Core Investigation | Membuka evidence, menjawab pertanyaan, memakai chatbot | Kewalahan membaca evidence; mengira evidence populer pasti benar; bingung bertanya ke chatbot | Batasi evidence 3–5; sediakan suggested questions; simpan draft otomatis; tampilkan sync status |
| 7. Final Decision | Memeriksa jawaban, menulis alasan, submit | Takut salah/menebak jawaban ideal; submit ganda | Tampilkan ringkasan jawaban; tandai field belum selesai; gunakan idempotency key; nonaktifkan tombol saat proses |
| 8. Feedback & City Impact | Melihat konsekuensi, score breakdown, dampak kota | Merasa feedback menghakimi; tidak percaya evaluasi AI | Tampilkan konsekuensi sebelum skor; pisahkan rule-based score vs AI feedback; gunakan bahasa netral |
| 9. Retention & Progression | Kembali ke kota, cek profil/leaderboard, pilih case berikutnya | Bosan jika case serupa; membandingkan diri secara tidak sehat | Variasikan tema case; rekomendasi berbasis kompetensi; hindari data sensitif pada leaderboard |

---

## 8. Mekanik Inti Permainan dan Requirement Layar

### 8.1 Alur Mekanik Inti
**Observe → Examine → Respond → Decide → See the Impact**

| Fase | Requirement Fungsional |
|---|---|
| Observe | Pemain melihat kondisi Kota Nusa dan memilih case dari dashboard. Case ditampilkan sebagai thumbnail simulatif: postingan media sosial, artikel berita, blog, thread forum, percakapan chatbot, atau pengumuman publik. |
| Examine | Pemain membuka evidence card, profil sumber, percakapan, statistik, dan pernyataan relevan dalam urutan bebas (tidak linear). |
| Respond | Pemain menjawab kombinasi structured choice, confidence input, claim classification, dan open question sesuai kompetensi yang diukur pada case tersebut. |
| Decide | Pemain memilih tindakan akhir, menetapkan confidence akhir, menuliskan alasan singkat, meninjau ulang seluruh jawaban, lalu mengirim final submission. |
| See the Impact | Sistem menampilkan konsekuensi case, score breakdown, perubahan confidence awal vs akhir, evidence yang digunakan/terlewat, feedback personal, serta perubahan statistik kota dan reputasi auditor. |

### 8.2 Requirement Layar: City Dashboard
Berfungsi sebagai visualisasi kondisi Kota Nusa — bukan city-builder. Pada MVP, kota adalah dashboard progres yang divisualisasikan melalui background, indikator, dan perubahan kondisi. Informasi wajib ditampilkan: Information Health, Public Trust, Social Stability, Public Wellbeing, Auditor Reputation, Level pemain, Case aktif, Dampak keputusan terakhir.

### 8.3 Requirement Layar: Case Selection
Setiap case card wajib memuat: thumbnail, judul, jenis sumber, tingkat risiko, estimasi durasi, statistik kota yang terancam, dan status penyelesaian.

### 8.4 Requirement Layar: Investigation Screen (Tiga Kolom)

| Kolom | Proporsi Lebar | Konten Wajib |
|---|---|---|
| Kiri | ≈ 15% | Case overview, daftar evidence, status evidence, indikator progress |
| Tengah | ≈ 50% | Visual case, detail evidence, artikel, social post, chat transcript, statistik, profil sumber |
| Kanan | ≈ 35% | Daftar pertanyaan, confidence input, claim classification, open question, contextual chatbot, final submission |

> Kolom kiri harus dapat diperkecil (collapsible) dan seluruh layar menggunakan responsive mode pada layar kecil.

### 8.5 Requirement Layar: Result Screen
Wajib menampilkan, dengan urutan: (1) konsekuensi case, (2) penjelasan keputusan, (3) skor kompetensi, (4) feedback, (5) dampak kota, (6) experience point, (7) auditor reputation, (8) rekomendasi case berikutnya. Konsekuensi ditampilkan sebelum skor, dan hanya disorot satu kekuatan serta satu area perbaikan agar tidak terasa menghakimi.

---

## 9. Functional Requirements per Modul

> Prioritas MoSCoW (Must/Should/Could/Won't) adalah usulan awal berdasarkan tingkat kekritisan terhadap alur inti (core loop), perlu dikonfirmasi ulang saat sprint planning.

### 9.1 Auth Module
Mengelola registrasi, autentikasi, sesi, dan otorisasi peran pengguna.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| AUTH-01 | Registrasi akun Auditor | Pengguna membuat akun baru dengan username unik, avatar, rentang usia (age range), dan status consent. | Must | Sistem menolak username duplikat; avatar wajib dipilih sebelum akun aktif; data tersimpan sesuai skema data minimum (lihat 12.2). |
| AUTH-02 | Login dengan kredensial | Pengguna login menggunakan identitas terdaftar. | Must | Login gagal menampilkan pesan generik (tidak membocorkan username/password mana yang salah); rate limit percobaan gagal. |
| AUTH-03 | Manajemen sesi/token | Sistem menerbitkan token/sesi dengan masa berlaku dan mekanisme refresh. | Must | Token kedaluwarsa memaksa re-autentikasi; token disimpan aman di client. |
| AUTH-04 | Logout & invalidasi token | Pengguna dapat logout dan token/sesi langsung tidak valid di backend. | Must | Request menggunakan token yang telah logout ditolak dengan 401. |
| AUTH-05 | Role-based access control | Sistem membedakan peran player dan admin (direvisi dari tiga role — `content_author` digabung ke `admin`, lihat Bagian 19.2). | Should | Endpoint content authoring dan seluruh endpoint `/admin/*` hanya dapat diakses role admin. |
| AUTH-06 | Consent & data minimization saat onboarding | Sistem mencatat persetujuan pengguna atas kebijakan privasi sebelum data gameplay pertama dikumpulkan. | Must | Tidak ada event gameplay yang tersimpan sebelum consent_status = true. |

### 9.2 User / Profile Module
Mengelola profil, level, experience point, dan statistik kompetensi pemain.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| USER-01 | Profil dasar | Menyimpan dan menampilkan username, avatar, level, dan XP. | Must | Perubahan avatar/username tersimpan dan tercermin di seluruh layar dalam <1 refresh. |
| USER-02 | Statistik kompetensi | Menyimpan lima skor kompetensi: Evidence Evaluation, Claim Analysis, Confidence Calibration, Reasoning, Safety Judgment. | Must | Statistik diperbarui otomatis setiap kali sebuah case diselesaikan. |
| USER-03 | Riwayat case (case history) | Menampilkan daftar case yang telah diselesaikan pemain beserta outcome-nya. | Must | Riwayat dapat difilter berdasarkan status dan diurutkan berdasarkan tanggal. |
| USER-04 | Auditor reputation tracking | Menghitung dan menampilkan reputasi auditor berdasarkan akumulasi keputusan. | Must | Reputasi tidak dapat menjadi negatif tak terbatas; ada batas bawah terdefinisi. |
| USER-05 | Edit profil terbatas | Pemain dapat mengganti avatar dan (dengan batasan) username. | Could | Perubahan username dibatasi frekuensinya (maks. 1x per 30 hari). |

### 9.3 City Module
Mengelola kondisi Kota Nusa sebagai representasi visual dari agregat performa investigasi.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| CITY-01 | City state & statistik utama | Menyimpan dan menyajikan Information Health, Public Trust, Social Stability, Public Wellbeing. | Must | Setiap statistik memiliki rentang nilai terdefinisi (mis. 0–100) dan status warna (aman/terancam/kritis). |
| CITY-02 | Visual state kota | Menentukan aset visual/background berdasarkan kondisi statistik agregat. | Should | Perubahan status memicu perubahan visual yang terlihat pemain. |
| CITY-03 | Impact history log | Mencatat riwayat perubahan kondisi kota akibat setiap keputusan case. | Must | Setiap entri log tertaut ke session/case ID penyebabnya. |
| CITY-04 | Update statistik pasca-case | Statistik kota diperbarui secara transaksional segera setelah scoring selesai. | Must | Update city state dan penyimpanan result terjadi dalam satu transaksi database. |

### 9.4 Case Module (Case Simulation Engine)
Mengelola katalog, skema, versi, dan konten seluruh case investigasi.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| CASE-01 | Case catalog & filtering | Menampilkan daftar case tersedia beserta metadata (risiko, durasi, jenis sumber, status). | Must | Case belum ter-unlock tetap terlihat dalam status locked dengan syarat pembuka jelas. |
| CASE-02 | Schema case generik | Setiap case (briefing, evidence, questions, chatbot context, scoring rule, outcome) disimpan dalam struktur data identik. | Must | Frontend dapat merender case baru tanpa perubahan kode/komponen. |
| CASE-03 | Case versioning & snapshot | Setiap sesi case mengambil snapshot versi case saat dimulai, bukan versi terbaru real-time. | Must | Perubahan konten case oleh content author tidak memengaruhi sesi berjalan. |
| CASE-04 | Prasyarat/unlocking case | Case tertentu memerlukan level/reputasi/penyelesaian case lain sebagai syarat akses. | Must | Percobaan akses case terkunci menghasilkan 403 dengan alasan jelas. |
| CASE-05 | Template sumber case | Mendukung minimal enam tipe template: social post, artikel, blog, thread forum, percakapan chatbot, pengumuman publik. | Must | Setiap template memiliki komponen render generik untuk data case apa pun sesuai tipenya. |
| CASE-06 | Bantuan draft case berbasis AI | AI membantu content author menghasilkan draft case terstruktur sesuai schema. | Could | Draft AI wajib melalui review manusia sebelum dipublikasikan. |

### 9.5 Session Module
Mengelola siklus hidup satu sesi permainan case.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| SESSION-01 | Start case & snapshot | Memulai sesi baru dan membuat snapshot case + session ID unik. | Must | Satu pemain tidak dapat memiliki dua sesi aktif untuk case yang sama bersamaan. |
| SESSION-02 | Autosave draft jawaban | Draft jawaban disimpan lokal (IndexedDB) dan disinkronkan periodik ke backend. | Must | Menutup aplikasi tidak sengaja tidak menghilangkan progres draft. |
| SESSION-03 | Session versioning | Setiap perubahan state sesi memiliki nomor versi untuk mencegah stale write. | Must | Request dengan session version tidak sesuai ditolak, frontend diminta refresh. |
| SESSION-04 | Resume setelah koneksi terputus | Sesi dapat dilanjutkan dari kondisi terakhir tersimpan setelah koneksi pulih. | Must | Status sinkronisasi ditampilkan; pengiriman diulang otomatis saat koneksi tersedia. |
| SESSION-05 | Idempotency pada seluruh mutasi | Setiap request mutasi menyertakan idempotency key. | Must | Pengiriman ulang dengan key sama tidak menghasilkan duplikasi data. |

### 9.6 Answer Module
Mengelola seluruh jenis jawaban pemain hingga final submission.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ANSWER-01 | Structured choice | Mendukung input pilihan terstruktur (single/multiple choice) per pertanyaan. | Must | Validasi tipe jawaban di frontend & divalidasi ulang di backend sesuai question type. |
| ANSWER-02 | Confidence input | Mendukung input tingkat keyakinan awal dan akhir (slider 0–100%). | Must | Perubahan confidence ekstrem meminta konfirmasi tambahan. |
| ANSWER-03 | Claim classification | Pemain mengklasifikasikan klaim (fakta/opini/pengalaman pribadi/belum terverifikasi). | Must | Kategori klasifikasi terdefinisi dalam schema case, bukan hardcode. |
| ANSWER-04 | Open question | Mendukung input teks bebas untuk penalaran tertulis. | Must | Open question dibatasi hanya pada keputusan penting. |
| ANSWER-05 | Validasi final submission | Sistem memastikan seluruh pertanyaan wajib telah dijawab sebelum submission diterima. | Must | Field belum lengkap ditandai visual; submission diblokir hingga lengkap. |
| ANSWER-06 | Revisi sebelum submit | Pemain dapat mengubah jawaban sebelum submit akhir, tidak setelahnya. | Must | Setelah final submit berhasil, submit ulang ditolak (idempotent), tidak ada opsi edit. |

### 9.7 Chatbot Module (Contextual AI Evaluation Layer)
Menyediakan chatbot dalam-case dengan guardrail ketat.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| CHATBOT-01 | Context building per case | Prompt dibangun dari briefing, allowed evidence, persona, knowledge boundary, prohibited behavior. | Must | Chatbot tidak menjawab di luar knowledge boundary case; percobaan keluar konteks direspons penolakan halus. |
| CHATBOT-02 | Suggested questions | Menyediakan daftar pertanyaan awal untuk mengurangi kebingungan. | Must | Minimal 3 suggested questions tersedia saat chatbot pertama dibuka. |
| CHATBOT-03 | Riwayat percakapan | Menyimpan chat history per sesi untuk debrief dan scoring konteks. | Must | Riwayat chat tetap dapat dilihat di result screen bila relevan. |
| CHATBOT-04 | Guardrail konten | Mencegah chatbot menghasilkan konten di luar cakupan pembelajaran. | Must | Respons AI divalidasi backend sebelum dikirim; respons melanggar guardrail diganti fallback. |
| CHATBOT-05 | Fallback response | Menyediakan respons cadangan saat AI gagal/timeout. | Must | Gameplay tetap dapat dilanjutkan tanpa error blocking. |

### 9.8 Scoring Module (Behavioral Epistemic Scoring Engine)
Menilai kualitas proses investigasi secara deterministik.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| SCORING-01 | Rule-based deterministic scoring | Backend Go menghitung skor final berdasarkan aturan eksplisit. | Must | Skor dari input identik selalu identik (reprodusibel, dapat diuji unit test). |
| SCORING-02 | Semantic evaluation open question | AI menghasilkan semantic signals dari jawaban terbuka, tanpa menentukan skor akhir. | Must | Kegagalan AI otomatis fallback deterministic rubric, tidak memblokir submission. |
| SCORING-03 | Score breakdown 5 dimensi | Menghasilkan skor terpisah: Evidence Evaluation, Claim Analysis, Confidence Calibration, Reasoning, Safety Judgment. | Must | Result screen menampilkan kelima skor terpisah, bukan hanya total. |
| SCORING-04 | Penentuan outcome & city impact | Menentukan narasi konsekuensi dan dampak statistik kota berdasarkan skor & keputusan. | Must | Aturan pemetaan skor→outcome→city impact terdokumentasi per case. |
| SCORING-05 | Explainability | Setiap komponen skor dapat dijelaskan ke pemain. | Must | Result screen menyorot minimal satu kekuatan dan satu area perbaikan. |

### 9.9 Progression Module
Mengelola experience point, level, reputasi, dan pembukaan case lanjutan.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| PROG-01 | Perhitungan XP & Level | XP bertambah berdasarkan hasil case; level naik pada ambang XP tertentu. | Must | Kenaikan level memicu notifikasi dan pembaruan case yang ter-unlock. |
| PROG-02 | Penyesuaian reputasi auditor | Reputasi berubah berdasarkan kualitas keputusan, bukan hanya penyelesaian case. | Must | Dua pemain dengan jumlah case sama tetapi kualitas berbeda dapat memiliki reputasi berbeda. |
| PROG-03 | Logika unlocking case | Menentukan case berikutnya berdasarkan level/reputasi/kompetensi. | Must | Preview case yang akan terbuka ditampilkan ke pemain. |
| PROG-04 | Update statistik kompetensi | Memperbarui lima statistik kompetensi setelah setiap case selesai. | Must | Statistik kompetensi menjadi dasar rekomendasi case berikutnya. |

### 9.10 Leaderboard Module
Menyediakan perbandingan pencapaian antar pemain secara sehat dan privacy-safe.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| LB-01 | Ranking mingguan | Menampilkan peringkat berdasarkan skor/aktivitas mingguan. | Should | Leaderboard di-reset/diperbarui setiap awal minggu sesuai zona waktu ditetapkan. |
| LB-02 | Cohort ranking | Mendukung ranking dalam kelompok (per sekolah/komunitas) untuk pengguna sekunder. | Could | Fitur ini didorong ke fase pasca-MVP. |
| LB-03 | Cached leaderboard | Leaderboard dilayani dari cache (Redis), bukan query langsung. | Must | Cache diinvalidasi/diperbarui setelah transaksi scoring berhasil. |
| LB-04 | Tampilan privacy-safe | Leaderboard tidak menampilkan data sensitif atau identitas asli di luar username/avatar. | Must | Tidak ada field selain username, avatar, level, dan skor agregat yang ditampilkan publik. |

### 9.11 Wallet, Shop, Redeem, dan Admin Console

Modul Wallet, Shop, Redeem, serta seluruh modul Admin Console (User Management, Case CMS, Shop & Redeem CMS, Moderation, Dashboard Operasional, Konfigurasi Game, Tata Kelola AI, Leaderboard Admin, Audit Log, Pengumuman, Reporting) ditambahkan pasca-versi awal PRD ini. Requirement lengkap masing-masing modul (ID, deskripsi, prioritas, acceptance criteria) ada di **Bagian 19 — Modul Tambahan: Shop, Redeem, dan Admin Console**, agar tidak menduplikasi konten di dua tempat.

---
## 10. Non-Functional Requirements

| Kategori | Requirement | Target / Kriteria |
|---|---|---|
| Performa | Latensi endpoint inti (dashboard, session, answers, submit) | p95 ≤ 300 ms tidak termasuk panggilan AI eksternal |
| Performa | Latensi respons chatbot kontekstual | p95 ≤ 3 detik, dengan fallback bila melebihi timeout |
| Skalabilitas | Sesi investigasi konkuren | Arsitektur modular monolith harus dapat diskalakan horizontal di tingkat proses backend Go tanpa mengubah domain module boundary |
| Ketersediaan | Uptime layanan inti (non-AI) | ≥ 99.5% selama periode hackathon dan pasca-launch awal |
| Ketahanan | Kegagalan AI eksternal | Sistem wajib tetap dapat menyelesaikan case (fallback chatbot & fallback rubric scoring) tanpa AI |
| Offline / PWA | Kehilangan koneksi selama investigasi | Draft jawaban tersimpan di IndexedDB dan disinkronkan otomatis saat koneksi pulih |
| Keamanan | Autentikasi & otorisasi | Seluruh endpoint mutasi memerlukan token valid; validasi kepemilikan session |
| Keamanan | Rate limiting | Endpoint auth, chatbot, dan submission dilindungi rate limit |
| Keamanan | Sanitasi input | Seluruh input teks disanitasi sebelum disimpan atau dikirim ke AI |
| Keamanan | Prompt/context boundary | Prompt ke AI dibangun dari context terbatas yang divalidasi backend, bukan input mentah pengguna langsung |
| Privasi | Minimisasi data | Tidak mengumpulkan riwayat media sosial asli, percakapan pribadi, preferensi politik nyata, atau diagnosis kesehatan |
| Konsistensi Data | Idempotency | Seluruh mutasi menggunakan idempotency key |
| Konsistensi Data | Transaksi atomik | Penyimpanan result, update progress, dan update city state dalam satu transaksi database |
| Aksesibilitas | Kontras & navigasi keyboard | WCAG 2.1 level AA sebagai target minimum pada komponen interaktif inti |
| Responsivitas | Dukungan perangkat | Layar investigasi tiga kolom wajib memiliki mode responsif untuk layar kecil |
| Auditabilitas | Explainability skor | Setiap keputusan scoring dapat ditelusuri ke rule, evidence, dan jawaban |
| Maintainability | Skema case generik | Penambahan case baru tidak memerlukan perubahan kode frontend |
| Lokalisasi | Bahasa | Bahasa Indonesia sebagai bahasa utama pada MVP |

---

## 11. Arsitektur dan Requirement Teknis

### 11.1 Prinsip Arsitektur
Arsitektur MVP menggunakan pendekatan **modular monolith**, dengan prinsip utama: *reusable content structure, deterministic game logic, AI assisted experience*. Dipilih karena tim pengembangan awal relatif kecil, batas domain masih dapat berubah, deployment lebih sederhana, transaksi case dan scoring membutuhkan konsistensi kuat, debugging lebih mudah, dan modul masih dapat diekstrak menjadi service terpisah di kemudian hari.

### 11.2 Diagram Arsitektur Tingkat Tinggi
- Web Browser / Installed PWA — Next.js, React, PixiJS → terhubung via HTTPS REST API
- Go Modular Monolith Backend — memuat Auth, User, City, Case, Session, Answer, Chatbot, Scoring, Progression, dan Leaderboard Module
- MariaDB — transactional source of truth
- Redis — cache, rate limit, session sementara, queue proses AI, distributed lock, leaderboard cache
- Object Storage kompatibel S3 — thumbnail, background kota, avatar, aset PixiJS
- External AI API — chatbot kontekstual & semantic evaluation, dipanggil dengan context terbatas dari backend

### 11.3 Tech Stack dan Rasionalisasi

**Frontend**

| Teknologi | Fungsi Utama | Rasionalisasi |
|---|---|---|
| Next.js + TypeScript | Routing, home, auth, case history, profile, leaderboard, PWA config | Fondasi aplikasi web terstruktur, App Router dengan Server/Client Components |
| React | Gameplay shell, evidence list, question panel, chatbot, form state, result screen | Sesuai untuk antarmuka dengan banyak state interaktif dan komponen reusable |
| PixiJS | City background, animasi kondisi kota, loading screen, efek transisi | Visualisasi 2D interaktif tanpa kompleksitas game engine penuh |
| IndexedDB | Draft jawaban, cache case, pending interaction events, pemulihan progres | Mencegah hilangnya jawaban sebelum final submit saat koneksi terganggu |

**Backend & Data**

| Teknologi | Fungsi Utama | Rasionalisasi |
|---|---|---|
| Go | REST API, auth, game session, case orchestration, scoring, city impact, progression, integrasi AI | Backend efisien, mudah dikemas, cocok menangani banyak event dan proses scoring eksplisit |
| Gin / Echo | HTTP framework | Routing, middleware, validation integration |
| sqlc | Generator kode Go type-safe dari query SQL | Query eksplisit dan mudah diaudit dengan typing aman |
| Goose | Migration database | Perubahan schema tercatat dan diterapkan konsisten |
| MariaDB | Source of truth relasional | Memenuhi kebutuhan transaksi & relasi tanpa database tambahan pada MVP |
| Redis | Cache, rate limit, session sementara, queue AI, distributed lock, leaderboard cache | Menangani data cepat/berumur pendek tanpa membebani database transaksi |
| S3-compatible Object Storage | Thumbnail, background kota, avatar, aset PixiJS | File media tidak sebaiknya disimpan dalam tabel MariaDB |

**AI & Scoring**

| Komponen | Fungsi Utama | Rasionalisasi |
|---|---|---|
| External AI API / open-source model endpoint | Contextual chatbot, semantic evaluation, adaptive feedback, draft case generation | Mempercepat implementasi fitur berbasis bahasa tanpa training pipeline sendiri |
| Rule-based Scoring Engine (Go) | Penilaian pilihan jawaban, confidence calibration, evidence usage, claim classification, final decision, city impact, XP | Keputusan stabil, dapat diuji, dan dapat dijelaskan (explainable) |

> **Prinsip non-negotiable:** AI tidak menentukan skor final, outcome, atau city impact. AI hanya menghasilkan semantic signals dan feedback bahasa natural; seluruh keputusan bisnis dijalankan oleh rule di backend Go.

---

## 12. Data dan API Requirements

### 12.1 Entitas Data Utama (MariaDB)

| Entitas | Deskripsi Singkat |
|---|---|
| Users / User Profiles | Identitas, avatar, level, XP, statistik kompetensi, reputasi |
| Cities / City Statistics | Kondisi kota, indikator utama, riwayat dampak |
| Cases / Case Versions | Konten case, versi/snapshot, scoring configuration |
| Evidence | Konten evidence per case |
| Questions | Struktur pertanyaan per case beserta tipe |
| Sessions | Sesi investigasi aktif/selesai per pengguna per case |
| Answers | Draft dan jawaban final pemain per pertanyaan |
| Chat Histories | Riwayat percakapan chatbot kontekstual per sesi |
| Scores / Outcomes | Hasil scoring per dimensi, outcome case, dan dampak kota |
| Progress | XP, level, reputasi, dan case yang telah diselesaikan |
| Leaderboard Records | Data ringkas untuk keperluan ranking |

### 12.2 Data yang Dikumpulkan vs. Tidak Dikumpulkan

**Dikumpulkan (Minimum yang Diperlukan)**
- Data profil minimum: User ID, username, age range, avatar, bahasa, consent status.
- Data gameplay: case yang dipilih, waktu mulai, evidence yang dibuka & urutannya, structured answers, confidence values, claim classification, open response, chatbot messages, final decision, waktu submit.
- Data sistem: case version, application version, kategori perangkat, status jaringan, session version, timestamp, idempotency key.

**Secara Eksplisit TIDAK Dikumpulkan**
- Riwayat media sosial asli pengguna.
- Percakapan pribadi di luar game.
- Preferensi politik nyata.
- Diagnosis kesehatan pengguna.
- Data sensitif lain yang tidak dibutuhkan untuk fungsi produk.

### 12.3 Requirement REST API (Endpoint Inti)

| Method & Path | Deskripsi | Auth | Catatan Kunci |
|---|---|---|---|
| GET /dashboard | Mengambil city state, profil, progress, dan case catalog | Wajib | Sumber data City Dashboard |
| POST /cases/{caseId}/sessions | Memulai sesi baru untuk sebuah case | Wajib | Validasi prerequisite & session aktif; membuat case session snapshot |
| PUT /sessions/{id}/answers | Menyimpan/memperbarui draft jawaban | Wajib | Upsert dengan idempotency key |
| POST /sessions/{id}/events | Mencatat event interaksi (mis. evidence opened) | Wajib | Validasi evidence reference & session version |
| POST /sessions/{id}/chat | Mengirim pesan ke contextual chatbot | Wajib | Backend membangun bounded context sebelum memanggil AI; fallback bila AI gagal |
| GET /sessions/{id}/submission-summary | Mengambil ringkasan seluruh draft jawaban sebelum submit | Wajib | Menampilkan completion status & missing fields |
| POST /sessions/{id}/submit | Mengirim final decision & memicu scoring | Wajib | Lock session, validasi payload lengkap, jalankan semantic evaluation + rule-based scoring dalam satu alur transaksional |
| POST /auth/register | Registrasi akun baru | Tidak | Diusulkan — mengikuti AUTH-01 |
| POST /auth/login | Login & penerbitan token | Tidak | Diusulkan — mengikuti AUTH-02/AUTH-03 |
| POST /auth/logout | Invalidasi token/sesi | Wajib | Diusulkan — mengikuti AUTH-04 |

### 12.4 Requirement Data Pipeline
1. **Data Acquisition** — data diperoleh langsung dari interaksi pengguna (bukan sumber pihak ketiga).
2. **Client Preprocessing** — validasi field, normalisasi input, pembuatan idempotency key, penyimpanan draft di IndexedDB, sanitasi input teks, pemeriksaan syarat final submission.
3. **Transmission** — seluruh data dikirim melalui HTTPS REST API dengan authentication credential, session ID, case version, session version, payload, timestamp, dan idempotency key.
4. **Backend Processing** — backend memeriksa case masih aktif, kepemilikan session, validitas question ID & evidence, kesesuaian tipe jawaban, kelengkapan pertanyaan wajib, dan status final submit.
5. **AI Processing** — contextual chatbot dibangun dari bounded context; open response evaluation mengirim question, jawaban, rubric, dan context terbatas, lalu menerima semantic signals (bukan skor final).
6. **Analytic Processing** — scoring engine menggabungkan seluruh sinyal menjadi skor final, XP, reputasi, outcome, dan city impact.
7. **Action** — backend menyimpan final answer, score result, outcome, city impact, level, reputasi, completed case, dan feedback record dalam satu transaksi, baru kemudian frontend menampilkan hasil ke pemain.

---

## 13. Requirement AI dan Guardrail

### 13.1 Batasan Peran AI
AI digunakan secara terbatas dan tidak pernah menjadi penentu skor final atau sumber kebenaran tunggal:
- Menjalankan chatbot kontekstual di dalam case.
- Memahami jawaban terbuka pemain (semantic understanding).
- Menghasilkan semantic signals untuk dikonsumsi scoring engine.
- Menyusun feedback yang lebih natural.
- Membantu content author menghasilkan draft case terstruktur (dengan review manusia wajib).

### 13.2 Guardrail Contextual Chatbot
- Prompt dibangun dari: case briefing, allowed evidence, chatbot persona, knowledge boundary, prohibited behavior, dan conversation history — seluruhnya disusun backend.
- Respons AI divalidasi backend sebelum dikirim ke frontend.
- Apabila AI gagal atau menghasilkan output tidak valid, sistem menampilkan fallback response dan gameplay tetap dapat dilanjutkan.

### 13.3 Semantic Signals untuk Open Response
AI mengembalikan sinyal semantik berikut (bukan skor akhir): mengakui ketidakpastian, mempertimbangkan risiko, membedakan pengalaman dan bukti, menggunakan alasan yang relevan, membuat generalisasi berlebihan, memerlukan human review.

> Apabila semantic analysis gagal, backend menggunakan fallback deterministic rubric sehingga proses scoring tidak pernah terhenti akibat kegagalan layanan AI eksternal.

---

## 14. Analitik dan Instrumentasi (Diusulkan)

| Event | Dipicu Saat | Metrik yang Didukung |
|---|---|---|
| account_registered | Registrasi akun berhasil | Aktivasi, funnel awareness→registration |
| tutorial_completed | Tutorial case selesai | Tutorial Completion Rate |
| case_started | Sesi case dibuat | Time to First Case, Engagement |
| evidence_opened | Pemain membuka evidence card | Evidence Evaluation Score, pola investigasi |
| answer_saved | Draft jawaban tersimpan | Progres pengisian, deteksi drop-off |
| chat_message_sent | Pemain mengirim pesan ke chatbot | AI Fallback Rate, pola penggunaan chatbot |
| session_submitted | Final submission berhasil | Weekly Verified Cases, Case Completed/Active User |
| result_viewed | Result screen ditampilkan & dwell time diukur | Distrust Signal Rate |
| confidence_revised | Confidence akhir berbeda dari confidence awal | Revision Willingness Rate |
| level_up | Kenaikan level pemain | Retensi & progression health |
| session_resumed | Sesi dilanjutkan setelah koneksi terputus | Reliability offline/PWA |

---

## 15. Risiko, Asumsi, dan Dependensi (RAID)

| Tipe | Deskripsi | Dampak | Mitigasi |
|---|---|---|---|
| Risiko | Chatbot AI menghasilkan jawaban di luar konteks / halusinasi | Tinggi | Guardrail context boundary ketat, validasi respons, fallback response |
| Risiko | Beban kognitif berlebih | Sedang | Batasi evidence 3–5 per case; open question hanya pada keputusan penting |
| Risiko | Pemain merasa "dihakimi" oleh AI evaluator | Sedang | Pisahkan rule-based score vs AI feedback; bahasa netral |
| Risiko | Ketergantungan pada API AI eksternal | Tinggi | Fallback rubric deterministic, caching, rate limiting, evaluasi model open-source |
| Risiko | Skalabilitas MariaDB saat traffic tinggi | Sedang | Redis caching; evaluasi read replica pasca-MVP |
| Risiko | Privasi data pengguna remaja | Tinggi | Data minimization, consent eksplisit, tanpa data sensitif |
| Risiko | Konten case dianggap bias/sensitif | Sedang | Proses editorial review manusia wajib sebelum publikasi |
| Asumsi | Pengguna primer memiliki akses perangkat & internet meski tidak selalu stabil | — | Pendekatan PWA + draft lokal |
| Asumsi | Tim pengembangan awal berukuran kecil | — | Arsitektur modular monolith |
| Dependensi | Ketersediaan model AI eksternal dengan kebijakan data sesuai | Tinggi | Keputusan vendor/model sebelum Sprint AI diprioritaskan |
| Dependensi | Jadwal & kriteria kurasi UNESCO Youth Hackathon | Tinggi | Roadmap MVP diselaraskan dengan milestone kompetisi |

---

## 16. Roadmap dan Rencana Rilis (Diusulkan)

| Fase | Fokus Utama | Fitur Kunci | Kriteria Keluar |
|---|---|---|---|
| Fase 0 — Discovery & Desain | Validasi konsep, schema case, wireframe | Persona & journey map final, schema case & scoring rule terdefinisi, 1 case pilot | Schema disepakati; 1 case pilot dimainkan internal |
| Fase 1 — MVP Core Loop | Alur inti end-to-end | Auth dasar, City Dashboard minimal, 2–3 case, Investigation Screen 3 kolom, rule-based scoring dasar, Result Screen | Pemain baru dapat 1 siklus penuh tanpa AI eksternal (fallback-only) |
| Fase 2 — AI & Progression | Mengaktifkan chatbot & progres penuh | Chatbot kontekstual + guardrail, semantic evaluation, XP/level/reputasi, unlocking case | AI Fallback Rate ≤ target; minimal 5 case dengan variasi template |
| Fase 3 — Engagement & Distribusi | Retensi & adopsi awal pengguna sekunder | Leaderboard mingguan, rekomendasi berbasis kompetensi, onboarding institusi | D7 retention & institutional sign-up pertama tercapai |
| Fase 4 — Scale & Institutional Tooling | Skala konten & alat institusi | Content authoring tool AI-assisted, cohort leaderboard, dashboard kompetensi | Content author eksternal dapat menerbitkan case terverifikasi |

---

## 17. Referensi Alur Sistem

### 17.1 Ringkasan Alur (End-to-End)
1. Pemain membuka City Dashboard → frontend memvalidasi sesi → backend mengambil city state, profil, dan case catalog.
2. Pemain memilih case → backend memvalidasi ketersediaan/prasyarat → backend membuat case session snapshot dan mengirim briefing, evidence index, questions, serta session ID.
3. Pemain mengisi initial judgment & confidence → tersimpan sebagai draft dengan idempotency key.
4. Loop investigasi: pemain membuka evidence (divalidasi terhadap session version) dan menjawab pertanyaan (divalidasi terhadap question type).
5. (Opsional) Pemain menggunakan chatbot → backend membangun bounded context, memanggil AI, fallback bila gagal.
6. Pemain membuka final submission → ringkasan & missing fields ditampilkan → submit → backend mengunci sesi dan memvalidasi payload final.
7. Backend menjalankan semantic evaluation (fallback rubric bila gagal) lalu rule-based scoring → menentukan outcome, XP, reputasi, city impact → menyimpan seluruh hasil dalam satu transaksi.
8. Frontend menampilkan consequence, score breakdown, dan animasi perubahan kondisi kota → pemain kembali ke City Dashboard yang telah diperbarui.

### 17.2 Prinsip Teknis Kunci (Non-Negotiable)
- MariaDB adalah source of truth untuk case, jawaban, skor, progres, dan kondisi kota.
- Case menggunakan snapshot versi untuk menjaga konsistensi selama sesi berlangsung.
- Frontend menyimpan draft lokal agar jawaban tidak hilang.
- Seluruh final submission menggunakan idempotency key.
- AI hanya menerima context terbatas dan tidak menentukan score, outcome, atau city impact.
- Backend Go menjalankan seluruh business rule.
- Chatbot memiliki fallback jika model tidak tersedia.
- Result disimpan sebelum ditampilkan kepada pengguna.
- Cache dashboard dan leaderboard diperbarui setelah transaksi berhasil.

---

## 18. Pertanyaan Terbuka dan Keputusan yang Diperlukan
- Berapa jumlah case minimum yang perlu tersedia saat demo/launch MVP di UNESCO Youth Hackathon?
- Model/vendor AI eksternal mana yang akan digunakan, termasuk implikasi biaya dan kebijakan data terhadap pengguna remaja?
- Apakah target usia secara resmi mencakup pengguna di bawah 18 tahun, dan mekanisme consent/parental consent apa yang diperlukan?
- Apakah MVP hanya mendukung Bahasa Indonesia, atau perlu menyiapkan struktur konten multi-bahasa sejak awal?
- Bagaimana model distribusi/lisensi untuk pengguna sekunder pasca-hackathon?
- Siapa yang bertindak sebagai content author case pada fase awal, dan bagaimana proses editorial review distandardisasi?
- Apakah dibutuhkan mekanisme moderasi tambahan untuk leaderboard/cohort ranking guna mencegah perbandingan sosial yang tidak sehat?

---

## 19. Modul Tambahan: Shop, Redeem, dan Admin Console

Bagian ini menggabungkan seluruh konten dari dua dokumen PRD Addendum (`PRD_KODEKABI_Shop_Redeem_Addendum.md` dan `PRD_KODEKABI_Admin_Console_Addendum.md`) ke dalam PRD utama, dengan revisi terakhir yang sudah disepakati: role `content_author` digabung ke `admin` (dua role total: `player`, `admin`), login admin wajib 2FA (password + OTP email), CMS Case tidak lagi memiliki tahap review terpisah dari editing, skema evidence tidak lagi memuat field avatar/foto profil, dan thumbnail case bersifat wajib per-case (prompt-assisted + upload manual), sedangkan evidence image tetap opsional.

### 19.1 Shop & Redeem

### 19.1.1 Ringkasan Fitur

Shop & Redeem Center menambahkan lapisan insentif baru di luar XP, level, dan reputasi auditor yang sudah ada. Pemain mengumpulkan **Koin Auditor** dari aktivitas gameplay, lalu membelanjakannya melalui dua jalur berbeda:

1. **Shop** — membeli item kosmetik (skin avatar, skin kota, tema chat, badge) yang murni estetika, tidak memengaruhi scoring atau kesulitan case.
2. **Redeem Center** — menukarkan Koin dengan reward dunia nyata dari mitra pihak ketiga, seperti voucher Grab dan kuota data gratis.

Fitur ini tidak melibatkan transaksi uang asli (top-up/IAP) dari pemain pada fase MVP. Koin hanya diperoleh melalui progres gameplay, tidak dapat dibeli.

### 19.1.2 Tujuan dan Rasional

- Menambah alasan pemain untuk kembali bermain (retensi) di luar dorongan naik level.
- Memberi *tangible reward* di dunia nyata agar nilai literasi digital terasa lebih "menguntungkan" secara langsung, khususnya bagi persona seperti Nadia yang termotivasi oleh manfaat praktis.
- Menjaga agar kosmetik tetap terpisah dari sistem kompetensi — supaya pemain tidak salah paham bahwa membeli skin = bermain lebih baik (menghindari kesan pay-to-win, meskipun tidak ada pembayaran uang asli).

### 19.1.3 Ruang Lingkup

#### 19.1.3.1 Dalam Lingkup MVP
- Sistem saldo Koin Auditor per pengguna beserta riwayat transaksi (wallet).
- Perolehan Koin dari penyelesaian case (base reward + bonus performa).
- Katalog Shop berisi item kosmetik, dengan pembelian menggunakan Koin.
- Inventory pemain: daftar item yang dimiliki, dan mekanisme *equip/unequip*.
- Katalog Redeem Center berisi reward mitra (voucher Grab, kuota data) dengan harga dalam Koin dan stok terbatas.
- Klaim redeem: penukaran Koin dengan satu kode unik dari stok yang tersedia, prosedur pencegahan klaim ganda/race condition.
- Riwayat klaim redeem pemain.
- Pengelolaan katalog Shop dan Redeem (termasuk unggah batch kode) oleh Admin — *lihat catatan keterkaitan dengan Admin Module di Bagian 9 dokumen ini.*

#### 19.1.3.2 Di Luar Lingkup MVP (Fase Berikutnya)
- Pembelian Koin dengan uang asli (in-app purchase).
- Marketplace atau perdagangan item antar-pemain.
- Integrasi API real-time langsung ke sistem mitra (Grab, operator seluler) untuk generate kode otomatis — pada MVP kode disiapkan manual dalam bentuk batch oleh Admin.
- Penarikan Koin menjadi uang tunai dalam bentuk apa pun.
- Item Shop musiman/edisi terbatas dengan mekanisme gacha/random box.
- Verifikasi identitas lanjutan (KYC) untuk klaim reward bernilai tinggi — lihat Bagian 9 (Risiko & Pertanyaan Terbuka).

#### 19.1.3.3 Batasan (Constraints)
- Item Shop **tidak boleh** memberi keuntungan gameplay apa pun (tidak menambah XP multiplier, tidak membuka evidence lebih cepat, tidak memengaruhi scoring).
- Kode redeem bersifat **satu kode untuk satu klaim**; begitu diklaim, kode tidak dapat diklaim ulang oleh siapa pun.
- Saldo Koin tidak dapat menjadi negatif dalam kondisi apa pun.
- Seluruh mutasi saldo Koin wajib tercatat sebagai transaksi (tidak ada perubahan saldo "diam-diam").

### 19.1.4 Konsep dan Mekanisme

#### 19.1.4.1 Koin Auditor (Mata Uang Dalam Game)

Koin Auditor adalah mata uang lunak (soft currency), terpisah dari **Auditor Reputation** yang sudah ada. Perbedaannya perlu ditegaskan agar tidak rancu:

| | Auditor Reputation | Koin Auditor |
|---|---|---|
| Fungsi | Mengukur kualitas/kepercayaan sebagai auditor (kompetensi) | Mata uang yang bisa dibelanjakan |
| Sumber | Kualitas keputusan pada case | Penyelesaian case + bonus tertentu |
| Dapat dibelanjakan? | Tidak | Ya |
| Rentang nilai | 0–1000 (clamped) | Tidak terbatas atas, tidak boleh negatif |

**Sumber perolehan Koin (MVP):**
- Reward dasar setiap kali menyelesaikan case (`final submission` berhasil).
- Bonus performa: skor investigasi tinggi, evidence dibuka lengkap sebelum menjawab, atau tidak ada kekeliruan besar.
- Bonus penyelesaian case pertama kali (first-time completion, untuk mencegah farming lewat replay case yang sama — *lihat 4.1.1*).

**4.1.1 Anti-Farming**
Case yang sudah pernah diselesaikan (`completed`) dan dimainkan ulang tidak memberikan Koin, hanya XP percobaan ulang (jika ada), untuk mencegah eksploitasi ekonomi Koin melalui pengulangan case yang sama.

#### 19.1.4.2 Shop (Kosmetik)

Kategori item pada MVP:
- **Avatar skin** — tampilan alternatif avatar auditor.
- **City skin / background variant** — variasi visual Kota Nusa pada City Dashboard (murni tampilan, tidak mengubah indikator kota).
- **Chat bubble theme** — tema visual pada panel chatbot kontekstual.
- **Title / badge** — label kosmetik yang tampil di profil dan leaderboard.

Setiap item memiliki harga tetap dalam Koin, tidak fluktuatif. Item yang sudah dimiliki tidak dapat dibeli ulang (tombol berubah menjadi "Equip"/"Dimiliki").

#### 19.1.4.3 Redeem Center (Reward Dunia Nyata)

Kategori reward pada MVP:
- **Voucher Grab** (mis. kode potongan/kredit Grab).
- **Kuota data gratis** (kode voucher operator seluler).
- Kategori mitra lain dapat ditambahkan tanpa mengubah struktur data (generik per `redeem_item.type`).

Karakteristik yang membedakan Redeem dari Shop:
- **Stok terbatas** per item (bukan digital tanpa batas seperti skin).
- Menghasilkan **kode unik** yang harus dikirim ke pemain (bukan sekadar entri di inventory).
- Berpotensi disalahgunakan (multi-akun untuk memborong reward bernilai nyata) sehingga membutuhkan pembatasan tambahan (lihat 4.3.1).

**4.3.1 Pembatasan Klaim**
- Maksimum klaim per item per pemain, mis. 1x per item per periode (harian/mingguan/bulanan — dikonfigurasi per item oleh Admin).
- Sistem menampilkan sisa stok secara real-time ("Tersisa 12 dari 50") agar ekspektasi pemain terkelola.
- Redeem Center dapat mensyaratkan level/reputasi minimum agar reward tidak habis oleh akun baru yang dibuat khusus untuk farming (lihat Bagian 9, Risiko).

### 19.1.5 Dampak terhadap Persona (Nadia)

Menambah *Emotional & Functional Jobs* baru pada persona yang sudah ada di PRD utama:
- **Gains baru:** "Mendapat manfaat nyata (kuota, voucher) dari waktu yang dihabiskan belajar literasi digital."
- **Pain relievers baru:** Nadia yang punya keterbatasan kuota internet dapat memperoleh kuota gratis sebagai insentif langsung untuk terus bermain, bukan sekadar poin abstrak.

### 19.1.6 Struktur Layar (UI)

#### 19.1.6.1 Navigasi
Dua tab baru ditambahkan pada navigasi utama (sejajar dengan City Dashboard, Case History, Leaderboard, Profile): **Shop** dan **Redeem**. Saldo Koin ditampilkan persisten di header/navbar di seluruh layar utama (mirip pola Public Trust/statistik kota yang selalu terlihat).

#### 19.1.6.2 Shop Screen
- Grid item, difilter berdasarkan kategori (avatar/city/chat theme/title).
- Setiap card menampilkan: thumbnail, nama, harga Koin, status (dimiliki/belum, sedang dipakai).
- Tap item → modal detail (preview lebih besar) → tombol Beli atau Equip.

#### 19.1.6.3 Purchase Confirmation
- Ringkasan: item, harga, saldo Koin sebelum & sesudah transaksi.
- Konfirmasi wajib sebelum saldo terpotong (mencegah pembelian tidak sengaja).

#### 19.1.6.4 Inventory Screen
- Daftar seluruh item yang dimiliki pemain, dikelompokkan per kategori.
- Indikator item yang sedang aktif/di-equip.

#### 19.1.6.5 Redeem Screen
- Daftar reward mitra dengan: nama, logo mitra, harga Koin, sisa stok, batas klaim per periode.
- Item dengan stok habis ditandai "Habis" dan tombol klaim dinonaktifkan.

#### 19.1.6.6 Redeem Claim Confirmation & Kode
- Konfirmasi sebelum saldo terpotong (sama seperti Shop).
- Setelah klaim berhasil: layar menampilkan kode unik hasil penukaran beserta instruksi penggunaan (mis. cara redeem ke akun Grab).
- Kode juga tersimpan permanen di **Redeem History** agar pemain dapat melihatnya kembali kapan saja (mencegah keluhan "kode hilang").

#### 19.1.6.7 Redeem History
- Daftar seluruh klaim yang pernah dilakukan pemain: item, tanggal, kode (dapat disalin ulang), status pengiriman (jika ada proses lanjutan di luar sistem).

### 19.1.7 Functional Requirements per Modul

#### 19.1.7.1 Wallet Module
Mengelola saldo dan riwayat transaksi Koin Auditor.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| WALLET-01 | Saldo Koin per pengguna | Sistem menyimpan saldo Koin yang terikat ke akun pengguna. | Must | Saldo tidak pernah negatif; selalu bilangan bulat non-negatif. |
| WALLET-02 | Perolehan Koin dari penyelesaian case | Koin ditambahkan otomatis sebagai bagian dari transaksi `final submission` yang sama dengan pemberian XP. | Must | Penambahan Koin dan XP terjadi dalam satu transaksi atomik; tidak ada kondisi XP bertambah tanpa Koin atau sebaliknya. |
| WALLET-03 | Anti-farming replay case | Case yang sudah berstatus `completed` sebelumnya tidak memberikan Koin pada penyelesaian berikutnya. | Must | Replay case tidak mengubah saldo Koin. |
| WALLET-04 | Riwayat transaksi Koin | Pemain dapat melihat daftar transaksi (masuk/keluar) beserta sumber/tujuannya. | Should | Riwayat menampilkan minimal: tanggal, jenis (earn/spend), jumlah, deskripsi sumber (nama case/item/reward). |
| WALLET-05 | Saldo tampil persisten | Saldo Koin terkini ditampilkan di navigasi utama pada seluruh layar dalam sesi aktif. | Should | Saldo diperbarui otomatis tanpa perlu refresh manual setelah transaksi. |

#### 19.1.7.2 Shop Module
Mengelola katalog item kosmetik dan proses pembelian.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| SHOP-01 | Katalog item Shop | Menampilkan seluruh item aktif beserta kategori, harga, dan status kepemilikan pemain. | Must | Item nonaktif (di-*retire* Admin) tidak muncul di katalog namun tetap ada di inventory pemilik lama. |
| SHOP-02 | Pembelian item | Pemain dapat membeli item selama saldo Koin mencukupi dan item belum dimiliki. | Must | Saldo tidak mencukupi → error jelas, tidak memotong saldo sebagian; item sudah dimiliki → tombol beli tidak aktif. |
| SHOP-03 | Item bersifat kosmetik murni | Item Shop tidak boleh memengaruhi scoring, unlock case, atau mekanisme gameplay apa pun. | Must | Verifikasi desain: tidak ada field pada item yang terhubung ke scoring rule/unlock requirement. |
| SHOP-04 | Equip/unequip item | Pemain dapat mengaktifkan satu item per kategori sebagai tampilan aktif. | Must | Hanya satu item aktif per kategori (mis. satu avatar skin aktif dalam satu waktu). |
| SHOP-05 | Idempotensi pembelian | Permintaan pembelian yang dikirim berulang (mis. akibat retry jaringan) tidak boleh memotong saldo lebih dari satu kali untuk transaksi yang sama. | Must | Menggunakan idempotency key sesuai pola yang sudah diterapkan pada Answer/Submission module. |

#### 19.1.7.3 Redeem Module
Mengelola katalog reward mitra, stok kode, dan proses klaim.

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| REDEEM-01 | Katalog Redeem | Menampilkan reward aktif beserta harga Koin, sisa stok, dan mitra penyedia. | Must | Item dengan stok 0 tetap tampil namun berstatus "Habis" dan tidak dapat diklaim. |
| REDEEM-02 | Klaim reward | Pemain menukar Koin dengan satu kode unik dari stok yang tersedia. | Must | Saldo terpotong dan satu kode ditandai `claimed` dalam satu transaksi atomik; tidak ada kondisi Koin terpotong tanpa kode diberikan atau sebaliknya. |
| REDEEM-03 | Pencegahan klaim ganda atas kode yang sama | Satu kode hanya dapat diberikan kepada satu pemain, walau terjadi permintaan klaim bersamaan (concurrent). | Must | Pengujian concurrent request menghasilkan tepat satu pemain yang menerima kode; permintaan lain menerima error "stok habis" jika kode terakhir sudah terambil. |
| REDEEM-04 | Batas klaim per pemain per item | Sistem membatasi jumlah klaim seorang pemain terhadap item redeem yang sama dalam periode tertentu (dikonfigurasi per item). | Must | Percobaan klaim melebihi batas periode ditolak dengan pesan yang menjelaskan kapan dapat mencoba lagi. |
| REDEEM-05 | Riwayat klaim pemain | Pemain dapat melihat kembali seluruh kode yang pernah diklaim beserta tanggal dan instruksi penggunaan. | Must | Kode yang sudah diklaim tetap dapat dilihat/disalin ulang kapan pun dari riwayat. |
| REDEEM-06 | Pengelolaan stok kode oleh Admin | Admin dapat menambah batch kode baru untuk suatu item redeem. | Should | Penambahan batch tidak mengganggu proses klaim yang sedang berjalan; kode baru langsung menambah sisa stok yang tampil ke pemain. |
| REDEEM-07 | Kerahasiaan kode sebelum klaim | Kode yang belum diklaim tidak boleh terlihat oleh siapa pun melalui endpoint publik atau log aplikasi. | Must | Audit menunjukkan kode hanya muncul di response API tepat saat berhasil diklaim oleh pemiliknya, dan tidak tercatat dalam log biasa (hanya pada penyimpanan data terenkripsi/terbatas akses). |

### 19.1.8 Data yang Dibutuhkan (Ringkasan)

Mengikuti prinsip *data minimization* pada dokumen induk, data tambahan yang dikumpulkan dibatasi pada:
- Saldo dan riwayat transaksi Koin.
- Item yang dimiliki/di-equip pemain.
- Riwayat klaim redeem (item, kode, waktu).

**Catatan khusus:** untuk kategori reward yang memerlukan pengiriman ke nomor telepon (misalnya kuota data), sistem membutuhkan nomor telepon aktif pemain — data ini **belum ada** dalam skema data minimum pengguna saat ini (lihat Bagian 9, Pertanyaan Terbuka).

### 19.1.9 Risiko dan Pertanyaan Terbuka

| Topik | Risiko/Pertanyaan | Catatan |
|---|---|---|
| Multi-akun farming | Pemain dapat membuat banyak akun untuk memborong reward bernilai nyata (voucher/kuota) yang stoknya terbatas. | Perlu diputuskan: apakah cukup dibatasi dengan limit level/reputasi minimum, atau perlu verifikasi tambahan (nomor telepon/OTP) khusus untuk fitur redeem. |
| Data nomor telepon | Kuota data mensyaratkan nomor telepon tujuan, sementara skema data saat ini tidak mengumpulkan nomor telepon. | Perlu keputusan produk: kumpulkan saat onboarding (mengubah prinsip data minimum) atau hanya diminta saat pemain pertama kali klaim kategori "kuota". |
| Ketersediaan kode habis mendadak | Reward populer bisa habis dalam waktu singkat, menimbulkan kekecewaan. | Pertimbangkan notifikasi "stok menipis" atau kuota klaim harian per item agar distribusi lebih merata. |
| Kesalahan pengiriman kode oleh mitra | Kode dari mitra (Grab/operator) tidak valid atau sudah dipakai di sisi mitra. | Di luar kendali sistem KODEKABI pada MVP karena kode diunggah manual; perlu SOP verifikasi kode sebelum diunggah Admin. |
| Ekonomi Koin tidak seimbang | Jika sumber Koin terlalu deras dibanding harga item, nilai reward jadi terlalu mudah didapat (atau sebaliknya, terlalu sulit sehingga fitur terasa hampa). | Perlu tuning harga/reward pasca-MVP berdasarkan data telemetri riil, bukan diputuskan sekali di awal. |


### 19.2 Admin Console

### 19.2.1 Ringkasan Fitur

Admin Console adalah antarmuka internal terpisah dari aplikasi player, digunakan oleh satu role tunggal: **`admin`**. Role ini menggantikan pemisahan `content_author`/`admin` yang sebelumnya tersirat di SRS induk — seluruh kebutuhan produksi konten (case, shop, redeem) dan operasional (user management, moderasi, konfigurasi, audit) dikelola oleh role yang sama, dengan lapisan keamanan tambahan berupa **autentikasi dua faktor (2FA)** karena cakupan aksesnya sangat luas.

### 19.2.2 Tujuan dan Rasional

- Memberi tempat kerja resmi bagi aktivitas yang selama ini hanya tersirat sebagai *role* di RBAC (publish case, kelola shop/redeem, moderasi) tapi tidak pernah punya spesifikasi antarmuka.
- Menyederhanakan operasional di tahap awal produk (tim kecil) dengan menghilangkan hand-off dua pihak (content author → admin) yang tidak diperlukan pada skala tim saat ini.
- Menjaga integritas data yang berdampak langsung ke pemain (scoring, ekonomi Koin, kode redeem bernilai nyata) melalui validasi otomatis dan audit trail yang ketat, mengingat kesalahan di sisi admin bisa memengaruhi banyak pemain sekaligus.

### 19.2.3 Ruang Lingkup

#### 19.2.3.1 Dalam Lingkup
- Login admin dengan 2FA (password + OTP email).
- User Management: direktori pengguna, detail profil, role toggle, moderasi akun (suspend/ban/force logout), hapus akun.
- CMS Case: form builder case sesuai schema generik, AI-assisted draft generation dengan validasi otomatis, preview sandbox, publish langsung, version control & rollback, konfigurasi unlock requirement.
- CMS Shop & Redeem: CRUD item Shop, CRUD item Redeem, upload batch kode dengan validasi duplikasi, void kode, dashboard ekonomi Koin.
- Moderasi: antrean tinjauan percakapan chatbot yang di-flag dan open response yang ditandai AI "perlu human review".
- Dashboard operasional: kesehatan sistem (AI latency, fallback rate, error rate), metrik gameplay agregat.
- Konfigurasi game: tabel pemetaan visual kota, tuning formula reward.
- Tata kelola AI/chatbot: guardrail, prompt rubric, monitoring biaya & fallback.
- Leaderboard admin: reset musiman, investigasi anomali.
- Audit log terpusat, append-only.
- Pengumuman in-app.
- Reporting & export data agregat.

#### 19.2.3.2 Di Luar Lingkup (Fase Berikutnya)
- Role tambahan selain `player`/`admin` (mis. `moderator` terpisah) — dicatat sebagai catatan arsitektur untuk direvisit bila tim admin membesar.
- Integrasi API real-time ke mitra Redeem (Grab, operator seluler) untuk generate kode otomatis — kode tetap diunggah manual oleh Admin pada fase ini.
- Mekanisme anonimisasi akun saat penghapusan — penghapusan bersifat hapus penuh (hard delete) sebagai default sederhana.
- Dashboard guru/fasilitator (pengguna sekunder) sebagai produk terpisah — Admin Console hanya menyiapkan aksesnya di fase ini, belum membangun dashboard-nya sendiri.
- Fallback channel OTP selain email (mis. SMS/authenticator app).

#### 19.2.3.3 Batasan (Constraints)
- Login admin **tidak boleh** berhasil tanpa OTP, termasuk di lingkungan development/staging.
- Item Shop tetap kosmetik murni (tidak boleh memengaruhi scoring/unlock), sesuai batasan yang sudah ditetapkan di addendum Shop & Redeem.
- Seluruh mutasi data yang menyentuh pemain (role, saldo, kode redeem, konten case published) **wajib** tercatat di audit log tanpa kecuali.
- Draft case hasil AI generation **wajib** lolos validasi konsistensi ID otomatis sebelum dapat disimpan sebagai `draft` yang bisa ditinjau — tidak boleh ada draft dengan referensi ID rusak yang lolos ke tahap review manual.

### 19.2.4 Model Role dan Akses

| Role | Cakupan Akses |
|---|---|
| `player` | Akses aplikasi gameplay (tidak berubah dari dokumen induk). |
| `admin` | Akses penuh ke seluruh Admin Console. |

Perubahan ini merevisi AUTH-05 pada PRD induk ("Sistem membedakan peran player, content author, dan admin") menjadi: **Sistem membedakan peran player dan admin**, dengan `admin` membawa akses penuh ke seluruh fungsi produksi dan operasional yang sebelumnya terbagi antara `content_author` dan `admin`.

### 19.2.5 Konsep dan Mekanisme

#### 19.2.5.1 Autentikasi Admin (2FA)
Login admin membutuhkan dua langkah: password, lalu kode OTP 6 digit yang dikirim ke email terdaftar, berlaku 5 menit dan sekali pakai. Percobaan OTP salah dibatasi mengikuti pola rate-limit yang sudah diterapkan pada modul Auth pemain (SRS-AUTH-003).

#### 19.2.5.2 User Management
Admin dapat melihat direktori pengguna, membuka detail profil lengkap (termasuk data dari Wallet dan Redeem — akses ini sendiri tercatat di audit log karena menyentuh data sensitif), mengubah role (`player` ↔ `admin` dengan konfirmasi tambahan), memoderasi akun (suspend/ban/force logout), dan menghapus akun secara permanen.

#### 19.2.5.3 CMS Case

*(Revisi — alur dua fase: create case dengan metadata saja, lalu lengkapi detail di halaman terpisah dengan tab per bagian.)*

**Fase 1 — Create Case (metadata saja).** Admin mengisi metadata case (judul, tema, risk level, estimasi durasi, unlock requirement, dan thumbnail) secara manual atau dengan bantuan AI. Pada mode AI, Admin cukup memberi parameter (tema, fokus kompetensi, tingkat kesulitan) dan AI mengembalikan seluruh field metadata **plus prompt gambar untuk thumbnail** dalam satu response — bukan gambar itu sendiri (lihat 19.2.5.3.2). Admin meninjau/menyunting, lalu menekan **Create**. Sistem menyimpan `case_versions` baru berstatus `draft` dengan `evidence[]`, `questions[]`, `chatbot_config`, dan `scoring_rule`/`outcome_rules` masih kosong, lalu mengarahkan Admin ke halaman **Case Detail**.

**Fase 2 — Case Detail (hub dengan tab).** Halaman ini menampilkan ringkasan metadata (dapat disunting inline atau lewat tombol edit) dan empat tab independen di bawahnya: **Evidence**, **Questions**, **Chatbot Config**, dan **Scoring & Outcome**. Setiap tab dapat diisi manual atau lewat generate AI, dengan cakupan sebagai berikut:

- **Evidence** — generate AI menghasilkan satu set utuh 3–5 evidence sekaligus (bukan satu per satu), masing-masing memilih salah satu dari 6 `template_type` (`social_post`, `article`, `blog`, `forum_thread`, `chat_transcript`, `public_announcement`, boleh berulang tipe). Case **tidak wajib** memuat keenam tipe sekaligus. Prompt gambar untuk evidence yang membutuhkan foto (opsional) ikut disertakan langsung di evidence terkait, bukan di tempat terpisah.
- **Questions** — generate AI **membutuhkan minimal satu evidence sudah ada** (dikirim sebagai context ke AI supaya `related_evidence_ids` merujuk ID yang valid); manual add pun tetap butuh evidence untuk direferensikan.
- **Chatbot Config** — independen, dapat diisi kapan saja setelah case dibuat.
- **Scoring & Outcome** — **terkunci** (baik untuk generate AI maupun input manual) sampai tab Evidence dan Questions punya isi, karena `scoring_rule` wajib merujuk `evidence_id`/`question_id` yang benar-benar ada. Saat generate AI di tab ini, backend mengirim daftar evidence dan questions yang sudah ada sebagai context.

Setiap kali generate AI dipanggil di tab manapun, backend mengirim **system prompt tetap khusus tahap itu** (lihat Lampiran A pada SRS untuk isi lengkap tiap prompt) plus **state case saat ini** yang relevan sebagai context — bukan satu prompt tunggal untuk seluruh case seperti rancangan sebelumnya. Backend memvalidasi konsistensi ID setiap kali evidence/questions/scoring diperbarui (baik manual maupun AI), bukan cuma sekali di awal.

**Publish** hanya dapat dilakukan setelah validasi lengkap: metadata beserta thumbnail terisi, evidence 3–5 dengan minimal satu `is_critical: true`, questions lengkap dengan `related_evidence_ids` valid, chatbot config terisi, dan scoring rule mencakup kelima kategori dengan outcome rules yang menutup seluruh rentang skor tanpa celah. Kalau ada bagian yang belum lengkap, publish ditolak dan Admin diarahkan kembali ke tab yang bersangkutan.

##### 19.2.5.3.1 Skema Evidence (Revisi) dan Produksi Aset Visual

**Field umum di setiap evidence**, apa pun tipenya: `evidence_id`, `template_type`, `label` (tampil di list evidence kolom kiri), `credibility_tags` (dipakai scoring, tidak tampil ke pemain), `is_critical` (evidence wajib dibuka untuk skor evidence evaluation maksimal).

**Field khusus per `template_type`** (avatar/foto profil sudah dihapus dari skema — cukup nama teks, tanpa gambar):

| template_type | Field Khusus |
|---|---|
| `social_post` | `author_name`, `author_handle`, `platform` (nama fiktif), `post_text`, `image_slot` (opsional), `timestamp`, `engagement_count`, `is_verified_account` |
| `article` | `headline`, `source_name` (fiktif), `author_name`, `publish_date`, `body_text`, `header_image_slot` (opsional) |
| `blog` | `title`, `author_name`, `blog_name`, `publish_date`, `body_text` |
| `forum_thread` | `thread_title`, `forum_name`, `posts[]` (`author_name`, `text`, `timestamp`, `upvote_count` — tanpa avatar) |
| `chat_transcript` | `participants[]` (nama saja, tanpa avatar), `messages[]` (`sender`, `text`, `timestamp`) |
| `public_announcement` | `issuing_body` (fiktif), `title`, `body_text`, `date` |

##### 19.2.5.3.2 Aset Visual per Slot

| Slot | Wajib/Opsional | Dihasilkan Pada Tahap | Sifat |
|---|---|---|---|
| **Thumbnail case** | **Wajib** — blocking untuk publish | Fase 1 (Create Case, bagian metadata) | Unik per case, bukan reusable per kategori |
| **Evidence image** (`image_slot`, `header_image_slot`) | Opsional | Fase 2, saat generate/isi tab Evidence | Kalau kosong, tampil sebagai placeholder box, tidak blocking publish |

Untuk kedua jenis slot, AI generator menyertakan sebuah **prompt gambar siap pakai** (bukan gambar itu sendiri) langsung menyatu dengan objek yang digambarkannya — prompt thumbnail keluar bareng respons generate metadata, prompt evidence image keluar bareng evidence yang bersangkutan saat generate tab Evidence — dibangun dari satu *style anchor* tetap (gaya visual, palet warna, mood — sama di semua case demi konsistensi identitas Kota Nusa) digabung deskripsi spesifik. Admin memakai prompt ini di tools image-generation eksternal pilihan sendiri, lalu mengunggah hasilnya secara manual ke slot yang sesuai. Prompt yang dipakai tetap disimpan bersama file yang diunggah, supaya dapat dipakai ulang atau disunting kalau Admin ingin regenerate.

#### 19.2.5.4 CMS Shop & Redeem
Admin mengelola katalog item Shop (kosmetik) dan item Redeem (reward mitra dunia nyata), termasuk mengunggah batch kode redeem dengan validasi duplikasi, serta memantau ekonomi Koin (sumber vs. penggunaan) untuk mendeteksi ketidakseimbangan sejak dini.

#### 19.2.5.5 Moderasi Konten
Percakapan chatbot yang di-flag guardrail dan jawaban open question yang ditandai AI sebagai "perlu human review" (semantic signal yang sudah dihasilkan sistem sejak PRD induk Bagian 3.5.5, namun belum punya tempat ditindaklanjuti) masuk ke antrean moderasi dengan konteks lengkap.

#### 19.2.5.6 Dashboard Operasional
Menampilkan kesehatan sistem (latency AI API, fallback rate chatbot, error rate submission) dan metrik gameplay agregat (funnel drop-off, distribusi skor kompetensi, case dengan completion rate rendah).

#### 19.2.5.7 Konfigurasi Game
Tabel pemetaan visual kota (yang menurut SRS-CITY-003 harus dapat dikonfigurasi, bukan hardcode) dan formula reward XP/Koin dikelola lewat antarmuka, dengan staging sebelum berlaku ke production.

#### 19.2.5.8 Tata Kelola AI/Chatbot
Guardrail, prompt rubric semantic evaluation, dan monitoring biaya/fallback AI dikelola terpusat, terpisah dari editor case biasa karena menyangkut keamanan konten lintas case.

#### 19.2.5.9 Leaderboard Admin
Reset musiman dan investigasi skor anomali (indikasi eksploitasi), dengan penyesuaian manual yang mewajibkan alasan tertulis dan tercatat di audit log.

#### 19.2.5.10 Audit Log
Satu tempat terpusat, append-only, mencatat seluruh aksi sensitif admin (publish/reject case, ubah role, suspend/ban/hapus akun, ubah harga/stok, lihat kode redeem milik pengguna, ubah konfigurasi scoring), dapat difilter berdasarkan aktor/jenis aksi/tanggal.

#### 19.2.5.11 Pengumuman In-App
Admin dapat mengirim pengumuman (case baru rilis, maintenance terjadwal, event Redeem terbatas) ke seluruh atau segmen pemain tertentu.

#### 19.2.5.12 Reporting & Export
Ekspor data agregat/anonim untuk keperluan riset atau laporan institusi, tidak memuat identitas individual di luar yang sudah publik (mis. username leaderboard).

### 19.2.6 Struktur Layar (UI)

#### 19.2.6.1 Login Screen
- Form password → form OTP terpisah (dua langkah, bukan satu form gabungan) agar jelas tahapannya bagi admin.
- Indikator waktu tersisa sebelum OTP kedaluwarsa.

#### 19.2.6.2 Navigasi Utama
```
Dashboard | Users | Case CMS | Shop & Redeem CMS | Moderation | Game Config | Leaderboard | Audit Log | Reports
```

#### 19.2.6.3 Dashboard
Ringkasan operasional (kesehatan sistem) dan ekonomi (tren Koin) dalam satu layar overview.

#### 19.2.6.4 Users
- Direktori dengan pencarian & filter.
- Detail profil bertab (Overview, Riwayat Case, Inventory & Redeem, Wallet, Consent & Privasi, Aksi Admin).

#### 19.2.6.5 Case CMS
- **Katalog Case** (semua status: draft/published/archived).
- **Create Case** — form metadata saja (manual atau AI-assisted, termasuk prompt thumbnail), diakhiri tombol Create yang membuat case draft kosong.
- **Case Detail** — ringkasan metadata (editable inline/lewat tombol edit) di atas, empat tab di bawahnya: Evidence, Questions, Chatbot Config, Scoring & Outcome. Tab Scoring & Outcome tampil terkunci (dengan indikator alasan) sampai Evidence dan Questions terisi. Tiap tab punya opsi manual add atau generate AI, dengan slot upload aset visual menyatu langsung di form evidence yang membutuhkannya — bukan panel terpisah.
- **Preview Sandbox** (mode mainkan sebagai player), tersedia begitu seluruh tab terisi.
- **Riwayat Versi** per case.

#### 19.2.6.6 Shop & Redeem CMS
- Item Shop (grid/table CRUD).
- Item Redeem (grid/table CRUD).
- Batch Kode (upload, riwayat batch, void kode).
- Economy Dashboard (grafik sumber vs. penggunaan Koin).

#### 19.2.6.7 Moderation
- Antrean chatbot flagged dan open response flagged, dengan tampilan konteks penuh (case, pemain, isi percakapan/jawaban).

#### 19.2.6.8 Game Config
- Editor tabel pemetaan visual kota.
- Editor formula reward XP/Koin dengan mode staging/preview.

#### 19.2.6.9 Leaderboard
- Kontrol reset musiman, daftar skor anomali yang terdeteksi, form penyesuaian manual (wajib alasan).

#### 19.2.6.10 Audit Log
- Tabel log dengan filter aktor/jenis aksi/rentang tanggal, tidak dapat diedit/dihapus dari UI manapun.

#### 19.2.6.11 Reports
- Form pemilihan jenis laporan dan rentang data, tombol export (CSV/JSON agregat).

### 19.2.7 Functional Requirements per Modul

#### 19.2.7.1 Auth Admin Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-AUTH-01 | Login dua langkah | Login admin membutuhkan password lalu OTP email. | Must | Login tidak berhasil hanya dengan password; berlaku di semua environment termasuk staging. |
| ADMIN-AUTH-02 | OTP sekali pakai & kedaluwarsa | Kode OTP berlaku 5 menit, sekali pakai. | Must | OTP yang sudah dipakai/kedaluwarsa ditolak. |
| ADMIN-AUTH-03 | Rate limit percobaan OTP | Percobaan OTP salah berulang memicu blokir sementara. | Must | Maksimum 5 percobaan salah sebelum diblokir 15 menit, konsisten dengan pola SRS-AUTH-003. |
| ADMIN-AUTH-04 | Masa berlaku token admin lebih pendek | Token sesi admin memiliki TTL lebih pendek dibanding sesi player. | Should | Token admin kedaluwarsa memaksa re-autentikasi penuh (password + OTP), bukan sekadar refresh. |

#### 19.2.7.2 User Management Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-USER-01 | Direktori & pencarian pengguna | Pencarian dan filter pengguna oleh Admin. | Must | Hasil pencarian <2 detik untuk basis data skala menengah. |
| ADMIN-USER-02 | Detail profil lengkap | Menampilkan overview, riwayat case, inventory, wallet, consent per pengguna. | Must | Data konsisten dengan yang dilihat pengguna di aplikasinya sendiri. |
| ADMIN-USER-03 | Toggle role player ↔ admin | Admin dapat menaikkan/menurunkan role pengguna. | Must | Perubahan tercatat di audit log; menaikkan ke admin memerlukan konfirmasi tambahan. |
| ADMIN-USER-04 | Suspend/ban akun | Admin dapat menangguhkan atau memblokir akun. | Must | Akun tersuspend/ban tidak dapat login; token aktif invalid. |
| ADMIN-USER-05 | Force logout | Admin dapat memaksa logout seluruh sesi aktif pengguna. | Should | Seluruh token masuk denylist dalam <5 detik. |
| ADMIN-USER-06 | Hapus akun permanen | Admin dapat menghapus akun secara penuh. | Should | Data akun terhapus; data agregat historis tidak dihitung ulang. |

#### 19.2.7.3 Case CMS Module

*(Revisi — requirement disusun ulang mengikuti alur dua fase: create metadata, lalu lengkapi detail per tab.)*

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-CASE-01 | Create case dengan metadata saja | Admin membuat case baru dengan mengisi metadata (manual/AI), tanpa perlu mengisi evidence/questions/scoring di langkah yang sama. | Must | Klik Create menghasilkan `case_versions` baru berstatus `draft` dengan `evidence[]`, `questions[]`, `chatbot_config`, `scoring_rule` kosong. |
| ADMIN-CASE-02 | AI-assisted metadata generation | Admin mengisi parameter (tema, fokus kompetensi, kesulitan), sistem generate seluruh field metadata plus prompt thumbnail. | Should | Response generate menyertakan prompt thumbnail; Admin dapat menyunting sebelum menekan Create. |
| ADMIN-CASE-03 | Halaman Case Detail dengan tab per bagian | Setelah create, Admin diarahkan ke halaman detail berisi ringkasan metadata dan tab Evidence/Questions/Chatbot Config/Scoring & Outcome. | Must | Seluruh tab dapat diakses dari satu halaman tanpa berpindah alamat/case ID. |
| ADMIN-CASE-04 | Generate AI tab Evidence (satu set) | Generate AI di tab Evidence menghasilkan 3–5 evidence sekaligus dalam satu response, bukan satu per satu. | Should | Response berisi array evidence lengkap dengan field sesuai `template_type` masing-masing, minimal satu `is_critical: true`. |
| ADMIN-CASE-05 | Generate/isi tab Questions membutuhkan evidence | Generate AI maupun manual add di tab Questions mensyaratkan minimal satu evidence sudah ada. | Must | Percobaan mengisi Questions saat evidence masih kosong ditolak dengan pesan yang mengarahkan ke tab Evidence. |
| ADMIN-CASE-06 | Tab Scoring & Outcome terkunci | Tab Scoring & Outcome (manual maupun AI) tidak dapat diakses sampai Evidence dan Questions memiliki isi. | Must | UI menampilkan status terkunci dengan alasan; percobaan akses lewat API langsung juga ditolak backend. |
| ADMIN-CASE-07 | Context-aware generation per tab | Generate AI di tab Questions dan Scoring & Outcome mengirim state case saat ini (evidence/questions yang sudah ada) sebagai context ke AI. | Must | Question/scoring hasil generate hanya merujuk ID yang benar-benar ada di case tersebut. |
| ADMIN-CASE-08 | Validasi konsistensi ID berkelanjutan | Sistem memvalidasi konsistensi referensi ID setiap kali evidence/questions/scoring diperbarui (manual maupun AI), bukan hanya sekali di awal. | Must | Perubahan yang merusak referensi (mis. menghapus evidence yang sudah dirujuk question) ditolak atau memicu peringatan eksplisit. |
| ADMIN-CASE-09 | Validasi publish menyeluruh | Publish memvalidasi kelengkapan seluruh bagian: metadata+thumbnail, evidence 3–5 dengan ≥1 kritis, questions valid, chatbot config terisi, scoring & outcome lengkap tanpa celah skor. | Must | Publish ditolak dengan daftar bagian yang belum lengkap; Admin diarahkan ke tab yang bersangkutan. |
| ADMIN-CASE-10 | Preview sandbox | Admin dapat memainkan case draft end-to-end tanpa memengaruhi data pemain nyata, tersedia setelah seluruh tab terisi. | Must | Sesi sandbox tidak menghasilkan entri pada tabel gameplay milik pemain nyata mana pun. |
| ADMIN-CASE-11 | Publish langsung tanpa approval terpisah | Admin mem-publish case tanpa tahap approval pihak lain. | Must | Status berubah `draft` → `published` dalam satu aksi setelah validasi ADMIN-CASE-09 lolos. |
| ADMIN-CASE-12 | Rollback versi published | Admin dapat menonaktifkan versi bermasalah. | Must | Sesi pemain aktif pada versi lama tidak terganggu (snapshot tetap berlaku). |
| ADMIN-CASE-13 | Konfigurasi unlock requirement & risk level | Admin mengatur syarat buka case dan tingkat risiko sebagai bagian metadata. | Must | Perubahan langsung tercermin di Case Card pemain setelah publish. |
| ADMIN-CASE-14 | Aktif/nonaktifkan case dari katalog | Admin dapat menyembunyikan case tanpa menghapus datanya. | Should | Case nonaktif tidak tampil di katalog publik; riwayat pemain yang sudah menyelesaikan tetap utuh. |
| ADMIN-CASE-15 | Prompt visual menyatu dengan objeknya | Prompt thumbnail keluar bersama generate metadata; prompt evidence image keluar bersama evidence yang bersangkutan saat generate tab Evidence — bukan array terpisah. | Must | Setiap respons generate yang menghasilkan objek berfoto menyertakan field prompt tepat di objek itu. |
| ADMIN-CASE-16 | Upload manual aset visual ke slot | Admin dapat mengunggah gambar hasil dari tools eksternal ke slot thumbnail/evidence image, dengan prompt yang dipakai tetap tersimpan untuk referensi ulang. | Must | File yang diunggah tersimpan ke Object Storage dan tertaut ke slot yang benar; prompt terkait dapat dilihat kembali kapan pun. |

#### 19.2.7.4 Shop & Redeem CMS Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-SHOP-01 | CRUD item Shop | Admin mengelola katalog item kosmetik. | Must | Item baru tampil di katalog player setelah status `active`. |
| ADMIN-REDEEM-01 | CRUD item Redeem | Admin mengelola katalog reward mitra. | Must | Perubahan harga/limit langsung berlaku untuk klaim berikutnya. |
| ADMIN-REDEEM-02 | Upload batch kode dengan validasi duplikasi | Admin mengunggah kode baru per item redeem. | Must | Batch dengan duplikasi ditolak seluruhnya, disertai daftar kode bentrok. |
| ADMIN-REDEEM-03 | Void kode belum terklaim | Admin dapat membatalkan kode yang belum diklaim. | Should | Kode `voided` tidak dapat diklaim; kode `claimed` tidak terpengaruh. |
| ADMIN-ECON-01 | Dashboard ekonomi Koin | Menampilkan tren sumber vs. penggunaan Koin. | Could | Data dapat difilter per periode (mingguan/bulanan). |

#### 19.2.7.5 Moderation Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-MOD-01 | Antrean moderasi chatbot & open response | Item yang di-flag sistem otomatis masuk antrean tinjauan. | Must | Setiap item menampilkan konteks lengkap (case, pemain, isi). |
| ADMIN-MOD-02 | Tindak lanjut hasil moderasi | Admin dapat menandai item sebagai selesai ditinjau, dengan catatan tindakan. | Should | Status item berubah `pending` → `reviewed`, tercatat di audit log. |

#### 19.2.7.6 Operational Dashboard Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-OPS-01 | Dashboard kesehatan sistem | Menampilkan latency AI, fallback rate, error rate submission. | Should | Data near real-time (lag maksimum beberapa menit dapat diterima). |
| ADMIN-OPS-02 | Metrik gameplay agregat | Menampilkan funnel drop-off dan distribusi skor kompetensi. | Could | Dapat difilter per rentang tanggal. |

#### 19.2.7.7 Game Config Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-CONFIG-01 | Editor tabel pemetaan visual kota | Admin mengubah mapping status indikator kota ke aset visual. | Should | Perubahan langsung berlaku tanpa deployment ulang. |
| ADMIN-CONFIG-02 | Editor formula reward dengan staging | Admin menguji perubahan formula XP/Koin sebelum berlaku ke production. | Could | Perubahan hanya berlaku ke production setelah konfirmasi eksplisit terpisah dari mode staging. |

#### 19.2.7.8 AI Governance Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-AI-01 | Kelola guardrail & prompt rubric | Admin mengatur batasan perilaku chatbot dan rubric evaluasi semantik. | Should | Perubahan tervalidasi sebelum berlaku ke case published. |
| ADMIN-AI-02 | Monitoring biaya & fallback AI | Menampilkan penggunaan token/biaya API dan tingkat fallback. | Could | Data dapat difilter per periode. |

#### 19.2.7.9 Leaderboard Admin Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-LB-01 | Reset leaderboard musiman | Admin mereset peringkat pada awal periode baru. | Could | Reset tercatat di audit log dengan timestamp. |
| ADMIN-LB-02 | Investigasi & penyesuaian skor anomali | Admin dapat menyesuaikan skor pemain secara manual dengan alasan wajib. | Could | Penyesuaian tercatat lengkap (nilai lama, baru, alasan, aktor) di audit log. |

#### 19.2.7.10 Audit Log Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-AUDIT-01 | Pencatatan otomatis seluruh aksi sensitif | Setiap mutasi data oleh Admin tercatat tanpa perlu aksi tambahan dari Admin. | Must | Tidak ada jalur mutasi data admin yang melewati pencatatan audit log. |
| ADMIN-AUDIT-02 | Log bersifat append-only | Log tidak dapat diedit/dihapus oleh admin manapun lewat antarmuka apa pun. | Must | Tidak ada endpoint update/delete pada tabel audit log. |
| ADMIN-AUDIT-03 | Filter & pencarian log | Admin dapat memfilter log berdasarkan aktor, jenis aksi, rentang tanggal. | Should | Hasil filter akurat dan responsif untuk volume log skala menengah. |

#### 19.2.7.11 Announcement Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-NOTIF-01 | Kirim pengumuman in-app | Admin membuat pengumuman yang tampil ke pemain. | Could | Pengumuman tampil sesuai target segmen dan jadwal yang ditentukan. |

#### 19.2.7.12 Reporting Module

| ID | Requirement | Deskripsi | Prioritas | Acceptance Criteria |
|---|---|---|---|---|
| ADMIN-REPORT-01 | Ekspor data agregat | Admin dapat mengekspor data agregat/anonim. | Could | Data ekspor tidak memuat identitas individual di luar yang sudah publik. |

### 19.2.8 Dampak terhadap Dokumen yang Sudah Ada

| Dokumen | Perubahan yang Diperlukan |
|---|---|
| PRD induk, AUTH-05 | Direvisi dari "player, content author, admin" menjadi "player, admin". |
| SRS induk, 2.3 (Kelas Pengguna) | Baris "Content Author" dihapus; baris "Admin" diperluas cakupannya. |
| SRS induk, SRS-AUTH-006 | RBAC direvisi dari tiga role menjadi dua role. |
| SRS induk, UC-12 | Digantikan oleh UC-18 pada SRS Addendum ini (aktor tunggal Admin, tanpa tahap approval terpisah). |
| SRS induk, kolom `role` pada data dictionary | Enum direvisi dari `player`\|`content_author`\|`admin` menjadi `player`\|`admin`. |
| SRS Addendum Shop & Redeem, UC-16 | Digantikan oleh UC-20 pada SRS Addendum ini (redaksi yang sama, aktor tetap Admin — tidak ada perubahan substansi karena UC-16 sejak awal sudah beraktor Admin). |

### 19.2.9 Risiko dan Pertanyaan Terbuka

| Topik | Risiko/Pertanyaan | Catatan |
|---|---|---|
| Konsentrasi akses pada satu role | Role `admin` tunggal membawa akses sangat luas (data pengguna, ekonomi Koin, kode redeem, konfigurasi scoring). Risiko meningkat bila human error atau kredensial admin bocor. | Dimitigasi oleh 2FA dan audit log wajib pada MVP. Perlu dicatat sebagai catatan arsitektur untuk direvisit (mis. role `moderator` terpisah) bila tim admin membesar. |
| Fallback pengiriman OTP | Belum ada mekanisme cadangan jika email OTP tidak terkirim/terlambat. | Untuk MVP, email dianggap channel yang cukup andal; perlu dipantau tingkat kegagalan pengiriman di produksi. |
| Kualitas draft AI-generated | Tanggung jawab kualitas konten sepenuhnya berada di Admin sebagai satu-satunya yang membuat dan menyunting, tanpa pihak kedua yang mengecek ulang. | Validasi otomatis (ADMIN-CASE-03) menjaga konsistensi struktural (ID, field-set), namun kualitas isi (akurasi, bias, kesesuaian tema) tetap sepenuhnya bergantung kedisiplinan Admin saat menyunting draft, tanpa lapisan review formal tambahan. |


---

## 20. Lampiran

### 20.1 Glosarium

| Istilah | Definisi |
|---|---|
| Auditor Digital | Peran yang dimainkan pengguna dalam game |
| Kota Nusa | Kota virtual yang menjadi latar dan representasi kondisi agregat |
| Case | Satu unit investigasi yang harus diselesaikan pemain |
| Evidence | Potongan informasi yang dapat dibuka pemain selama investigasi |
| Gameplay Shell | Struktur antarmuka & logika permainan yang dapat digunakan ulang |
| Confidence Calibration | Kesesuaian keyakinan pemain dengan kualitas bukti sebenarnya |
| Semantic Signals | Sinyal kualitatif hasil analisis AI atas jawaban terbuka |
| City Impact | Perubahan statistik Kota Nusa akibat keputusan pemain |
| Idempotency Key | Identifier unik pada request mutasi untuk mencegah duplikasi |
| Session Snapshot | Salinan versi case yang dikunci saat sesi dimulai |

### 20.2 Sumber Dokumen
PRD ini disusun berdasarkan Dokumen Konsep Produk "KODEKABI: Jejak Algoritma" untuk konteks UNESCO Youth Hackathon, mencakup Concept Overview, Value Proposition Canvas, Arsitektur Teknis & Pipeline Sistem, User Journey Map, serta Activity dan Sequence Diagram alur "Penyelesaian Satu Case KODEKABI".
