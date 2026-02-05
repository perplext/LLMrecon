"""
Application Security Checks for JATS-NextG
===========================================

Pemeriksaan keamanan aplikasi perdagangan JATS-NextG termasuk matching
engine, order management system, risk management, dan clearing system.

Domain Audit:
- APP-01: Matching engine integrity & race condition
- APP-02: Input validation & business logic
- APP-03: Error handling & information disclosure
- APP-04: Trading limit enforcement
- APP-05: Risk management system bypass
- APP-06: Clearing & settlement system security
- APP-07: Market surveillance system integrity
- APP-08: Application patch management
"""

from dataclasses import dataclass, field
from typing import List, Dict
from .network_security import Finding, Severity, CheckStatus


def run_application_security_checks(config: Dict) -> List[Finding]:
    """
    Menjalankan pemeriksaan keamanan aplikasi JATS-NextG.
    """
    findings = []

    # =========================================================================
    # APP-01: Matching Engine Integrity & Race Condition
    # =========================================================================
    findings.append(Finding(
        check_id="APP-01",
        title="Integritas Matching Engine & Race Condition",
        description=(
            "Audit integritas order matching engine JATS-NextG yang "
            "mencocokkan pesanan beli dan jual secara otomatis. Verifikasi "
            "ketahanan terhadap race condition, order manipulation, dan "
            "timestamp manipulation yang dapat mempengaruhi fairness "
            "perdagangan."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="JATS-NextG Matching Engine",
        evidence=(
            "Simulasi audit matching engine:\n"
            "1. Time-Of-Check-to-Time-Of-Use (TOCTOU) vulnerability pada "
            "   validasi trading limit - limit dicek sebelum order masuk "
            "   queue namun bisa berubah sebelum execution\n"
            "2. Timestamp granularity hanya millisecond - pada volume "
            "   tinggi, ordering fairness tidak terjamin untuk order "
            "   yang masuk pada millisecond yang sama\n"
            "3. Order priority manipulation dimungkinkan melalui "
            "   cancel-replace (amend) yang mempertahankan time priority "
            "   pada price improvement kecil\n"
            "4. Matching engine recovery setelah crash tidak diverifikasi "
            "   konsistensi order book-nya secara otomatis\n"
            "5. Tidak ada chaos testing / fault injection yang teratur "
            "   pada matching engine\n"
            "6. Dead code path terdeteksi pada error handler - potential "
            "   undefined behavior pada edge case"
        ),
        recommendation=(
            "1. Implementasi atomic check-and-execute untuk trading limit "
            "   verification - gunakan pessimistic locking\n"
            "2. Upgrade timestamp ke microsecond atau nanosecond precision\n"
            "3. Review cancel-replace logic - pastikan amend yang "
            "   mengubah harga kehilangan time priority\n"
            "4. Implementasi automated consistency check pada matching "
            "   engine recovery (order book reconciliation)\n"
            "5. Jadwalkan chaos testing bulanan dengan game day exercises\n"
            "6. Lakukan code audit menyeluruh pada matching engine - "
            "   eliminasi dead code paths"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.14.2.5", "OJK POJK 38/2016 Pasal 38",
                         "IOSCO Principles for Financial Markets"],
        risk_score=9.6
    ))

    # =========================================================================
    # APP-02: Input Validation & Business Logic
    # =========================================================================
    findings.append(Finding(
        check_id="APP-02",
        title="Validasi Input & Keamanan Business Logic",
        description=(
            "Pemeriksaan validasi input pada seluruh entry point "
            "JATS-NextG dan evaluasi keamanan business logic termasuk "
            "order validation, price validation, dan quantity check."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Order Management System - Input Validation",
        evidence=(
            "Simulasi audit input validation:\n"
            "1. Integer overflow tidak di-handle pada quantity field - "
            "   order dengan quantity mendekati MAX_INT bisa melewati "
            "   certain checks\n"
            "2. Price tick validation hanya dilakukan pada server-side - "
            "   RT client bisa mengirim harga fraksi yang dibulatkan "
            "   di server (rounding bias)\n"
            "3. Cross-market order validation (jika ada dual listing) "
            "   tidak atomic - timing attack dimungkinkan\n"
            "4. Special order types (pre-opening, closing) memiliki "
            "   validasi yang lebih longgar - bisa dieksploitasi\n"
            "5. Instrument status check (suspend, halt) menggunakan "
            "   cache yang diupdate asinkron - gap window ada"
        ),
        recommendation=(
            "1. Terapkan safe integer arithmetic - gunakan big integer "
            "   atau explicit overflow check pada seluruh numeric fields\n"
            "2. Implementasi validasi harga tick di kedua sisi (client & "
            "   server) dengan identical logic\n"
            "3. Terapkan atomic cross-validation untuk cross-market orders\n"
            "4. Review dan perketat validasi untuk special order types\n"
            "5. Gunakan synchronous instrument status check atau "
            "   sangat kurangi cache TTL (<1 detik)\n"
            "6. Implementasi comprehensive input fuzzing test suite"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.14.1.2", "OWASP ASVS 5.1",
                         "NIST SP 800-53 SI-10"],
        risk_score=7.8
    ))

    # =========================================================================
    # APP-03: Error Handling & Information Disclosure
    # =========================================================================
    findings.append(Finding(
        check_id="APP-03",
        title="Penanganan Error & Pencegahan Information Disclosure",
        description=(
            "Evaluasi error handling pada aplikasi JATS-NextG untuk "
            "memastikan tidak ada kebocoran informasi sensitif melalui "
            "error message, log, atau diagnostic output."
        ),
        severity=Severity.MEDIUM,
        status=CheckStatus.WARNING,
        affected_component="Application Error Handling",
        evidence=(
            "Simulasi audit error handling:\n"
            "1. FIX rejection message (MsgType=3) mengandung internal "
            "   error code yang mengungkap arsitektur internal\n"
            "2. Debug mode masih aktif pada staging environment yang "
            "   bisa diakses dari production network\n"
            "3. Stack trace terekspos pada REST API error response\n"
            "4. Log level DEBUG pada beberapa komponen menulis "
            "   credential dan token ke log file\n"
            "5. Error counter metric terekspos tanpa authentication "
            "   pada monitoring endpoint"
        ),
        recommendation=(
            "1. Gunakan generic error code pada FIX rejection - map "
            "   internal code ke user-facing code\n"
            "2. Isolasi staging dari production network sepenuhnya\n"
            "3. Sanitize seluruh API error response - hapus stack trace\n"
            "4. Audit log level dan pastikan credentials di-mask pada "
            "   semua log level\n"
            "5. Tambahkan authentication pada monitoring endpoints\n"
            "6. Implementasi centralized error handling framework"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.14.1.2", "OWASP ASVS 7.1",
                         "CWE-209"],
        risk_score=6.0
    ))

    # =========================================================================
    # APP-04: Trading Limit Enforcement
    # =========================================================================
    findings.append(Finding(
        check_id="APP-04",
        title="Penegakan Trading Limit",
        description=(
            "Verifikasi penegakan trading limit pada seluruh level: "
            "per-broker, per-investor, per-instrument, dan market-wide "
            "circuit breaker."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="Risk Management - Trading Limits",
        evidence=(
            "Simulasi audit trading limits:\n"
            "1. Trading limit check menggunakan eventual consistency - "
            "   limit bisa terexceed pada burst order karena async update\n"
            "2. Aggregate exposure calculation tidak real-time - "
            "   disinkronkan setiap 5 detik, cukup untuk gaming\n"
            "3. Auto-rejection price band check menggunakan previous "
            "   closing price yang di-cache - stale data pada split/CA\n"
            "4. Circuit breaker threshold configuration bisa diubah "
            "   oleh operator tanpa dual-authorization\n"
            "5. Short selling limit enforcement bergantung pada data "
            "   dari lembaga settlement yang tidak real-time"
        ),
        recommendation=(
            "1. Implementasi synchronous (blocking) limit check dengan "
            "   atomic counter per-broker\n"
            "2. Terapkan real-time aggregate exposure calculation - "
            "   sub-second update\n"
            "3. Validasi price band terhadap multiple data source "
            "   (real-time reference price, not just cache)\n"
            "4. Wajibkan dual-authorization untuk circuit breaker "
            "   configuration change\n"
            "5. Implementasi real-time interface dengan lembaga "
            "   settlement untuk short selling data\n"
            "6. Tambahkan pre-trade risk check layer sebelum order "
            "   masuk ke matching engine"
        ),
        remediation_effort="High",
        compliance_refs=["OJK POJK 38/2016 Pasal 39",
                         "IOSCO Market Integrity Principles",
                         "BEI Peraturan Perdagangan"],
        risk_score=9.0
    ))

    # =========================================================================
    # APP-05: Risk Management System Bypass
    # =========================================================================
    findings.append(Finding(
        check_id="APP-05",
        title="Potensi Bypass Risk Management System",
        description=(
            "Identifikasi kemungkinan bypass terhadap risk management "
            "controls pada JATS-NextG yang dapat memungkinkan trading "
            "yang melanggar aturan pasar."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="Risk Management System",
        evidence=(
            "Simulasi risk management bypass assessment:\n"
            "1. Risk management check bisa di-bypass melalui 'force "
            "   execute' flag yang seharusnya hanya untuk testing - "
            "   flag masih aktif di production\n"
            "2. Batch order submission bisa mengexceed individual "
            "   order limit jika disubmit dalam satu batch API call\n"
            "3. Cross-account trading limit tidak terkonsolidasi - "
            "   satu beneficial owner dengan multiple accounts bisa "
            "   exceed limit\n"
            "4. Market making exemption flag tidak diaudit - broker "
            "   bisa memiliki exemption yang sudah expired\n"
            "5. Risk parameter update propagation delay 10-30 detik - "
            "   window untuk trading sebelum new limit berlaku"
        ),
        recommendation=(
            "1. Hapus 'force execute' flag dari production code - "
            "   gunakan feature flag yang controlled di config server\n"
            "2. Terapkan aggregate limit check pada batch submission\n"
            "3. Implementasi beneficial owner consolidation untuk "
            "   cross-account limit enforcement\n"
            "4. Automated audit market making exemptions - auto-expire\n"
            "5. Implementasi zero-downtime risk parameter update - "
            "   atomic switch tanpa propagation delay\n"
            "6. Tambahkan real-time risk dashboard dengan alert"
        ),
        remediation_effort="High",
        compliance_refs=["OJK POJK 38/2016 Pasal 40",
                         "ISO 27001:A.14.2.5",
                         "IOSCO Risk Management Principles"],
        risk_score=9.4
    ))

    # =========================================================================
    # APP-06: Clearing & Settlement System Security
    # =========================================================================
    findings.append(Finding(
        check_id="APP-06",
        title="Keamanan Sistem Kliring & Setelmen",
        description=(
            "Audit keamanan interface antara JATS-NextG dan sistem "
            "kliring/setelmen (KPEI/KSEI) termasuk data integrity, "
            "reconciliation, dan failover."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Clearing & Settlement Interface",
        evidence=(
            "Simulasi audit clearing interface:\n"
            "1. Reconciliation antara JATS trade register dan KPEI "
            "   dilakukan batch (T+0 EOD) - discrepancy baru terdeteksi "
            "   keesokan hari\n"
            "2. Settlement instruction file transfer menggunakan SFTP "
            "   tanpa file integrity verification (no GPG signature)\n"
            "3. Failover ke backup clearing path belum di-test dalam "
            "   12 bulan terakhir\n"
            "4. Timeout handling pada clearing interface tidak "
            "   terdefinisi dengan jelas - hanging transaction possible\n"
            "5. Clearing data retention tidak terenkripsi pada archive"
        ),
        recommendation=(
            "1. Implementasi near-real-time reconciliation (setiap 15 "
            "   menit) antara JATS dan clearing system\n"
            "2. Tambahkan GPG/PGP signature pada file transfer dan "
            "   verify sebelum processing\n"
            "3. Jadwalkan clearing failover test setiap kuartal\n"
            "4. Definisikan explicit timeout dan retry policy untuk "
            "   clearing interface\n"
            "5. Enkripsi clearing data archive menggunakan AES-256-GCM\n"
            "6. Implementasi idempotency pada clearing messages"
        ),
        remediation_effort="Medium",
        compliance_refs=["OJK POJK 38/2016 Pasal 41",
                         "ISO 27001:A.14.1.2",
                         "CPMI-IOSCO PFMI Principles"],
        risk_score=7.2
    ))

    # =========================================================================
    # APP-07: Market Surveillance System Integrity
    # =========================================================================
    findings.append(Finding(
        check_id="APP-07",
        title="Integritas Sistem Pengawasan Pasar",
        description=(
            "Evaluasi integritas sistem market surveillance yang "
            "digunakan untuk mendeteksi insider trading, market "
            "manipulation, dan abnormal trading activity."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Market Surveillance System",
        evidence=(
            "Simulasi audit market surveillance:\n"
            "1. Surveillance alert threshold statis - tidak "
            "   beradaptasi dengan market condition (volatile vs calm)\n"
            "2. Data feed ke surveillance system memiliki delay 1-3 "
            "   detik dari matching engine - manipulasi bisa terjadi "
            "   sebelum terdeteksi\n"
            "3. Layering/spoofing detection rule tidak mencakup "
            "   variasi modern (phased layering)\n"
            "4. Cross-market surveillance tidak terintegrasi "
            "   (saham, derivatif, obligasi)\n"
            "5. False positive rate 45% menyebabkan alert fatigue\n"
            "6. Surveillance data bisa diakses oleh market operations "
            "   - potensial information leak"
        ),
        recommendation=(
            "1. Implementasi adaptive threshold berbasis ML yang "
            "   menyesuaikan dengan volatility regime\n"
            "2. Kurangi surveillance data feed delay ke <100ms\n"
            "3. Update detection rules untuk mencakup:\n"
            "   - Phased layering/spoofing\n"
            "   - Cross-product manipulation\n"
            "   - Momentum ignition\n"
            "   - Quote stuffing\n"
            "4. Integrasikan cross-market surveillance\n"
            "5. Tuning rules untuk false positive rate <20%\n"
            "6. Restrict surveillance data access hanya untuk "
            "   surveillance team dengan need-to-know"
        ),
        remediation_effort="High",
        compliance_refs=["OJK POJK 38/2016 Pasal 42",
                         "IOSCO Market Surveillance Recommendations",
                         "MAR (Market Abuse Regulation) equivalent"],
        risk_score=7.5
    ))

    # =========================================================================
    # APP-08: Application Patch Management
    # =========================================================================
    findings.append(Finding(
        check_id="APP-08",
        title="Manajemen Patch Aplikasi",
        description=(
            "Evaluasi proses patch management untuk seluruh komponen "
            "aplikasi JATS-NextG termasuk matching engine, RT client, "
            "gateway, dan supporting systems."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Application Lifecycle Management",
        evidence=(
            "Simulasi audit patch management:\n"
            "1. Critical security patch untuk FIX engine library belum "
            "   diterapkan >60 hari sejak rilis\n"
            "2. RT client software update tidak mandatory - 20% broker "
            "   menggunakan versi lama dengan known vulnerability\n"
            "3. Patch testing hanya dilakukan pada functional level - "
            "   tidak ada security regression testing\n"
            "4. Rollback procedure tidak otomatis - memerlukan 2-4 jam\n"
            "5. Dependency management tidak terpusat - third-party "
            "   library vulnerability tidak ter-track"
        ),
        recommendation=(
            "1. Terapkan SLA patch: Critical <7 hari, High <30 hari, "
            "   Medium <90 hari\n"
            "2. Wajibkan minimum software version untuk RT client - "
            "   blokir koneksi dari versi yang vulnerable\n"
            "3. Tambahkan security regression testing pada patch process\n"
            "4. Implementasi automated rollback capability (<15 menit)\n"
            "5. Deploy Software Composition Analysis (SCA) untuk "
            "   track third-party dependency vulnerabilities\n"
            "6. Implementasi CI/CD pipeline dengan security scanning"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.12.6.1", "OJK POJK 38/2016 Pasal 43",
                         "NIST SP 800-40r4", "PCI DSS 6.2"],
        risk_score=7.0
    ))

    return findings
