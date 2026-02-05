"""
JATS-NextG VA Report Generator
===============================

Generator laporan audit Vulnerability Assessment untuk JATS-NextG dalam
format JSON, Markdown, dan executive summary.
"""

import json
import os
from datetime import datetime
from typing import List, Dict, Tuple
from collections import Counter
from .checks.network_security import Finding, Severity, CheckStatus


class JASTReportGenerator:
    """Generator laporan VA untuk JATS-NextG."""

    SEVERITY_ORDER = {
        Severity.CRITICAL: 0,
        Severity.HIGH: 1,
        Severity.MEDIUM: 2,
        Severity.LOW: 3,
        Severity.INFO: 4,
    }

    SEVERITY_COLORS = {
        Severity.CRITICAL: "🔴",
        Severity.HIGH: "🟠",
        Severity.MEDIUM: "🟡",
        Severity.LOW: "🔵",
        Severity.INFO: "⚪",
    }

    STATUS_LABELS = {
        CheckStatus.VULNERABLE: "RENTAN (Vulnerable)",
        CheckStatus.WARNING: "PERINGATAN (Warning)",
        CheckStatus.SECURE: "AMAN (Secure)",
        CheckStatus.NOT_TESTED: "BELUM DIUJI (Not Tested)",
        CheckStatus.PARTIAL: "SEBAGIAN (Partial)",
    }

    def __init__(self, findings: List[Finding], config: Dict):
        self.findings = sorted(
            findings,
            key=lambda f: (self.SEVERITY_ORDER.get(f.severity, 99), -f.risk_score)
        )
        self.config = config
        self.timestamp = datetime.now()

    def calculate_statistics(self) -> Dict:
        """Menghitung statistik keseluruhan temuan VA."""
        severity_counts = Counter(f.severity for f in self.findings)
        status_counts = Counter(f.status for f in self.findings)

        critical_findings = [f for f in self.findings if f.severity == Severity.CRITICAL]
        high_findings = [f for f in self.findings if f.severity == Severity.HIGH]

        avg_risk_score = (
            sum(f.risk_score for f in self.findings) / len(self.findings)
            if self.findings else 0.0
        )

        max_risk = max((f.risk_score for f in self.findings), default=0.0)

        # Calculate overall risk rating
        if severity_counts.get(Severity.CRITICAL, 0) >= 3:
            overall_rating = "SANGAT TINGGI (Very High)"
        elif severity_counts.get(Severity.CRITICAL, 0) >= 1:
            overall_rating = "TINGGI (High)"
        elif severity_counts.get(Severity.HIGH, 0) >= 5:
            overall_rating = "TINGGI (High)"
        elif severity_counts.get(Severity.HIGH, 0) >= 2:
            overall_rating = "MENENGAH-TINGGI (Medium-High)"
        else:
            overall_rating = "MENENGAH (Medium)"

        return {
            "total_findings": len(self.findings),
            "severity_distribution": {s.value: severity_counts.get(s, 0) for s in Severity},
            "status_distribution": {s.value: status_counts.get(s, 0) for s in CheckStatus},
            "average_risk_score": round(avg_risk_score, 1),
            "max_risk_score": round(max_risk, 1),
            "critical_count": severity_counts.get(Severity.CRITICAL, 0),
            "high_count": severity_counts.get(Severity.HIGH, 0),
            "medium_count": severity_counts.get(Severity.MEDIUM, 0),
            "vulnerable_count": status_counts.get(CheckStatus.VULNERABLE, 0),
            "warning_count": status_counts.get(CheckStatus.WARNING, 0),
            "overall_risk_rating": overall_rating,
        }

    def generate_json_report(self) -> Dict:
        """Generate laporan dalam format JSON."""
        stats = self.calculate_statistics()

        report = {
            "report_metadata": {
                "title": "Vulnerability Assessment Report - JATS-NextG",
                "subtitle": "Sistem Perdagangan Otomatis Jakarta (JATS)-NextG",
                "organization": "Bursa Efek Indonesia (BEI)",
                "department": "Tim Audit IT & Cybersecurity",
                "assessment_type": "Vulnerability Assessment Simulation",
                "scope": "JATS-NextG Trading System including Remote Trading (RT) Gateway",
                "date": self.timestamp.strftime("%Y-%m-%d"),
                "timestamp": self.timestamp.isoformat(),
                "classification": "RAHASIA / CONFIDENTIAL",
                "version": "1.0",
                "framework_version": "JATS-VA-SIM v1.0.0",
            },
            "executive_summary": {
                "overall_risk_rating": stats["overall_risk_rating"],
                "total_findings": stats["total_findings"],
                "severity_distribution": stats["severity_distribution"],
                "status_distribution": stats["status_distribution"],
                "average_risk_score": stats["average_risk_score"],
                "max_risk_score": stats["max_risk_score"],
                "key_observations": self._generate_key_observations(stats),
                "top_5_risks": [
                    {
                        "check_id": f.check_id,
                        "title": f.title,
                        "risk_score": f.risk_score,
                        "severity": f.severity.value,
                    }
                    for f in self.findings[:5]
                ],
            },
            "findings": [
                {
                    "check_id": f.check_id,
                    "title": f.title,
                    "description": f.description,
                    "severity": f.severity.value,
                    "status": f.status.value,
                    "risk_score": f.risk_score,
                    "affected_component": f.affected_component,
                    "evidence": f.evidence,
                    "recommendation": f.recommendation,
                    "remediation_effort": f.remediation_effort,
                    "cve_references": f.cve_references,
                    "compliance_references": f.compliance_refs,
                }
                for f in self.findings
            ],
            "statistics": stats,
            "compliance_mapping": self._generate_compliance_mapping(),
            "remediation_roadmap": self._generate_remediation_roadmap(),
        }

        return report

    def generate_markdown_report(self) -> str:
        """Generate laporan dalam format Markdown."""
        stats = self.calculate_statistics()
        lines = []

        # Header
        lines.append("# LAPORAN VULNERABILITY ASSESSMENT")
        lines.append("# Sistem Perdagangan Otomatis Jakarta (JATS)-NextG")
        lines.append("")
        lines.append("---")
        lines.append("")
        lines.append("| Field | Detail |")
        lines.append("|-------|--------|")
        lines.append(f"| **Organisasi** | Bursa Efek Indonesia (BEI) |")
        lines.append(f"| **Departemen** | Tim Audit IT & Cybersecurity |")
        lines.append(f"| **Jenis Assessment** | Vulnerability Assessment Simulation |")
        lines.append(f"| **Ruang Lingkup** | JATS-NextG Trading System & Remote Trading (RT) Gateway |")
        lines.append(f"| **Tanggal** | {self.timestamp.strftime('%d %B %Y')} |")
        lines.append(f"| **Klasifikasi** | RAHASIA / CONFIDENTIAL |")
        lines.append(f"| **Versi Laporan** | 1.0 |")
        lines.append("")
        lines.append("---")
        lines.append("")

        # Executive Summary
        lines.append("## 1. RINGKASAN EKSEKUTIF")
        lines.append("")
        lines.append(f"### Rating Risiko Keseluruhan: **{stats['overall_risk_rating']}**")
        lines.append("")
        lines.append(f"Simulasi Vulnerability Assessment terhadap sistem JATS-NextG telah "
                     f"mengidentifikasi **{stats['total_findings']} temuan keamanan** yang "
                     f"memerlukan perhatian. Dari total temuan tersebut:")
        lines.append("")
        lines.append(f"| Severity | Jumlah |")
        lines.append(f"|----------|--------|")
        lines.append(f"| CRITICAL | **{stats['severity_distribution']['CRITICAL']}** |")
        lines.append(f"| HIGH     | **{stats['severity_distribution']['HIGH']}** |")
        lines.append(f"| MEDIUM   | **{stats['severity_distribution']['MEDIUM']}** |")
        lines.append(f"| LOW      | **{stats['severity_distribution'].get('LOW', 0)}** |")
        lines.append(f"| INFO     | **{stats['severity_distribution'].get('INFO', 0)}** |")
        lines.append("")
        lines.append(f"- **Skor Risiko Rata-rata**: {stats['average_risk_score']}/10.0")
        lines.append(f"- **Skor Risiko Tertinggi**: {stats['max_risk_score']}/10.0")
        lines.append(f"- **Temuan VULNERABLE**: {stats['vulnerable_count']}")
        lines.append(f"- **Temuan WARNING**: {stats['warning_count']}")
        lines.append("")

        # Key Observations
        lines.append("### Observasi Kunci:")
        lines.append("")
        for obs in self._generate_key_observations(stats):
            lines.append(f"- {obs}")
        lines.append("")

        # Top 5 Risks
        lines.append("### Top 5 Risiko Tertinggi:")
        lines.append("")
        lines.append("| No | ID | Temuan | Risk Score | Severity |")
        lines.append("|----|-----|--------|------------|----------|")
        for i, f in enumerate(self.findings[:5], 1):
            lines.append(f"| {i} | {f.check_id} | {f.title} | "
                        f"**{f.risk_score}** | {f.severity.value} |")
        lines.append("")
        lines.append("---")
        lines.append("")

        # Detailed Findings by Domain
        lines.append("## 2. TEMUAN DETAIL PER DOMAIN")
        lines.append("")

        domains = {
            "NET": ("Keamanan Jaringan (Network Security)", []),
            "PROTO": ("Keamanan Protokol Perdagangan (Protocol Security)", []),
            "AUTH": ("Otentikasi & Kontrol Akses (Authentication & Access Control)", []),
            "APP": ("Keamanan Aplikasi (Application Security)", []),
            "DATA": ("Integritas Data & Enkripsi (Data Integrity & Encryption)", []),
            "INFRA": ("Keamanan Infrastruktur (Infrastructure Security)", []),
            "MON": ("Pemantauan & Respons Insiden (Monitoring & Incident Response)", []),
            "COMP": ("Kepatuhan Regulasi (Regulatory Compliance)", []),
        }

        for f in self.findings:
            prefix = f.check_id.split("-")[0]
            if prefix in domains:
                domains[prefix][1].append(f)

        domain_num = 1
        for prefix, (domain_name, domain_findings) in domains.items():
            if not domain_findings:
                continue

            lines.append(f"### 2.{domain_num}. {domain_name}")
            lines.append("")

            for f in domain_findings:
                severity_icon = self.SEVERITY_COLORS.get(f.severity, "")
                status_label = self.STATUS_LABELS.get(f.status, f.status.value)

                lines.append(f"#### {severity_icon} {f.check_id}: {f.title}")
                lines.append("")
                lines.append(f"| Atribut | Nilai |")
                lines.append(f"|---------|-------|")
                lines.append(f"| **Severity** | {f.severity.value} |")
                lines.append(f"| **Status** | {status_label} |")
                lines.append(f"| **Risk Score** | {f.risk_score}/10.0 |")
                lines.append(f"| **Komponen Terdampak** | {f.affected_component} |")
                lines.append(f"| **Upaya Remediasi** | {f.remediation_effort} |")
                lines.append("")
                lines.append("**Deskripsi:**")
                lines.append(f"> {f.description}")
                lines.append("")
                lines.append("**Evidence (Bukti Simulasi):**")
                lines.append("```")
                lines.append(f"{f.evidence}")
                lines.append("```")
                lines.append("")
                lines.append("**Rekomendasi:**")
                lines.append("```")
                lines.append(f"{f.recommendation}")
                lines.append("```")
                lines.append("")

                if f.cve_references:
                    lines.append(f"**Referensi CVE:** {', '.join(f.cve_references)}")
                    lines.append("")

                if f.compliance_refs:
                    lines.append(f"**Referensi Compliance:** {', '.join(f.compliance_refs)}")
                    lines.append("")

                lines.append("---")
                lines.append("")

            domain_num += 1

        # Compliance Mapping
        lines.append("## 3. PEMETAAN KEPATUHAN (COMPLIANCE MAPPING)")
        lines.append("")
        mapping = self._generate_compliance_mapping()
        for framework, controls in mapping.items():
            lines.append(f"### {framework}")
            lines.append("")
            lines.append("| Control | Related Finding(s) |")
            lines.append("|---------|-------------------|")
            for control, finding_ids in controls.items():
                lines.append(f"| {control} | {', '.join(finding_ids)} |")
            lines.append("")

        # Remediation Roadmap
        lines.append("## 4. ROADMAP REMEDIASI")
        lines.append("")
        roadmap = self._generate_remediation_roadmap()

        for phase in roadmap:
            lines.append(f"### {phase['phase']}: {phase['title']}")
            lines.append(f"**Timeframe:** {phase['timeframe']}")
            lines.append("")
            lines.append("| Priority | ID | Temuan | Risk Score | Effort |")
            lines.append("|----------|-----|--------|------------|--------|")
            for item in phase["items"]:
                lines.append(f"| {item['priority']} | {item['check_id']} | "
                            f"{item['title']} | {item['risk_score']} | "
                            f"{item['effort']} |")
            lines.append("")

        # Risk Matrix
        lines.append("## 5. MATRIKS RISIKO")
        lines.append("")
        lines.append("```")
        lines.append("          IMPACT")
        lines.append("     Low    Med    High   Critical")
        lines.append("  +-------+------+-------+---------+")
        lines.append("H |       |      | NET-08| APP-01  |  L")
        lines.append("i |       |      | NET-06| APP-05  |  I")
        lines.append("g |       |      |       | PROTO-01|  K")
        lines.append("h |       |      |       | PROTO-06|  E")
        lines.append("  +-------+------+-------+---------+  L")
        lines.append("M |       |DATA04| NET-10| NET-04  |  I")
        lines.append("e |       |APP-03| MON-03| AUTH-01 |  H")
        lines.append("d |       |      | MON-05| AUTH-02 |  O")
        lines.append("  |       |      |       | AUTH-05 |  O")
        lines.append("  +-------+------+-------+---------+  D")
        lines.append("L |       |      |       |         |")
        lines.append("o |       |      |       |         |")
        lines.append("w |       |      |       |         |")
        lines.append("  +-------+------+-------+---------+")
        lines.append("```")
        lines.append("")

        # Conclusion
        lines.append("## 6. KESIMPULAN & LANGKAH SELANJUTNYA")
        lines.append("")
        lines.append(f"Simulasi Vulnerability Assessment pada JATS-NextG telah "
                     f"mengidentifikasi **{stats['severity_distribution']['CRITICAL']} "
                     f"kerentanan CRITICAL** dan **{stats['severity_distribution']['HIGH']} "
                     f"kerentanan HIGH** yang memerlukan penanganan segera.")
        lines.append("")
        lines.append("### Langkah Selanjutnya yang Direkomendasikan:")
        lines.append("")
        lines.append("1. **Immediate (0-30 hari):** Remediasi seluruh temuan CRITICAL")
        lines.append("2. **Short-term (30-90 hari):** Remediasi temuan HIGH dan mulai "
                     "implementasi quick wins")
        lines.append("3. **Medium-term (90-180 hari):** Implementasi improvement program "
                     "untuk MEDIUM findings")
        lines.append("4. **Long-term (180-365 hari):** Strategic improvements dan "
                     "compliance gap closure")
        lines.append("")
        lines.append("### Rekomendasi Investasi Prioritas:")
        lines.append("")
        lines.append("1. **PAM Solution** - Mengatasi AUTH-05 dan beberapa temuan terkait")
        lines.append("2. **SIEM Enhancement** - Mengatasi MON-01, MON-02, MON-06")
        lines.append("3. **PKI & HSM Upgrade** - Mengatasi AUTH-07, DATA-05, DATA-06")
        lines.append("4. **FIX Protocol Security Hardening** - Mengatasi PROTO-01 s/d PROTO-08")
        lines.append("5. **DR Site Capacity Upgrade** - Mengatasi INFRA-05, COMP-04")
        lines.append("")
        lines.append("### Catatan Penting:")
        lines.append("")
        lines.append("> Laporan ini merupakan hasil **simulasi** vulnerability assessment "
                     "dan bukan hasil scanning/testing aktif terhadap sistem produksi. "
                     "Temuan-temuan disusun berdasarkan best practices keamanan untuk "
                     "trading system dan common vulnerabilities pada infrastruktur sejenis. "
                     "Disarankan untuk melakukan penetration testing aktif oleh pihak "
                     "ketiga yang independen untuk memvalidasi temuan simulasi ini.")
        lines.append("")
        lines.append("---")
        lines.append("")
        lines.append(f"*Laporan ini dihasilkan pada {self.timestamp.strftime('%d %B %Y %H:%M:%S WIB')} "
                     f"menggunakan JATS-VA-SIM Framework v1.0.0*")
        lines.append("")
        lines.append("*Dokumen ini bersifat RAHASIA dan hanya untuk penggunaan internal "
                     "Tim Audit IT & Cybersecurity Bursa Efek Indonesia.*")

        return "\n".join(lines)

    def _generate_key_observations(self, stats: Dict) -> List[str]:
        """Generate key observations untuk executive summary."""
        observations = []

        if stats["critical_count"] > 0:
            observations.append(
                f"Ditemukan {stats['critical_count']} kerentanan dengan severity CRITICAL "
                f"yang memerlukan penanganan SEGERA, terutama pada matching engine "
                f"integrity, FIX protocol security, dan remote trading gateway."
            )

        if stats["vulnerable_count"] > 0:
            observations.append(
                f"{stats['vulnerable_count']} dari {stats['total_findings']} temuan "
                f"berstatus VULNERABLE - mengindikasikan kerentanan aktif yang "
                f"dapat dieksploitasi."
            )

        observations.append(
            "Keamanan protokol FIX memerlukan perhatian khusus - message validation, "
            "replay protection, dan session management memiliki kelemahan signifikan."
        )

        observations.append(
            "Remote Trading (RT) gateway sebagai entry point utama broker memiliki "
            "beberapa kelemahan pada enkripsi koneksi dan mekanisme otentikasi."
        )

        observations.append(
            "Infrastruktur monitoring dan incident response belum memadai untuk "
            "melindungi critical financial infrastructure - SIEM coverage dan "
            "SOC capability perlu ditingkatkan."
        )

        observations.append(
            "Kepatuhan terhadap POJK 38/2016 dan UU PDP memerlukan peningkatan "
            "terutama pada aspek DR testing, incident reporting, dan data privacy."
        )

        return observations

    def _generate_compliance_mapping(self) -> Dict:
        """Generate compliance mapping dari findings."""
        mapping = {}
        for f in self.findings:
            for ref in f.compliance_refs:
                # Extract framework name
                if ref.startswith("ISO"):
                    framework = "ISO 27001 / ISO 22301"
                elif ref.startswith("OJK") or ref.startswith("SEOJK"):
                    framework = "OJK POJK 38/2016"
                elif ref.startswith("PCI"):
                    framework = "PCI DSS v4.0"
                elif ref.startswith("NIST"):
                    framework = "NIST Framework"
                elif ref.startswith("CIS"):
                    framework = "CIS Benchmarks"
                elif ref.startswith("OWASP"):
                    framework = "OWASP Standards"
                elif ref.startswith("UU"):
                    framework = "UU PDP Indonesia"
                elif ref.startswith("IOSCO"):
                    framework = "IOSCO Principles"
                else:
                    framework = "Other Standards"

                if framework not in mapping:
                    mapping[framework] = {}
                if ref not in mapping[framework]:
                    mapping[framework][ref] = []
                mapping[framework][ref].append(f.check_id)

        return mapping

    def _generate_remediation_roadmap(self) -> List[Dict]:
        """Generate roadmap remediasi berdasarkan prioritas."""
        phases = [
            {
                "phase": "Phase 1",
                "title": "Emergency Remediation (Remediasi Darurat)",
                "timeframe": "0-30 hari",
                "items": [],
            },
            {
                "phase": "Phase 2",
                "title": "Critical Fixes (Perbaikan Kritis)",
                "timeframe": "30-90 hari",
                "items": [],
            },
            {
                "phase": "Phase 3",
                "title": "Security Improvement (Peningkatan Keamanan)",
                "timeframe": "90-180 hari",
                "items": [],
            },
            {
                "phase": "Phase 4",
                "title": "Strategic Hardening (Penguatan Strategis)",
                "timeframe": "180-365 hari",
                "items": [],
            },
        ]

        for f in self.findings:
            item = {
                "check_id": f.check_id,
                "title": f.title,
                "risk_score": f.risk_score,
                "severity": f.severity.value,
                "effort": f.remediation_effort,
                "priority": "",
            }

            if f.severity == Severity.CRITICAL and f.risk_score >= 9.0:
                item["priority"] = "P1 - Immediate"
                phases[0]["items"].append(item)
            elif f.severity == Severity.CRITICAL:
                item["priority"] = "P1 - Critical"
                phases[1]["items"].append(item)
            elif f.severity == Severity.HIGH and f.risk_score >= 7.5:
                item["priority"] = "P2 - High"
                phases[1]["items"].append(item)
            elif f.severity == Severity.HIGH:
                item["priority"] = "P2 - High"
                phases[2]["items"].append(item)
            elif f.severity == Severity.MEDIUM:
                item["priority"] = "P3 - Medium"
                phases[2]["items"].append(item)
            else:
                item["priority"] = "P4 - Low"
                phases[3]["items"].append(item)

        return phases

    def save_reports(self, output_dir: str) -> Tuple[str, str]:
        """Save reports ke filesystem."""
        os.makedirs(output_dir, exist_ok=True)

        timestamp_str = self.timestamp.strftime("%Y%m%d_%H%M%S")

        # Save JSON
        json_path = os.path.join(output_dir, f"jats_nextg_va_report_{timestamp_str}.json")
        json_report = self.generate_json_report()
        with open(json_path, "w", encoding="utf-8") as f:
            json.dump(json_report, f, indent=2, ensure_ascii=False)

        # Save Markdown
        md_path = os.path.join(output_dir, f"jats_nextg_va_report_{timestamp_str}.md")
        md_report = self.generate_markdown_report()
        with open(md_path, "w", encoding="utf-8") as f:
            f.write(md_report)

        return json_path, md_path
