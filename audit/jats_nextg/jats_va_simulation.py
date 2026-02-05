#!/usr/bin/env python3
"""
JATS-NextG Vulnerability Assessment Simulation
================================================

Engine simulasi Vulnerability Assessment (VA) untuk Sistem Perdagangan
Otomatis Jakarta (JATS)-NextG, Bursa Efek Indonesia (BEI).

Simulasi ini mencakup 8 domain keamanan:
1. Keamanan Jaringan (Network Security)
2. Keamanan Protokol Perdagangan (Protocol Security)
3. Otentikasi & Kontrol Akses (Authentication & Access Control)
4. Keamanan Aplikasi (Application Security)
5. Integritas Data & Enkripsi (Data Integrity & Encryption)
6. Keamanan Infrastruktur (Infrastructure Security)
7. Pemantauan & Respons Insiden (Monitoring & Incident Response)
8. Kepatuhan Regulasi (Regulatory Compliance)

Penggunaan:
    python3 -m audit.jats_nextg.jats_va_simulation [options]

    Options:
        --output-dir DIR   Direktori output untuk laporan (default: ./reports/jats_va)
        --format FORMAT    Format output: all, json, markdown (default: all)
        --domain DOMAIN    Domain spesifik: network, protocol, auth, app, data,
                          infra, monitoring, compliance, all (default: all)
        --summary-only     Tampilkan ringkasan saja tanpa detail
        --json-stdout      Output JSON ke stdout

Disclaimer:
    Framework ini dirancang HANYA untuk simulasi audit keamanan defensif.
    BUKAN penetration testing aktif terhadap sistem produksi.

Author: Tim Audit IT & Cybersecurity BEI
Version: 1.0.0
"""

import argparse
import json
import os
import sys
import time
from datetime import datetime
from typing import List, Dict

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.dirname(
    os.path.abspath(__file__)))))

from audit.jats_nextg.checks.network_security import (
    run_network_security_checks, Finding, Severity, CheckStatus
)
from audit.jats_nextg.checks.protocol_security import run_protocol_security_checks
from audit.jats_nextg.checks.authentication import run_authentication_checks
from audit.jats_nextg.checks.application_security import run_application_security_checks
from audit.jats_nextg.checks.data_integrity import run_data_integrity_checks
from audit.jats_nextg.checks.infrastructure import run_infrastructure_checks
from audit.jats_nextg.checks.monitoring import run_monitoring_checks
from audit.jats_nextg.checks.compliance import run_compliance_checks
from audit.jats_nextg.report_generator import JASTReportGenerator


# ANSI color codes for terminal output
class Colors:
    RESET = "\033[0m"
    BOLD = "\033[1m"
    RED = "\033[91m"
    GREEN = "\033[92m"
    YELLOW = "\033[93m"
    BLUE = "\033[94m"
    MAGENTA = "\033[95m"
    CYAN = "\033[96m"
    WHITE = "\033[97m"
    BG_RED = "\033[41m"
    BG_GREEN = "\033[42m"
    BG_YELLOW = "\033[43m"


SEVERITY_COLORS = {
    Severity.CRITICAL: Colors.BG_RED + Colors.WHITE,
    Severity.HIGH: Colors.RED,
    Severity.MEDIUM: Colors.YELLOW,
    Severity.LOW: Colors.BLUE,
    Severity.INFO: Colors.CYAN,
}

STATUS_COLORS = {
    CheckStatus.VULNERABLE: Colors.RED,
    CheckStatus.WARNING: Colors.YELLOW,
    CheckStatus.SECURE: Colors.GREEN,
    CheckStatus.NOT_TESTED: Colors.CYAN,
    CheckStatus.PARTIAL: Colors.MAGENTA,
}

BANNER = f"""{Colors.CYAN}
================================================================================
     ____  _____  ___  ____        _   _            _   ____
    |    |/ ___ \\|   ||  __|      | \\ | | _____ __ | | / ___|
    |    / /___\\ \\   || |__  _____|  \\| |/ _ \\ \\/ /| || |  __
 _  |   | |   | |   ||__  ||_____|   .` |  __/>  < | || | |_ |
| |_|   | \\___/ |   | __| |      | |\\  |\\___/_/\\_\\|_| \\ \\__|
|_______|\\______/|___||____|      |_| \\_|              \\_____|

   VULNERABILITY ASSESSMENT SIMULATION FRAMEWORK v1.0.0
================================================================================
{Colors.RESET}
{Colors.BOLD}  Sistem Perdagangan Otomatis Jakarta (JATS)-NextG
  Bursa Efek Indonesia (BEI) - Tim Audit IT & Cybersecurity
{Colors.RESET}
{Colors.YELLOW}  DISCLAIMER: Simulasi ini BUKAN penetration testing aktif.
  Hanya untuk defensive security assessment.{Colors.RESET}
================================================================================
"""

DOMAIN_MAP = {
    "network": ("Keamanan Jaringan", run_network_security_checks),
    "protocol": ("Keamanan Protokol Perdagangan", run_protocol_security_checks),
    "auth": ("Otentikasi & Kontrol Akses", run_authentication_checks),
    "app": ("Keamanan Aplikasi", run_application_security_checks),
    "data": ("Integritas Data & Enkripsi", run_data_integrity_checks),
    "infra": ("Keamanan Infrastruktur", run_infrastructure_checks),
    "monitoring": ("Pemantauan & Respons Insiden", run_monitoring_checks),
    "compliance": ("Kepatuhan Regulasi", run_compliance_checks),
}


def print_separator():
    """Print line separator."""
    print(f"{Colors.CYAN}{'─' * 80}{Colors.RESET}")


def print_domain_header(domain_name: str, domain_key: str):
    """Print header untuk setiap domain audit."""
    print()
    print(f"{Colors.BOLD}{Colors.CYAN}{'═' * 80}{Colors.RESET}")
    print(f"{Colors.BOLD}{Colors.WHITE}  [{domain_key.upper()}] {domain_name}{Colors.RESET}")
    print(f"{Colors.CYAN}{'═' * 80}{Colors.RESET}")
    print()


def print_finding_summary(finding: Finding):
    """Print ringkasan satu finding."""
    sev_color = SEVERITY_COLORS.get(finding.severity, Colors.WHITE)
    stat_color = STATUS_COLORS.get(finding.status, Colors.WHITE)

    print(f"  {sev_color}[{finding.severity.value:8s}]{Colors.RESET} "
          f"{finding.check_id:10s} "
          f"{stat_color}{finding.status.value:12s}{Colors.RESET} "
          f"Risk:{Colors.BOLD}{finding.risk_score:4.1f}{Colors.RESET}/10  "
          f"{finding.title}")


def print_finding_detail(finding: Finding):
    """Print detail lengkap satu finding."""
    sev_color = SEVERITY_COLORS.get(finding.severity, Colors.WHITE)
    stat_color = STATUS_COLORS.get(finding.status, Colors.WHITE)

    print()
    print(f"  {sev_color}{Colors.BOLD}[{finding.check_id}] {finding.title}{Colors.RESET}")
    print(f"  Severity: {sev_color}{finding.severity.value}{Colors.RESET} | "
          f"Status: {stat_color}{finding.status.value}{Colors.RESET} | "
          f"Risk Score: {Colors.BOLD}{finding.risk_score}/10.0{Colors.RESET} | "
          f"Effort: {finding.remediation_effort}")
    print(f"  Component: {finding.affected_component}")
    print()
    print(f"  {Colors.BOLD}Evidence:{Colors.RESET}")
    for line in finding.evidence.split("\n"):
        print(f"    {line}")
    print()
    print(f"  {Colors.BOLD}Rekomendasi:{Colors.RESET}")
    for line in finding.recommendation.split("\n"):
        print(f"    {line}")

    if finding.cve_references:
        print(f"\n  {Colors.BOLD}CVE References:{Colors.RESET} {', '.join(finding.cve_references)}")
    if finding.compliance_refs:
        print(f"  {Colors.BOLD}Compliance:{Colors.RESET} {', '.join(finding.compliance_refs)}")
    print_separator()


def print_statistics(findings: List[Finding]):
    """Print statistik keseluruhan."""
    from collections import Counter

    severity_counts = Counter(f.severity for f in findings)
    status_counts = Counter(f.status for f in findings)
    avg_risk = sum(f.risk_score for f in findings) / len(findings) if findings else 0
    max_risk = max((f.risk_score for f in findings), default=0)

    print()
    print(f"{Colors.BOLD}{Colors.CYAN}{'═' * 80}{Colors.RESET}")
    print(f"{Colors.BOLD}{Colors.WHITE}  RINGKASAN STATISTIK VULNERABILITY ASSESSMENT{Colors.RESET}")
    print(f"{Colors.CYAN}{'═' * 80}{Colors.RESET}")
    print()

    print(f"  Total Temuan: {Colors.BOLD}{len(findings)}{Colors.RESET}")
    print(f"  Skor Risiko Rata-rata: {Colors.BOLD}{avg_risk:.1f}/10.0{Colors.RESET}")
    print(f"  Skor Risiko Tertinggi: {Colors.BOLD}{max_risk:.1f}/10.0{Colors.RESET}")
    print()

    print(f"  {Colors.BOLD}Distribusi Severity:{Colors.RESET}")
    for sev in [Severity.CRITICAL, Severity.HIGH, Severity.MEDIUM, Severity.LOW, Severity.INFO]:
        count = severity_counts.get(sev, 0)
        bar = "█" * (count * 3)
        sev_color = SEVERITY_COLORS.get(sev, Colors.WHITE)
        print(f"    {sev_color}{sev.value:10s}{Colors.RESET} {count:3d} {sev_color}{bar}{Colors.RESET}")
    print()

    print(f"  {Colors.BOLD}Distribusi Status:{Colors.RESET}")
    for stat in [CheckStatus.VULNERABLE, CheckStatus.WARNING, CheckStatus.PARTIAL,
                 CheckStatus.SECURE, CheckStatus.NOT_TESTED]:
        count = status_counts.get(stat, 0)
        bar = "█" * (count * 3)
        stat_color = STATUS_COLORS.get(stat, Colors.WHITE)
        print(f"    {stat_color}{stat.value:12s}{Colors.RESET} {count:3d} {stat_color}{bar}{Colors.RESET}")
    print()

    # Overall risk rating
    crit = severity_counts.get(Severity.CRITICAL, 0)
    high = severity_counts.get(Severity.HIGH, 0)
    if crit >= 3:
        rating = f"{Colors.BG_RED}{Colors.WHITE} SANGAT TINGGI (Very High) {Colors.RESET}"
    elif crit >= 1:
        rating = f"{Colors.RED}{Colors.BOLD} TINGGI (High) {Colors.RESET}"
    elif high >= 5:
        rating = f"{Colors.RED}{Colors.BOLD} TINGGI (High) {Colors.RESET}"
    elif high >= 2:
        rating = f"{Colors.YELLOW}{Colors.BOLD} MENENGAH-TINGGI (Medium-High) {Colors.RESET}"
    else:
        rating = f"{Colors.YELLOW} MENENGAH (Medium) {Colors.RESET}"

    print(f"  {Colors.BOLD}Rating Risiko Keseluruhan:{Colors.RESET} {rating}")
    print()

    # Top 5 risks
    sorted_findings = sorted(findings, key=lambda f: -f.risk_score)
    print(f"  {Colors.BOLD}Top 5 Risiko Tertinggi:{Colors.RESET}")
    for i, f in enumerate(sorted_findings[:5], 1):
        sev_color = SEVERITY_COLORS.get(f.severity, Colors.WHITE)
        print(f"    {i}. {sev_color}[{f.severity.value}]{Colors.RESET} "
              f"{f.check_id} - {f.title} "
              f"(Risk: {Colors.BOLD}{f.risk_score}{Colors.RESET})")
    print()


def simulate_progress(message: str, steps: int = 5):
    """Simulate progress indicator."""
    print(f"  {Colors.CYAN}[*]{Colors.RESET} {message}", end="", flush=True)
    for _ in range(steps):
        time.sleep(0.1)
        print(".", end="", flush=True)
    print(f" {Colors.GREEN}Done{Colors.RESET}")


def run_simulation(args) -> List[Finding]:
    """Menjalankan simulasi VA."""
    all_findings = []
    config = {
        "target": "JATS-NextG",
        "scope": "Full System Assessment",
        "date": datetime.now().isoformat(),
    }

    domains_to_run = (
        list(DOMAIN_MAP.keys()) if args.domain == "all"
        else [args.domain]
    )

    print(f"\n{Colors.BOLD}  Memulai simulasi VA untuk {len(domains_to_run)} domain...{Colors.RESET}\n")

    for domain_key in domains_to_run:
        domain_name, check_func = DOMAIN_MAP[domain_key]
        print_domain_header(domain_name, domain_key)

        simulate_progress(f"Menjalankan pemeriksaan {domain_name}")
        findings = check_func(config)
        all_findings.extend(findings)

        print(f"  {Colors.GREEN}[+]{Colors.RESET} {len(findings)} temuan diidentifikasi\n")

        if not args.summary_only:
            for f in findings:
                print_finding_summary(f)
        print()

    return all_findings


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="JATS-NextG Vulnerability Assessment Simulation",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=(
            "Contoh penggunaan:\n"
            "  python3 -m audit.jats_nextg.jats_va_simulation\n"
            "  python3 -m audit.jats_nextg.jats_va_simulation --domain network\n"
            "  python3 -m audit.jats_nextg.jats_va_simulation --format json --output-dir ./reports\n"
            "  python3 -m audit.jats_nextg.jats_va_simulation --summary-only\n"
        )
    )
    parser.add_argument(
        "--output-dir",
        default="./reports/jats_va",
        help="Direktori output untuk laporan (default: ./reports/jats_va)"
    )
    parser.add_argument(
        "--format",
        choices=["all", "json", "markdown"],
        default="all",
        help="Format output (default: all)"
    )
    parser.add_argument(
        "--domain",
        choices=list(DOMAIN_MAP.keys()) + ["all"],
        default="all",
        help="Domain audit spesifik (default: all)"
    )
    parser.add_argument(
        "--summary-only",
        action="store_true",
        help="Tampilkan ringkasan tanpa detail temuan"
    )
    parser.add_argument(
        "--json-stdout",
        action="store_true",
        help="Output JSON report ke stdout"
    )
    parser.add_argument(
        "--no-banner",
        action="store_true",
        help="Jangan tampilkan banner"
    )
    parser.add_argument(
        "--detail",
        action="store_true",
        help="Tampilkan detail lengkap setiap temuan"
    )

    args = parser.parse_args()

    # Print banner
    if not args.no_banner and not args.json_stdout:
        print(BANNER)

    # Run simulation
    if not args.json_stdout:
        start_time = time.time()
        findings = run_simulation(args)
        elapsed = time.time() - start_time
    else:
        config = {"target": "JATS-NextG", "scope": "Full System Assessment"}
        findings = []
        for domain_key, (_, check_func) in DOMAIN_MAP.items():
            if args.domain == "all" or args.domain == domain_key:
                findings.extend(check_func(config))

    # Print detailed findings if requested
    if args.detail and not args.json_stdout:
        print()
        print(f"{Colors.BOLD}{Colors.CYAN}{'═' * 80}{Colors.RESET}")
        print(f"{Colors.BOLD}{Colors.WHITE}  DETAIL TEMUAN{Colors.RESET}")
        print(f"{Colors.CYAN}{'═' * 80}{Colors.RESET}")
        for f in sorted(findings, key=lambda x: (-x.risk_score)):
            print_finding_detail(f)

    # Print statistics
    if not args.json_stdout:
        print_statistics(findings)

    # Generate reports
    if args.json_stdout:
        config_dict = {"target": "JATS-NextG", "scope": "Full System Assessment"}
        generator = JASTReportGenerator(findings, config_dict)
        json_report = generator.generate_json_report()
        print(json.dumps(json_report, indent=2, ensure_ascii=False))
    else:
        config_dict = {"target": "JATS-NextG", "scope": "Full System Assessment"}
        generator = JASTReportGenerator(findings, config_dict)

        if args.format in ("all", "json", "markdown"):
            print(f"\n{Colors.BOLD}  Generating reports...{Colors.RESET}")
            json_path, md_path = generator.save_reports(args.output_dir)

            print(f"  {Colors.GREEN}[+]{Colors.RESET} JSON report: {json_path}")
            print(f"  {Colors.GREEN}[+]{Colors.RESET} Markdown report: {md_path}")

        # Print completion
        print(f"\n{'═' * 80}")
        print(f"{Colors.GREEN}{Colors.BOLD}  Simulasi VA selesai.{Colors.RESET}")
        print(f"  Total temuan: {len(findings)}")
        print(f"  Waktu eksekusi: {elapsed:.1f} detik")
        print(f"  Output: {args.output_dir}/")
        print(f"{'═' * 80}\n")


if __name__ == "__main__":
    main()
