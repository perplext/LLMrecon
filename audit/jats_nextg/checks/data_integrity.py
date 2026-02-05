"""
Data Integrity & Encryption Checks for JATS-NextG
===================================================

Pemeriksaan keamanan integritas data, enkripsi, dan perlindungan informasi
pada seluruh komponen penyimpanan dan transmisi data JATS-NextG.

Domain Audit:
- DATA-01: Database encryption (at-rest & in-transit)
- DATA-02: Transaction log integrity & non-repudiation
- DATA-03: Backup security & recovery verification
- DATA-04: Data classification & handling
- DATA-05: Cryptographic implementation review
- DATA-06: Key management & rotation
"""

from dataclasses import dataclass, field
from typing import List, Dict
from .network_security import Finding, Severity, CheckStatus


def run_data_integrity_checks(config: Dict) -> List[Finding]:
    """
    Menjalankan pemeriksaan keamanan data dan enkripsi JATS-NextG.
    """
    findings = []

    # =========================================================================
    # DATA-01: Database Encryption
    # =========================================================================
    findings.append(Finding(
        check_id="DATA-01",
        title="Enkripsi Database (At-Rest & In-Transit)",
        description=(
            "Verifikasi bahwa seluruh database JATS-NextG terenkripsi "
            "baik saat disimpan (at-rest) maupun saat ditransmisikan "
            "(in-transit), termasuk database order book, trade register, "
            "member database, dan audit log database."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="Database Infrastructure",
        evidence=(
            "Simulasi audit enkripsi database:\n"
            "1. Database order book menggunakan Transparent Data "
            "   Encryption (TDE) namun tablespace temporary tidak "
            "   terenkripsi - sensitive data bisa leak via temp tables\n"
            "2. Database replication stream tidak terenkripsi - data "
            "   order dan trade terbaca pada network capture antara "
            "   primary dan standby\n"
            "3. Database backup disimpan tanpa encryption - NFS mount "
            "   ke backup storage accessible oleh storage admin\n"
            "4. Column-level encryption untuk PII (broker contact, NIK) "
            "   belum diterapkan - bergantung sepenuhnya pada TDE\n"
            "5. Encryption key untuk TDE disimpan di filesystem lokal - "
            "   bukan di key management system/HSM\n"
            "6. Database audit log tablespace tidak terenkripsi secara "
            "   terpisah dari data tablespace"
        ),
        recommendation=(
            "1. Aktifkan TDE pada SEMUA tablespace termasuk temporary "
            "   dan undo/rollback tablespace\n"
            "2. Aktifkan enkripsi pada database replication (SSL/TLS)\n"
            "3. Implementasi encrypted backup menggunakan AES-256-GCM\n"
            "4. Terapkan column-level encryption pada PII fields\n"
            "5. Migrasi TDE key ke HSM (FIPS 140-2 Level 3)\n"
            "6. Segregasi enkripsi key antara data dan audit log\n"
            "7. Implementasi Database Activity Monitoring (DAM)"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.10.1.1", "OJK POJK 38/2016 Pasal 44",
                         "PCI DSS 3.4", "NIST SP 800-111"],
        risk_score=9.0
    ))

    # =========================================================================
    # DATA-02: Transaction Log Integrity & Non-Repudiation
    # =========================================================================
    findings.append(Finding(
        check_id="DATA-02",
        title="Integritas Transaction Log & Non-Repudiation",
        description=(
            "Verifikasi integritas dan immutability transaction log "
            "JATS-NextG yang mencatat seluruh aktivitas perdagangan. "
            "Log ini krusial untuk audit, dispute resolution, dan "
            "regulatory compliance."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.VULNERABLE,
        affected_component="Transaction Logging System",
        evidence=(
            "Simulasi audit transaction log:\n"
            "1. Transaction log disimpan dalam database yang bisa di-"
            "   UPDATE dan DELETE oleh DBA - bukan append-only/immutable\n"
            "2. Tidak ada cryptographic hash chain pada transaction log - "
            "   modifikasi tidak terdeteksi\n"
            "3. Transaction log pada primary dan DR site tidak memiliki "
            "   independent hash verification - manipulasi bisa "
            "   terreplikasi\n"
            "4. Timestamp pada transaction log berasal dari application "
            "   server - bisa dimanipulasi jika server compromised\n"
            "5. Non-repudiation hanya mengandalkan FIX session "
            "   authentication - tidak ada digital signature per-order\n"
            "6. Log retention policy 5 tahun namun tidak ada integrity "
            "   verification pada archived logs"
        ),
        recommendation=(
            "1. Implementasi append-only transaction log dengan write-once "
            "   storage (WORM) atau blockchain-based logging\n"
            "2. Terapkan cryptographic hash chain (Merkle tree) pada "
            "   transaction log - setiap entry di-hash dan linked\n"
            "3. Verifikasi hash chain secara independen pada DR site\n"
            "4. Gunakan trusted timestamp service (RFC 3161) yang "
            "   independen dari application server\n"
            "5. Implementasi per-order digital signature untuk "
            "   non-repudiation yang kuat\n"
            "6. Automated integrity verification pada archived logs "
            "   secara periodik (bulanan)"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.12.4.1", "OJK POJK 38/2016 Pasal 45",
                         "NIST SP 800-92", "IOSCO Audit Trail Principles"],
        risk_score=9.3
    ))

    # =========================================================================
    # DATA-03: Backup Security & Recovery Verification
    # =========================================================================
    findings.append(Finding(
        check_id="DATA-03",
        title="Keamanan Backup & Verifikasi Recovery",
        description=(
            "Audit keamanan proses backup data dan verifikasi kemampuan "
            "recovery untuk memastikan business continuity JATS-NextG."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Backup & Recovery Infrastructure",
        evidence=(
            "Simulasi audit backup:\n"
            "1. Backup encryption menggunakan AES-128 - di bawah "
            "   rekomendasi minimum AES-256 untuk data keuangan\n"
            "2. Backup verification (restore test) terakhir dilakukan "
            "   6 bulan lalu - seharusnya minimal kuartalan\n"
            "3. Offsite backup transfer menggunakan jalur yang tidak "
            "   terpisah dari production traffic\n"
            "4. Recovery Time Objective (RTO) belum di-validasi "
            "   dengan actual restore test dalam 1 tahun terakhir\n"
            "5. Backup retention key management tidak terintegrasi "
            "   dengan HSM - key bisa hilang sebelum data expired\n"
            "6. Tidak ada air-gapped backup untuk ransomware protection"
        ),
        recommendation=(
            "1. Upgrade backup encryption ke AES-256-GCM\n"
            "2. Jadwalkan restore test kuartalan dengan documented results\n"
            "3. Gunakan dedicated backup network yang terpisah\n"
            "4. Validasi RTO dan RPO dengan actual test setiap 6 bulan\n"
            "5. Integrasikan backup key management dengan HSM\n"
            "6. Implementasi air-gapped / immutable backup (3-2-1 rule "
            "   dengan 1 copy offline)\n"
            "7. Otomasi backup integrity verification setiap backup cycle"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.12.3.1", "OJK POJK 38/2016 Pasal 46",
                         "NIST SP 800-34r1", "PCI DSS 9.5"],
        risk_score=7.5
    ))

    # =========================================================================
    # DATA-04: Data Classification & Handling
    # =========================================================================
    findings.append(Finding(
        check_id="DATA-04",
        title="Klasifikasi Data & Penanganan Informasi",
        description=(
            "Evaluasi implementasi data classification scheme dan "
            "prosedur penanganan data pada JATS-NextG sesuai dengan "
            "tingkat sensitivitas data."
        ),
        severity=Severity.MEDIUM,
        status=CheckStatus.PARTIAL,
        affected_component="Data Governance",
        evidence=(
            "Simulasi audit data classification:\n"
            "1. Data classification policy sudah ada namun belum "
            "   diterapkan secara teknis (label/tag pada data)\n"
            "2. Non-public market information (NPMI) tidak dilabeli - "
            "   sulit menerapkan DLP (Data Loss Prevention)\n"
            "3. Penanganan data debug/diagnostic mengandung PII dan "
            "   data transaksi - tidak di-anonymize\n"
            "4. Data sharing dengan regulator (OJK) tidak memiliki "
            "   data minimization review\n"
            "5. Screen capture dan print tidak dibatasi pada workstation "
            "   yang menangani data market sensitive"
        ),
        recommendation=(
            "1. Implementasi teknis data classification (labeling/tagging)\n"
            "   Level: Top Secret, Confidential, Internal, Public\n"
            "2. Deploy DLP solution untuk NPMI protection\n"
            "3. Anonymize/pseudonymize data pada debug dan diagnostic\n"
            "4. Terapkan data minimization review untuk regulatory sharing\n"
            "5. Batasi screen capture dan print pada sensitive workstation\n"
            "6. Implementasi automated data discovery dan classification"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.8.2.1", "OJK POJK 38/2016 Pasal 47",
                         "NIST SP 800-60"],
        risk_score=6.0
    ))

    # =========================================================================
    # DATA-05: Cryptographic Implementation Review
    # =========================================================================
    findings.append(Finding(
        check_id="DATA-05",
        title="Review Implementasi Kriptografi",
        description=(
            "Audit implementasi kriptografi pada seluruh komponen "
            "JATS-NextG untuk memastikan penggunaan algoritma yang "
            "kuat dan implementasi yang benar."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.VULNERABLE,
        affected_component="Cryptographic Infrastructure",
        evidence=(
            "Simulasi crypto review:\n"
            "1. Beberapa komponen legacy menggunakan SHA-1 untuk "
            "   message digest - collision attack telah terbukti\n"
            "2. TLS cipher suite termasuk CBC mode ciphers yang rentan "
            "   terhadap BEAST dan Lucky13 attacks\n"
            "3. Random number generation pada session token menggunakan "
            "   PRNG yang di-seed dengan timestamp - predictable\n"
            "4. RSA key size 1024-bit terdeteksi pada internal service - "
            "   di bawah minimum 2048-bit\n"
            "5. Hardcoded initialization vector (IV) pada AES-CBC "
            "   encryption di beberapa modul\n"
            "6. Tidak ada crypto agility plan untuk post-quantum "
            "   cryptography migration"
        ),
        recommendation=(
            "1. Migrasi seluruh SHA-1 ke SHA-256 atau SHA-3\n"
            "2. Konfigurasi TLS cipher suite hanya GCM mode:\n"
            "   TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305\n"
            "3. Gunakan CSPRNG (Cryptographically Secure PRNG) untuk "
            "   semua security-sensitive random generation\n"
            "4. Upgrade semua RSA key ke minimum 3072-bit\n"
            "5. Gunakan random IV untuk setiap encryption operation\n"
            "6. Buat crypto agility plan dan mulai evaluasi "
            "   post-quantum algorithms (CRYSTALS-Kyber, CRYSTALS-Dilithium)\n"
            "7. Lakukan crypto inventory dan maintain crypto bill of materials"
        ),
        remediation_effort="High",
        cve_references=["CVE-2017-15361", "CVE-2011-3389", "CVE-2013-0169"],
        compliance_refs=["ISO 27001:A.10.1.1", "NIST SP 800-131A",
                         "PCI DSS 4.1", "NIST Post-Quantum Cryptography"],
        risk_score=8.2
    ))

    # =========================================================================
    # DATA-06: Key Management & Rotation
    # =========================================================================
    findings.append(Finding(
        check_id="DATA-06",
        title="Manajemen & Rotasi Kunci Kriptografi",
        description=(
            "Audit proses manajemen lifecycle kunci kriptografi termasuk "
            "generation, distribution, storage, rotation, dan destruction."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Key Management System",
        evidence=(
            "Simulasi audit key management:\n"
            "1. Key rotation jadwal tidak konsisten - TLS cert dirotasi "
            "   tahunan namun symmetric key (AES) tidak pernah dirotasi\n"
            "2. Key generation tidak selalu menggunakan HSM - beberapa "
            "   key di-generate menggunakan software di server\n"
            "3. Key escrow/backup procedure tidak terdokumentasi\n"
            "4. Destroyed key material tidak di-overwrite secara "
            "   kriptografis (crypto erase)\n"
            "5. Key access audit log tidak diintegrasi dengan SIEM\n"
            "6. Tidak ada key ceremony procedure yang terdokumentasi"
        ),
        recommendation=(
            "1. Terapkan key rotation policy:\n"
            "   - Symmetric keys: setiap 90 hari\n"
            "   - Asymmetric keys: setiap 1-2 tahun\n"
            "   - TLS certificates: setiap 1 tahun (max)\n"
            "2. Wajibkan HSM untuk semua key generation\n"
            "3. Dokumentasikan key escrow procedure dengan split custody\n"
            "4. Implementasi crypto erase untuk key destruction\n"
            "5. Integrasikan key access log ke SIEM\n"
            "6. Dokumentasikan dan laksanakan key ceremony procedure\n"
            "7. Implementasi automated key lifecycle management"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.10.1.2", "NIST SP 800-57",
                         "PCI DSS 3.5", "FIPS 140-2"],
        risk_score=7.0
    ))

    return findings
