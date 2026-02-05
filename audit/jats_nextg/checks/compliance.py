"""
Regulatory Compliance Checks for JATS-NextG
=============================================

Pemeriksaan kepatuhan regulasi untuk JATS-NextG terhadap standar dan
regulasi yang berlaku di Indonesia dan internasional.

Domain Audit:
- COMP-01: OJK POJK 38/2016 (Penerapan Manajemen Risiko TI)
- COMP-02: ISO 27001:2022 compliance
- COMP-03: PCI DSS (untuk data pembayaran)
- COMP-04: Business continuity regulatory requirements
- COMP-05: Data privacy (UU PDP / Personal Data Protection)
- COMP-06: Third-party / vendor risk management
"""

from dataclasses import dataclass, field
from typing import List, Dict
from .network_security import Finding, Severity, CheckStatus


def run_compliance_checks(config: Dict) -> List[Finding]:
    """
    Menjalankan pemeriksaan kepatuhan regulasi JATS-NextG.
    """
    findings = []

    # =========================================================================
    # COMP-01: OJK POJK 38/2016 Compliance
    # =========================================================================
    findings.append(Finding(
        check_id="COMP-01",
        title="Kepatuhan OJK POJK 38/POJK.03/2016",
        description=(
            "Evaluasi kepatuhan terhadap Peraturan OJK Nomor 38 Tahun 2016 "
            "tentang Penerapan Manajemen Risiko dalam Penggunaan Teknologi "
            "Informasi oleh Lembaga Jasa Keuangan, yang berlaku bagi BEI "
            "sebagai Self-Regulatory Organization (SRO)."
        ),
        severity=Severity.CRITICAL,
        status=CheckStatus.PARTIAL,
        affected_component="Regulatory Compliance - OJK",
        evidence=(
            "Simulasi gap analysis POJK 38/2016:\n"
            "1. Pasal 7 - IT Strategic Plan: Sudah ada namun belum "
            "   di-update untuk mencerminkan JATS-NextG upgrade plan\n"
            "2. Pasal 12 - Security Testing: Penetration testing "
            "   dilakukan tahunan - regulasi mensyaratkan 'secara "
            "   berkala' yang lazim diinterpretasikan semi-annual\n"
            "3. Pasal 15 - IT Audit: Internal IT audit terakhir "
            "   8 bulan lalu, namun beberapa temuan belum ditindaklanjuti\n"
            "4. Pasal 18 - Penanganan Insiden: Prosedur pelaporan "
            "   insiden ke OJK belum terintegrasi dengan IR process\n"
            "5. Pasal 20 - BCP/DRP: DR test terakhir 9 bulan lalu "
            "   (lihat INFRA-05) - melebihi batas 6 bulan\n"
            "6. Pasal 24 - Vendor Management: Risk assessment vendor "
            "   IT belum dilakukan secara menyeluruh\n"
            "7. Pasal 26 - Pelaporan: Laporan insiden TI belum "
            "   menggunakan format yang disyaratkan OJK"
        ),
        recommendation=(
            "1. Update IT Strategic Plan untuk mencakup JATS-NextG "
            "   modernization roadmap\n"
            "2. Tingkatkan penetration testing ke semi-annual (2x/tahun)\n"
            "3. Tindaklanjuti seluruh temuan IT audit dalam 90 hari\n"
            "4. Integrasikan prosedur pelaporan OJK ke incident response "
            "   workflow - automated notification template\n"
            "5. Jadwalkan DR test setiap 6 bulan (lihat INFRA-05)\n"
            "6. Lakukan comprehensive vendor risk assessment\n"
            "7. Implementasi format pelaporan insiden sesuai SEOJK\n"
            "8. Assign dedicated compliance officer untuk monitoring "
            "   OJK regulatory changes"
        ),
        remediation_effort="High",
        compliance_refs=["OJK POJK 38/POJK.03/2016",
                         "SEOJK 21/SEOJK.03/2017"],
        risk_score=8.5
    ))

    # =========================================================================
    # COMP-02: ISO 27001:2022 Compliance
    # =========================================================================
    findings.append(Finding(
        check_id="COMP-02",
        title="Kepatuhan ISO 27001:2022",
        description=(
            "Gap analysis kepatuhan terhadap standar ISO 27001:2022 "
            "Information Security Management System (ISMS) untuk "
            "infrastruktur JATS-NextG."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.PARTIAL,
        affected_component="ISMS - ISO 27001:2022",
        evidence=(
            "Simulasi gap analysis ISO 27001:2022:\n"
            "1. Annex A.5 (Organizational Controls): Information Security "
            "   Policy di-review terakhir 14 bulan lalu (requirement: "
            "   planned intervals, best practice: annual)\n"
            "2. Annex A.6 (People Controls): Security awareness training "
            "   completion rate 75% - target 100%\n"
            "3. Annex A.7 (Physical Controls): Gap pada physical security "
            "   monitoring (lihat NET-06)\n"
            "4. Annex A.8 (Technological Controls):\n"
            "   - 8.7 (Malware Protection): Endpoint protection coverage 90%\n"
            "   - 8.8 (Vulnerability Management): Gap pada patch SLA\n"
            "   - 8.12 (Data Leakage Prevention): DLP belum diimplementasi\n"
            "   - 8.16 (Monitoring): Gap pada SIEM coverage (lihat MON-01)\n"
            "   - 8.23 (Web Filtering): Tidak relevan (closed network) namun "
            "     management network perlu web filtering\n"
            "   - 8.25 (Secure Development): SDLC security belum matang\n"
            "5. New controls in 2022 edition (A.5.7 Threat Intelligence, "
            "   A.8.11 Data Masking, A.8.28 Secure Coding) belum "
            "   sepenuhnya diimplementasi"
        ),
        recommendation=(
            "1. Review dan update Information Security Policy (annual)\n"
            "2. Targetkan 100% security awareness training completion\n"
            "3. Remediasi physical security gaps\n"
            "4. Implementasi missing technological controls:\n"
            "   - Complete endpoint protection coverage ke 100%\n"
            "   - Deploy DLP solution\n"
            "   - Extend SIEM coverage ke 100%\n"
            "5. Implementasi new 2022 controls:\n"
            "   - A.5.7: Integrate threat intelligence\n"
            "   - A.8.11: Implement data masking\n"
            "   - A.8.28: Formalize secure coding practices\n"
            "6. Jadwalkan internal ISMS audit setiap 6 bulan\n"
            "7. Target re-certification audit dalam 6 bulan"
        ),
        remediation_effort="High",
        compliance_refs=["ISO 27001:2022", "ISO 27002:2022"],
        risk_score=7.0
    ))

    # =========================================================================
    # COMP-03: PCI DSS Applicability
    # =========================================================================
    findings.append(Finding(
        check_id="COMP-03",
        title="Penerapan PCI DSS (jika applicable)",
        description=(
            "Evaluasi apakah komponen JATS-NextG yang menangani "
            "data pembayaran atau settlement perlu mematuhi PCI DSS "
            "dan sejauh mana kepatuhannya."
        ),
        severity=Severity.MEDIUM,
        status=CheckStatus.PARTIAL,
        affected_component="Payment Data Handling Components",
        evidence=(
            "Simulasi PCI DSS scoping assessment:\n"
            "1. JATS-NextG tidak langsung memproses data kartu namun "
            "   interface dengan settlement system yang menghandle "
            "   fund transfer - limited PCI scope\n"
            "2. Bank account information pada member database tidak "
            "   termasuk PCI scope namun perlu proteksi setara\n"
            "3. Settlement file transfer mengandung financial data "
            "   yang memerlukan encryption (lihat APP-06)\n"
            "4. Third-party payment processor compliance status "
            "   belum diverifikasi\n"
            "5. Segmentasi antara trading dan settlement component "
            "   belum divalidasi untuk PCI scoping"
        ),
        recommendation=(
            "1. Lakukan formal PCI DSS scoping assessment\n"
            "2. Terapkan PCI DSS controls pada in-scope components:\n"
            "   - Network segmentation (Req 1)\n"
            "   - Data encryption (Req 3, 4)\n"
            "   - Access control (Req 7, 8)\n"
            "   - Monitoring (Req 10)\n"
            "3. Verifikasi third-party payment processor PCI compliance "
            "   (AOC - Attestation of Compliance)\n"
            "4. Validasi network segmentation untuk PCI scope reduction\n"
            "5. Proteksi bank account data setara dengan PCI controls"
        ),
        remediation_effort="Medium",
        compliance_refs=["PCI DSS v4.0", "PA-DSS"],
        risk_score=5.5
    ))

    # =========================================================================
    # COMP-04: Business Continuity Regulatory Requirements
    # =========================================================================
    findings.append(Finding(
        check_id="COMP-04",
        title="Kepatuhan Regulasi Business Continuity",
        description=(
            "Evaluasi kepatuhan program business continuity management "
            "terhadap regulasi OJK dan standar internasional (ISO 22301)."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Business Continuity Management",
        evidence=(
            "Simulasi audit BCM compliance:\n"
            "1. Business Impact Analysis (BIA) terakhir di-update "
            "   18 bulan lalu - perlu annual review\n"
            "2. Maximum Tolerable Downtime (MTD) untuk JATS-NextG "
            "   belum didefinisikan secara formal oleh management\n"
            "3. BCP testing hanya mencakup IT recovery - belum "
            "   termasuk business process recovery\n"
            "4. Crisis communication plan belum mencakup semua "
            "   broker member (155+ broker)\n"
            "5. Peraturan BEI tentang trading halt procedures saat "
            "   sistem failure perlu di-update\n"
            "6. Reciprocal agreement dengan bursa lain belum ada"
        ),
        recommendation=(
            "1. Update BIA secara annual dan setelah perubahan signifikan\n"
            "2. Definisikan MTD formal: trading disruption max 2 jam\n"
            "3. Perluas BCP testing ke business process level\n"
            "4. Update crisis communication plan termasuk seluruh broker\n"
            "5. Review dan update trading halt SOP\n"
            "6. Jajaki reciprocal agreement dengan bursa partner "
            "   (SGX, SET, dst.) untuk mutual backup\n"
            "7. Jadwalkan comprehensive BCP exercise (dengan broker "
            "   participation) setiap tahun"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 22301:2019", "OJK POJK 38/2016 Pasal 20",
                         "IOSCO BCP Guidelines"],
        risk_score=7.0
    ))

    # =========================================================================
    # COMP-05: UU PDP (Personal Data Protection)
    # =========================================================================
    findings.append(Finding(
        check_id="COMP-05",
        title="Kepatuhan UU Perlindungan Data Pribadi (UU PDP)",
        description=(
            "Evaluasi kepatuhan terhadap UU No. 27 Tahun 2022 tentang "
            "Perlindungan Data Pribadi yang berlaku bagi BEI sebagai "
            "pengendali data pribadi broker, investor, dan karyawan."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Data Privacy Compliance",
        evidence=(
            "Simulasi gap analysis UU PDP:\n"
            "1. Data Protection Officer (DPO) belum ditunjuk secara "
            "   formal sesuai UU PDP Pasal 53\n"
            "2. Record of Processing Activities (ROPA) belum "
            "   terdokumentasi untuk JATS-NextG data flows\n"
            "3. Hak subjek data (akses, koreksi, hapus) belum "
            "   memiliki mekanisme teknis yang jelas\n"
            "4. Data Protection Impact Assessment (DPIA) belum "
            "   dilakukan untuk JATS-NextG\n"
            "5. Cross-border data transfer assessment belum dilakukan "
            "   (data ke provider sistem/vendor asing)\n"
            "6. Data breach notification procedure belum sesuai "
            "   requirement UU PDP (72 jam)\n"
            "7. Consent management untuk broker/investor data belum "
            "   diimplementasi secara teknis"
        ),
        recommendation=(
            "1. Tunjuk DPO dan register ke otoritas yang berwenang\n"
            "2. Dokumentasikan ROPA untuk seluruh data processing di "
            "   JATS-NextG\n"
            "3. Implementasi teknis untuk hak subjek data:\n"
            "   - Portal akses data pribadi\n"
            "   - Mekanisme koreksi dan penghapusan\n"
            "   - Data portability export\n"
            "4. Lakukan DPIA untuk JATS-NextG\n"
            "5. Evaluasi cross-border data transfer dan terapkan "
            "   safeguards yang sesuai\n"
            "6. Update breach notification procedure: <72 jam ke "
            "   otoritas, 'tanpa penundaan' ke subjek data\n"
            "7. Deploy consent management platform"
        ),
        remediation_effort="High",
        compliance_refs=["UU No. 27 Tahun 2022 (UU PDP)",
                         "PP tentang PDP (pending)",
                         "GDPR (reference/best practice)"],
        risk_score=7.5
    ))

    # =========================================================================
    # COMP-06: Third-Party / Vendor Risk Management
    # =========================================================================
    findings.append(Finding(
        check_id="COMP-06",
        title="Manajemen Risiko Vendor / Pihak Ketiga",
        description=(
            "Evaluasi program manajemen risiko vendor dan pihak ketiga "
            "yang menyediakan komponen, layanan, atau akses ke "
            "infrastruktur JATS-NextG."
        ),
        severity=Severity.HIGH,
        status=CheckStatus.WARNING,
        affected_component="Vendor Risk Management",
        evidence=(
            "Simulasi audit vendor risk management:\n"
            "1. Vendor risk assessment tidak dilakukan secara "
            "   periodik - hanya saat onboarding awal\n"
            "2. 3 dari 8 critical vendor tidak memiliki SLA yang "
            "   mencakup security requirements\n"
            "3. Right to audit clause hanya ada pada 50% kontrak vendor\n"
            "4. Vendor access ke JATS-NextG environment tidak melalui "
            "   PAM solution - unmonitored access\n"
            "5. Supply chain risk assessment belum mencakup "
            "   sub-contractors vendor\n"
            "6. Vendor incident notification requirement tidak "
            "   terstandarisasi\n"
            "7. Tidak ada vendor security scorecard atau rating system"
        ),
        recommendation=(
            "1. Implementasi annual vendor risk reassessment\n"
            "2. Update kontrak untuk mencakup security SLA pada "
            "   seluruh critical vendors\n"
            "3. Tambahkan right to audit clause pada semua kontrak\n"
            "4. Route semua vendor access melalui PAM solution dengan "
            "   session recording\n"
            "5. Extend risk assessment ke vendor sub-contractors\n"
            "6. Standarisasi vendor incident notification (<24 jam)\n"
            "7. Implementasi vendor security scorecard/rating system\n"
            "8. Maintain vendor risk register yang diupdate kuartalan"
        ),
        remediation_effort="Medium",
        compliance_refs=["ISO 27001:A.15.1", "OJK POJK 38/2016 Pasal 24",
                         "NIST SP 800-53 SA-9", "NIST CSF ID.SC"],
        risk_score=7.2
    ))

    return findings
