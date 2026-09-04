#!/usr/bin/env python3
"""WITS Level 0 frame generator for local serial/PTY emulation.

Emits synthetic drilling telemetry frames and renders them with rich.
"""

from __future__ import annotations

import math
import random
import time
from datetime import datetime, timezone

from rich.console import Console
from rich.live import Live
from rich.table import Table

WELL_ID = "WELL-DEMO-01"
INTERVAL_S = 1.0


def synthesize(t: float) -> dict[str, float | str]:
    """Generate a plausible synthetic sample at elapsed seconds t."""
    bit_depth = 1000.0 + t * 0.15 + random.uniform(-0.05, 0.05)
    rop = 25.0 + 5.0 * math.sin(t / 30.0) + random.uniform(-1.0, 1.0)
    wob = 15.0 + 2.0 * math.sin(t / 17.0) + random.uniform(-0.5, 0.5)
    gamma_ray = 80.0 + 20.0 * math.sin(t / 45.0) + random.uniform(-3.0, 3.0)
    return {
        "well_id": WELL_ID,
        "bit_depth": round(bit_depth, 2),
        "rop": round(max(rop, 0.0), 2),
        "wob": round(max(wob, 0.0), 2),
        "gamma_ray": round(max(gamma_ray, 0.0), 2),
    }


def to_wits_frame(sample: dict[str, float | str]) -> str:
    """Minimal WITS-like ASCII frame (record 01 style placeholders)."""
    # !!0108 = bit depth, !!0113 = ROP, !!010A = WOB, !!0122 = gamma ray (illustrative)
    return (
        f"&&\n"
        f"0108{sample['bit_depth']:.2f}\n"
        f"0113{sample['rop']:.2f}\n"
        f"010A{sample['wob']:.2f}\n"
        f"0122{sample['gamma_ray']:.2f}\n"
        f"!!\n"
    )


def sample_table(sample: dict[str, float | str], frame: str) -> Table:
    table = Table(title="WITS Generator — Live Telemetry")
    table.add_column("Field", style="cyan")
    table.add_column("Value", style="green")
    table.add_row("time (UTC)", datetime.now(timezone.utc).isoformat())
    table.add_row("well_id", str(sample["well_id"]))
    table.add_row("bit_depth", f"{sample['bit_depth']}")
    table.add_row("rop", f"{sample['rop']}")
    table.add_row("wob", f"{sample['wob']}")
    table.add_row("gamma_ray", f"{sample['gamma_ray']}")
    table.add_row("frame", frame.replace("\n", "\\n"))
    return table


def main() -> None:
    console = Console()
    console.print("[bold]WITS generator started[/bold] (Ctrl+C to stop)")
    start = time.monotonic()
    with Live(console=console, refresh_per_second=4) as live:
        try:
            while True:
                t = time.monotonic() - start
                sample = synthesize(t)
                frame = to_wits_frame(sample)
                live.update(sample_table(sample, frame))
                time.sleep(INTERVAL_S)
        except KeyboardInterrupt:
            console.print("\n[yellow]Stopped.[/yellow]")


if __name__ == "__main__":
    main()
