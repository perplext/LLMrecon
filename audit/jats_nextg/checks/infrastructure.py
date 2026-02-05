"""
Infrastructure Security Checks for JATS-NextG
===============================================

Pemeriksaan keamanan infrastruktur server, operating system, dan
supporting systems untuk JATS-NextG.

Domain Audit:
- INFRA-01: OS hardening & CIS benchmark compliance
- INFRA-02: Patch management & vulnerability scanning
- INFRA-03: Service inventory & port security
- INFRA-04: Virtualization / container security
- INFRA-05: Disaster recovery & business continuity
- INFRA-06: Time synchronization (NTP) security
"""

from dataclasses import dataclass, field
from typing import List, Dict
from .network_security import Finding, Severity, CheckStatus


def run_infrastructure_checks(config: Dict) -> List[Finding]:
    """
    Menjalankan pemeriksaan keamanan infrastruktur JATS-NextG.
    """
    findings = []

    # =========================================================================
    # INFRA-01: OS Hardening & CIS Benchmark Compliance
    # =========================================================================
    findings.append(Finding(
        check_id="INFRA-01",
        title="OS Hardening & Kepatuhan CIS Benchmark",
        description=(
            "Evaluasi konfigurasi hardening OS pada seluruh server "
            "JATS-NextG terhadap CIS (Center for Internet Security) "
            "Benchmark yang berlaku."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.VULNERABLE,
        affected_component="Server OS - All JATS-NextG Servers",
        evidence=(
            "Simulasi CIS Benchmark audit:\n"
            "1. CIS Benchmark compliance rate rata-rata 65% - target "
            "   minimum 90% untuk critical infrastructure\n"
            "2. SELinux/AppArmor dalam mode permissive pada 3 server - "
            "   seharusnya enforcing\n"
            "3. Unnecessary services aktif: CUPS, Avahi, Bluetooth "
            "   subsystem pada server production\n"
            "4. File permission terlalu permisif pada /etc/shadow - "
            "   world-readable pada 1 server\n"
            "5. Core dump enabled untuk semua users - bisa leak memory "
            "   content termasuk keys\n"
            "6. ASLR (Address Space Layout Randomization) disabled pada "
            "   matching engine server untuk performance - mengurangi "
            "   exploit mitigation\n"
            "7. USB storage device tidak di-block pada server"
        ),
        recommendation=(
            "1. Remediasi CIS Benchmark gaps hingga compliance >95%\n"
            "2. Set SELinux/AppArmor ke enforcing mode pada semua server\n"
            "3. Nonaktifkan semua unnecessary services\n"
            "4. Fix file permission sesuai CIS Benchmark\n"
            "5. Disable core dump untuk non-debug environments\n"
            "6. Evaluasi re-enabling ASLR pada matching engine - modern "
            "   implementations memiliki overhead minimal\n"
            "7. Implementasi USB device blocking policy via udev rules\n"
            "8. Deploy automated CIS compliance scanning (weekly)"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.12.6.1", "CIS Benchmarks",
                         "NIST SP 800-123", "PCI DSS 2.2"],
        risk_score=7.8
    ))

    # =========================================================================
    # INFRA-02: Patch Management & Vulnerability Scanning
    # =========================================================================
    findings.append(Finding(
        check_id="INFRA-02",
        title="Manajemen Patch & Vulnerability Scanning",
        description=(
            "Evaluasi proses patch management dan vulnerability scanning "
            "untuk seluruh infrastruktur JATS-NextG."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.VULNERABLE,
        affected_component="Patch Management Infrastructure",
        evidence=(
            "Simulasi vulnerability scan:\n"
            "1. 47 vulnerabilities teridentifikasi:\n"
            "   - 3 Critical (CVSS >9.0)\n"
            "   - 12 High (CVSS 7.0-8.9)\n"
            "   - 18 Medium (CVSS 4.0-6.9)\n"
            "   - 14 Low (CVSS <4.0)\n"
            "2. Critical vulnerability: OpenSSL heap buffer overflow "
            "   (CVE-equivalent) pada 2 server - patch tersedia 45 hari\n"
            "3. Kernel vulnerability pada 5 server - local privilege "
            "   escalation\n"
            "4. Vulnerability scanning hanya dijalankan bulanan - "
            "   seharusnya mingguan untuk critical infrastructure\n"
            "5. Patch testing environment tidak mirror production - "
            "   configuration drift terdeteksi\n"
            "6. Emergency patch procedure rata-rata memerlukan 72 jam - "
            "   terlalu lama untuk critical vulnerability"
        ),
        recommendation=(
            "1. Remediasi 3 Critical vulnerabilities dalam 7 hari\n"
            "2. Remediasi 12 High vulnerabilities dalam 30 hari\n"
            "3. Tingkatkan vulnerability scanning ke mingguan\n"
            "4. Sinkronkan patch testing environment dengan production\n"
            "5. Kurangi emergency patch SLA menjadi 24 jam\n"
            "6. Implementasi virtual patching (WAF/IPS) sebagai "
            "   interim mitigation sebelum actual patching\n"
            "7. Deploy continuous vulnerability assessment tool"
        ),
        remediation_effort="High",
        cve_references=["CVE-2024-0567 (representative)",
                         "CVE-2024-1086 (representative)"],
        compliance_refs=["ISO 27001:A.12.6.1", "OJK POJK 38/2016 Pasal 48",
                         "NIST SP 800-40r4", "PCI DSS 6.2"],
        risk_score=8.5
    ))

    # =========================================================================
    # INFRA-03: Service Inventory & Port Security
    # =========================================================================
    findings.append(Finding(
        check_id="INFRA-03",
        title="Inventaris Service & Keamanan Port",
        description=(
            "Pemeriksaan service yang berjalan dan port yang terbuka "
            "pada seluruh server JATS-NextG untuk memastikan hanya "
            "service yang diperlukan yang aktif."
        ),
        severity=Severity.MEDIUM,
        status=CheckStatus.WARNING,
        affected_component="Server Services & Ports",
        evidence=(
            "Simulasi port scan dan service audit:\n"
            "1. 12 port non-essential terbuka pada server production:\n"
            "   - Port 8080 (debug web server) pada matching engine\n"
            "   - Port 5432 (PostgreSQL) accessible dari management VLAN\n"
            "   - Port 6379 (Redis) tanpa authentication\n"
            "   - Port 9090 (Prometheus) tanpa TLS\n"
            "2. Service inventory tidak ter-maintain - dokumentasi "
            "   terakhir update 8 bulan lalu\n"
            "3. Cron job pada 3 server menjalankan script yang tidak "
            "   terdokumentasi\n"
            "4. Web server (Nginx) dengan default configuration aktif "
            "   pada management interface"
        ),
        recommendation=(
            "1. Tutup seluruh port non-essential:\n"
            "   - Hapus debug web server dari production\n"
            "   - Restrict database port ke application server only\n"
            "   - Enable Redis authentication (requirepass)\n"
            "   - Enable TLS pada Prometheus\n"
            "2. Update service inventory dan jadwalkan review kuartalan\n"
            "3. Audit dan dokumentasikan seluruh cron job\n"
            "4. Harden Nginx configuration - hapus default page\n"
            "5. Implementasi automated service/port discovery dan "
            "   baseline comparison"
        ),
        remediation_effort="Low",
        compliance_refs=["ISO 27001:A.13.1.1", "CIS Benchmarks",
                         "NIST SP 800-123", "PCI DSS 2.2.2"],
        risk_score=6.5
    ))

    # =========================================================================
    # INFRA-04: Virtualization / Container Security
    # =========================================================================
    findings.append(Finding(
        check_id="INFRA-04",
        title="Keamanan Virtualisasi & Container",
        description=(
            "Audit keamanan platform virtualisasi dan container yang "
            "digunakan untuk menjalankan komponen JATS-NextG non-core "
            "(monitoring, reporting, management tools)."
        ),
        severity=Severity.MEDIUM,
        status=CheckStatus.WARNING,
        affected_component="Virtualization & Container Platform",
        evidence=(
            "Simulasi audit virtualisasi:\n"
            "1. Hypervisor patch terakhir diterapkan 3 bulan lalu - "
            "   2 CVE critical telah dirilis sejak itu\n"
            "2. VM escape mitigation (SMEP, SMAP) belum diaktifkan "
            "   pada seluruh host\n"
            "3. Container image scanning tidak diimplementasi - "
            "   base image mengandung 23 known vulnerabilities\n"
            "4. Container running as root pada 60% workloads\n"
            "5. Container network policy terlalu permisif - "
            "   semua container bisa berkomunikasi satu sama lain\n"
            "6. Hypervisor management interface accessible dari "
            "   regular management VLAN"
        ),
        recommendation=(
            "1. Terapkan hypervisor patching sesuai SLA (critical <7 hari)\n"
            "2. Aktifkan SMEP dan SMAP pada seluruh hypervisor host\n"
            "3. Implementasi container image scanning di CI/CD pipeline\n"
            "4. Jalankan container sebagai non-root user\n"
            "5. Terapkan container network policy (zero-trust)\n"
            "6. Isolasi hypervisor management ke dedicated VLAN\n"
            "7. Implementasi container runtime security (Falco/Sysdig)"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.12.6.1", "CIS Docker Benchmark",
                         "NIST SP 800-190"],
        risk_score=6.8
    ))

    # =========================================================================
    # INFRA-05: Disaster Recovery & Business Continuity
    # =========================================================================
    findings.append(Finding(
        check_id="INFRA-05",
        title="Disaster Recovery & Business Continuity",
        description=(
            "Evaluasi kesiapan disaster recovery dan business continuity "
            "untuk memastikan JATS-NextG dapat beroperasi jika terjadi "
            "gangguan besar pada infrastruktur utama."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.WARNING,
        affected_component="DR/BCP Infrastructure",
        evidence=(
            "Simulasi audit DR/BCP:\n"
            "1. DR site memiliki kapasitas hanya 60% dari production - "
            "   tidak bisa handle full load saat failover\n"
            "2. DR failover test terakhir 9 bulan lalu - seharusnya "
            "   setiap 6 bulan per regulasi OJK\n"
            "3. RTO target 2 jam namun actual test terakhir memerlukan "
            "   4.5 jam - tidak memenuhi target\n"
            "4. RPO target 0 (zero data loss) menggunakan synchronous "
            "   replication namun belum divalidasi pada failure scenario\n"
            "5. DR runbook terakhir di-update 14 bulan lalu - tidak "
            "   mencerminkan perubahan infrastruktur terkini\n"
            "6. Communication plan tidak mencakup semua stakeholder "
            "   (regulator, broker, media)\n"
            "7. Tidak ada automated failover - seluruh proses manual"
        ),
        recommendation=(
            "1. Upgrade DR site ke 100% capacity parity\n"
            "2. Jadwalkan DR test setiap 6 bulan sesuai regulasi\n"
            "3. Optimasi failover process untuk memenuhi RTO 2 jam\n"
            "4. Validasi RPO 0 dengan actual failure simulation\n"
            "5. Update DR runbook setiap ada perubahan infrastruktur\n"
            "6. Lengkapi communication plan untuk semua stakeholder\n"
            "7. Evaluasi implementasi automated failover dengan "
            "   manual confirmation step\n"
            "8. Lakukan table-top exercise setiap kuartal"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:A.17.1", "OJK POJK 38/2016 Pasal 49",
                         "ISO 22301", "NIST SP 800-34r1"],
        risk_score=8.5
    ))

    # =========================================================================
    # INFRA-06: Time Synchronization (NTP) Security
    # =========================================================================
    findings.append(Finding(
        check_id="INFRA-06",
        title="Keamanan Sinkronisasi Waktu (NTP)",
        description=(
            "Audit keamanan time synchronization pada infrastruktur "
            "JATS-NextG. Akurasi waktu sangat kritis untuk fairness "
            "perdagangan, audit trail, dan regulatory compliance."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="NTP Infrastructure",
        evidence=(
            "Simulasi audit NTP:\n"
            "1. NTP authentication (symmetric key / Autokey) tidak "
            "   diaktifkan - rentan terhadap NTP spoofing\n"
            "2. NTP source hanya dari 1 stratum-1 server - single "
            "   point of failure\n"
            "3. Time accuracy requirement ±1ms untuk trading namun "
            "   monitoring hanya mengecek ±100ms\n"
            "4. NTP traffic tidak terpisah dari regular traffic - "
            "   bisa di-intercept dan di-modify\n"
            "5. Tidak ada PTP (Precision Time Protocol) IEEE 1588 "
            "   pada matching engine - untuk sub-microsecond accuracy\n"
            "6. Leap second handling belum di-test"
        ),
        recommendation=(
            "1. Aktifkan NTP authentication (NTS - Network Time Security)\n"
            "2. Gunakan minimum 4 NTP source dari 2 stratum-1 provider "
            "   berbeda\n"
            "3. Implementasi time accuracy monitoring dengan threshold "
            "   ±1ms dan automated alert\n"
            "4. Isolasi NTP traffic pada dedicated VLAN atau gunakan "
            "   out-of-band NTP\n"
            "5. Deploy PTP (IEEE 1588) pada matching engine cluster\n"
            "6. Test dan dokumentasikan leap second handling procedure\n"
            "7. Deploy GPS-disciplined NTP server sebagai primary source"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.12.4.4", "MiFID II RTS 25 (reference)",
                         "NIST SP 800-53 AU-8"],
        risk_score=7.2
    ))

    return findings
