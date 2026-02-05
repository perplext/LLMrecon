"""
Trading Protocol Security Checks for JATS-NextG
================================================

Pemeriksaan keamanan protokol perdagangan yang digunakan oleh JATS-NextG,
termasuk FIX Protocol, ITCH/OUCH, dan proprietary messaging protocol.

Domain Audit:
- PROTO-01: FIX Protocol message validation & integrity
- PROTO-02: Session layer security & sequence number management
- PROTO-03: Heartbeat mechanism & connection monitoring
- PROTO-04: Order routing protocol security
- PROTO-05: Market data feed integrity
- PROTO-06: FIX message injection & manipulation detection
- PROTO-07: Protocol fuzzing resistance
- PROTO-08: Message replay attack protection
"""

from dataclasses import dataclass, field
from typing import List, Dict
from .network_security import Finding, Severity, CheckStatus


def run_protocol_security_checks(config: Dict) -> List[Finding]:
    """
    Menjalankan pemeriksaan keamanan protokol perdagangan JATS-NextG.
    """
    findings = []

    # =========================================================================
    # PROTO-01: FIX Protocol Message Validation & Integrity
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-01",
        title="Validasi & Integritas Pesan FIX Protocol",
        description=(
            "Verifikasi bahwa seluruh pesan FIX Protocol divalidasi secara "
            "ketat sebelum diproses oleh matching engine. Termasuk validasi "
            "field mandatory, format data, range checking, dan checksum "
            "integrity. FIX Protocol adalah standar komunikasi utama untuk "
            "order entry dan execution pada JATS-NextG."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="FIX Engine - Message Parser & Validator",
        evidence=(
            "Simulasi pengujian protokol FIX mendeteksi:\n"
            "1. FIX message parser tidak melakukan deep validation pada "
            "   tag value - hanya mengecek tag existence\n"
            "2. Negative order quantity (tag 38) tidak ditolak pada level "
            "   protocol - hanya dicek di business logic layer\n"
            "3. Oversized message (>10MB) tidak ditolak di parser level - "
            "   berpotensi buffer overflow atau memory exhaustion\n"
            "4. Custom tag (5000+) tidak divalidasi - bisa digunakan untuk "
            "   data injection\n"
            "5. FIX checksum (tag 10) hanya menggunakan modulo 256 - "
            "   mudah dipalsukan (collision attack)\n"
            "6. Tidak ada digital signature pada pesan FIX untuk "
            "   non-repudiation"
        ),
        recommendation=(
            "1. Implementasi strict FIX message validation:\n"
            "   - Validasi tipe data untuk setiap tag\n"
            "   - Range checking pada numeric fields (price, quantity)\n"
            "   - Length limit pada string fields\n"
            "   - Whitelist allowed custom tags\n"
            "2. Tambahkan maximum message size limit (1MB) pada parser\n"
            "3. Implementasi HMAC-SHA256 pada message integrity sebagai "
            "   pengganti/tambahan checksum FIX standar\n"
            "4. Tambahkan digital signature (XMLDSIG atau custom) untuk "
            "   non-repudiation pada order-critical messages\n"
            "5. Terapkan FIX Protocol conformance testing secara berkala"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.14.1.2", "OJK POJK 38/2016 Pasal 27",
                         "FIX Protocol Best Practices", "NIST SP 800-53 SI-10"],
        risk_score=9.5
    ))

    # =========================================================================
    # PROTO-02: Session Layer Security & Sequence Number
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-02",
        title="Keamanan Session Layer & Manajemen Sequence Number",
        description=(
            "Audit mekanisme session management pada FIX Protocol termasuk "
            "Logon/Logout handling, sequence number management, gap fill "
            "processing, dan session reset security."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.VULNERABLE,
        affected_component="FIX Session Manager",
        evidence=(
            "Simulasi audit session management:\n"
            "1. FIX Logon message (MsgType=A) mengirimkan password dalam "
            "   bentuk plaintext pada tag 554 - meskipun dalam TLS tunnel, "
            "   ini berisiko jika TLS terminated di load balancer\n"
            "2. Sequence number reset (tag 141=Y) dapat dilakukan tanpa "
            "   re-authentication - berpotensi disalahgunakan\n"
            "3. Gap fill request (ResendRequest MsgType=2) tidak memiliki "
            "   rate limiting - bisa digunakan untuk DoS\n"
            "4. Session timeout terlalu lama (30 menit) - sesi idle "
            "   rentan terhadap session hijacking\n"
            "5. Tidak ada binding antara FIX session ID dan TLS "
            "   certificate - session bisa di-migrate ke koneksi lain\n"
            "6. Concurrent session dari CompID yang sama tidak diblokir"
        ),
        recommendation=(
            "1. Implementasi SRP (Secure Remote Password) atau challenge-"
            "   response untuk FIX authentication, menggantikan plaintext "
            "   password\n"
            "2. Wajibkan re-authentication sebelum sequence number reset\n"
            "3. Rate limit gap fill request (max 3/menit per session)\n"
            "4. Kurangi session timeout menjadi 5 menit untuk idle session\n"
            "5. Bind FIX session ke TLS client certificate fingerprint\n"
            "6. Blokir concurrent session dari CompID yang sama\n"
            "7. Implementasi session anomaly detection"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.9.4.2", "FIX Session Protocol Spec",
                         "OJK POJK 38/2016 Pasal 28", "NIST SP 800-53 SC-23"],
        risk_score=8.7
    ))

    # =========================================================================
    # PROTO-03: Heartbeat Mechanism & Connection Monitoring
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-03",
        title="Mekanisme Heartbeat & Monitoring Koneksi",
        description=(
            "Evaluasi mekanisme heartbeat FIX Protocol dan monitoring "
            "koneksi untuk mendeteksi disconnection, connection hijacking, "
            "dan man-in-the-middle attack pada koneksi trading."
        ),
        severity=Severity.MEDIUM,
        status=CheckStatus.WARNING,
        affected_component="FIX Heartbeat & Connection Monitor",
        evidence=(
            "Simulasi evaluasi heartbeat:\n"
            "1. Heartbeat interval 30 detik terlalu lama - hijack window "
            "   besar antara heartbeat\n"
            "2. Heartbeat message (MsgType=0) tidak mengandung "
            "   cryptographic proof-of-liveness - hanya TestReqID\n"
            "3. Tidak ada server-initiated heartbeat challenge - hanya "
            "   client-initiated TestRequest\n"
            "4. Connection latency spike tidak di-monitor secara real-time "
            "   - bisa mengindikasikan MITM\n"
            "5. Heartbeat response time threshold tidak dikonfigurasi - "
            "   tidak ada automated disconnect pada anomaly"
        ),
        recommendation=(
            "1. Kurangi heartbeat interval menjadi 5-10 detik\n"
            "2. Tambahkan cryptographic nonce pada heartbeat untuk "
            "   proof-of-liveness\n"
            "3. Implementasi server-initiated heartbeat challenge\n"
            "4. Monitor connection latency dan alert pada spike >10ms\n"
            "5. Konfigurasi automated disconnect jika heartbeat response "
            "   time >3x normal average\n"
            "6. Log seluruh heartbeat anomaly untuk forensic analysis"
        ),
        remediation_effort="Medium",
        compliance_refs=["FIX Protocol Best Practices",
                         "OJK POJK 38/2016 Pasal 29"],
        risk_score=6.3
    ))

    # =========================================================================
    # PROTO-04: Order Routing Protocol Security
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-04",
        title="Keamanan Protokol Routing Order",
        description=(
            "Pemeriksaan keamanan order routing dari broker melalui RT "
            "gateway ke matching engine, termasuk order validation, "
            "routing logic security, dan order manipulation prevention."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="Order Routing Engine",
        evidence=(
            "Simulasi audit order routing:\n"
            "1. Order routing tidak memvalidasi apakah broker memiliki "
            "   otorisasi untuk efek tertentu - hanya dicek di business layer\n"
            "2. Cross-order manipulation: order cancel/replace (MsgType=G) "
            "   bisa dikirim untuk order milik broker lain jika ClOrdID "
            "   diketahui (insecure direct object reference)\n"
            "3. Order timestamp tidak di-validate di routing layer - "
            "   stale order bisa disubmit\n"
            "4. Tidak ada throttling per-broker per-instrument pada "
            "   routing level\n"
            "5. Order routing decision log tidak immutable - bisa dimodif "
            "   post-hoc menghambat audit trail"
        ),
        recommendation=(
            "1. Implementasi authorization check di routing layer - "
            "   sebelum order mencapai matching engine\n"
            "2. Bind order ID ke broker session - order cancel/replace "
            "   hanya bisa dilakukan oleh originator\n"
            "3. Validasi order timestamp - tolak order >5 detik stale\n"
            "4. Terapkan per-broker per-instrument throttling\n"
            "5. Gunakan append-only write-ahead log untuk routing decisions "
            "   dengan cryptographic hash chain\n"
            "6. Implementasi order flow monitoring untuk mendeteksi "
            "   unusual routing patterns"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.14.1.3", "OJK POJK 38/2016 Pasal 30",
                         "NIST SP 800-53 AC-4"],
        risk_score=9.0
    ))

    # =========================================================================
    # PROTO-05: Market Data Feed Integrity
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-05",
        title="Integritas Market Data Feed",
        description=(
            "Verifikasi integritas data feed harga dan perdagangan yang "
            "didistribusikan oleh JATS-NextG ke broker dan information "
            "vendor, mencegah market data manipulation."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Market Data Distribution System",
        evidence=(
            "Simulasi audit market data feed:\n"
            "1. Market data multicast tidak menggunakan authentication - "
            "   penerima tidak bisa memverifikasi asal data\n"
            "2. Tidak ada sequence numbering pada market data packets - "
            "   gap/duplicate tidak terdeteksi oleh receiver\n"
            "3. Latency antara trade execution dan market data publication "
            "   tidak di-monitor - bisa dieksploitasi untuk latency "
            "   arbitrage jika ada information leak\n"
            "4. Market data feed backup path tidak terenkripsi\n"
            "5. Tidak ada market data integrity check (hash) per-message"
        ),
        recommendation=(
            "1. Implementasi source authentication pada market data feed "
            "   (HMAC per-message atau signed batches)\n"
            "2. Tambahkan sequence numbering dan gap detection\n"
            "3. Monitor dan alert pada latency anomaly antara execution "
            "   dan market data publication\n"
            "4. Enkripsi backup market data path\n"
            "5. Tambahkan per-message integrity hash\n"
            "6. Implementasi market data anomaly detection untuk mendeteksi "
            "   potential tampering"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.14.1.2", "OJK POJK 38/2016 Pasal 31",
                         "IOSCO Principles"],
        risk_score=7.6
    ))

    # =========================================================================
    # PROTO-06: FIX Message Injection & Manipulation Detection
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-06",
        title="Deteksi Injeksi & Manipulasi Pesan FIX",
        description=(
            "Evaluasi kemampuan sistem dalam mendeteksi dan mencegah "
            "injeksi pesan FIX palsu dan manipulasi pesan FIX yang sah "
            "pada jalur komunikasi."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="FIX Message Security Layer",
        evidence=(
            "Simulasi pengujian injeksi FIX:\n"
            "1. Tidak ada message authentication code (MAC) pada pesan "
            "   FIX individual - bergantung sepenuhnya pada transport-level "
            "   TLS yang terminasi di gateway\n"
            "2. FIX tag delimiter injection (SOH character 0x01) dalam "
            "   free-text field (tag 58 Text) tidak di-sanitize\n"
            "3. BeginString (tag 8) dan BodyLength (tag 9) tidak "
            "   diverifikasi secara konsisten oleh semua komponen\n"
            "4. SenderCompID (tag 49) spoofing dimungkinkan jika "
            "   FIX session compromised - tidak ada secondary verification\n"
            "5. Tidak ada anomaly detection pada FIX message pattern - "
            "   burst of cancel requests tidak terdeteksi"
        ),
        recommendation=(
            "1. Implementasi per-message HMAC-SHA256 menggunakan shared "
            "   secret per-broker sebagai application-level integrity\n"
            "2. Sanitize SOH dan control characters pada semua free-text "
            "   FIX fields\n"
            "3. Strict validation pada BeginString dan BodyLength di "
            "   setiap processing layer\n"
            "4. Implementasi secondary CompID verification (certificate "
            "   binding atau IP binding)\n"
            "5. Deploy FIX message anomaly detection:\n"
            "   - Abnormal cancel-to-order ratio\n"
            "   - Unusual message frequency\n"
            "   - Out-of-hours message detection\n"
            "   - Cross-validation dengan broker trading pattern"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.14.1.2", "OJK POJK 38/2016 Pasal 32",
                         "NIST SP 800-53 SI-10", "FIX Protocol Security"],
        risk_score=9.2
    ))

    # =========================================================================
    # PROTO-07: Protocol Fuzzing Resistance
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-07",
        title="Ketahanan Terhadap Protocol Fuzzing",
        description=(
            "Evaluasi ketahanan FIX engine dan komponen protocol handler "
            "terhadap malformed input dan protocol fuzzing, yang dapat "
            "menyebabkan crash, memory corruption, atau undefined behavior."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="FIX Engine - Parser & Handler",
        evidence=(
            "Simulasi fuzzing assessment:\n"
            "1. Tidak ada evidence bahwa FIX parser telah di-fuzz-test "
            "   menggunakan automated fuzzer (AFL, LibFuzzer, dll.)\n"
            "2. Error handling pada malformed message menggunakan generic "
            "   exception catch - tidak ada specific handling per error type\n"
            "3. Stack trace terekspos pada log saat parsing error - "
            "   information disclosure\n"
            "4. Oversized tag value tidak di-truncate - potensial "
            "   heap buffer overflow pada legacy C components\n"
            "5. Nested repeating group dengan depth >10 menyebabkan "
            "   stack overflow pada parser rekursif"
        ),
        recommendation=(
            "1. Lakukan automated fuzz testing pada FIX parser:\n"
            "   - Gunakan AFL++ atau LibFuzzer\n"
            "   - Buat corpus dari real FIX messages\n"
            "   - Jalankan fuzzing minimum 72 jam\n"
            "2. Implementasi safe error handling - jangan expose stack trace\n"
            "3. Terapkan hard limit pada tag value length\n"
            "4. Batasi repeating group nesting depth (max 5)\n"
            "5. Gunakan memory-safe parsing (bounds checking, ASAN)\n"
            "6. Jadwalkan protocol fuzz testing setiap rilis baru"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.14.2.8", "NIST SP 800-53 SI-10",
                         "OWASP Testing Guide v4"],
        risk_score=7.2
    ))

    # =========================================================================
    # PROTO-08: Message Replay Attack Protection
    # =========================================================================
    findings.append(Finding(
        check_id="PROTO-08",
        title="Proteksi Terhadap Message Replay Attack",
        description=(
            "Pemeriksaan mekanisme perlindungan terhadap serangan replay "
            "pada pesan FIX Protocol, dimana attacker menangkap dan "
            "mengirimkan ulang pesan order yang valid."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.VULNERABLE,
        affected_component="FIX Protocol Replay Protection",
        evidence=(
            "Simulasi replay attack assessment:\n"
            "1. Replay protection hanya mengandalkan sequence number - "
            "   jika sequence direset, pesan lama bisa di-replay\n"
            "2. Tidak ada timestamp-based replay window - pesan dari "
            "   hari sebelumnya dengan sequence number valid bisa diproses\n"
            "3. SendingTime (tag 52) tidak diverifikasi terhadap server "
            "   time - bisa dimanipulasi\n"
            "4. FIX session key (untuk encryption) tidak di-rotate "
            "   per-session\n"
            "5. Resend mechanism (MsgType=2) tidak memiliki replay "
            "   detection - legitimate resend vs. malicious replay "
            "   tidak dibedakan"
        ),
        recommendation=(
            "1. Implementasi dual protection: sequence number + timestamp "
            "   verification (max clock skew 5 detik)\n"
            "2. Tambahkan per-message nonce/unique identifier yang "
            "   tidak bisa diprediksi\n"
            "3. Implementasi sliding window replay detection\n"
            "4. Rotate session encryption key setiap 1 jam\n"
            "5. Tambahkan replay detection logic pada resend mechanism - "
            "   verifikasi bahwa resend request valid berdasarkan gap\n"
            "6. Log semua message rejection dengan alasan untuk forensic"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.14.1.3", "NIST SP 800-53 SC-8",
                         "FIX Protocol Security Recommendations"],
        risk_score=8.1
    ))

    return findings
