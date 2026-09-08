#!/usr/bin/env python3
"""WITS Level 0 frame generator with PTY/serial output for local edge testing.

Usage:
  python wits_generator.py                  # create PTY, print SERIAL_PORT=...
  python wits_generator.py --port /dev/pts/N
"""

from __future__ import annotations

import argparse
import math
import os
import random
import sys
import time
from datetime import datetime, timezone
from typing import Protocol

import serial
from rich.console import Console
from rich.live import Live
from rich.table import Table

from wits_ids import (
    FRAME_END,
    FRAME_START,
    ITEM_BIT_DEPTH,
    ITEM_GAMMA_RAY,
    ITEM_ROP,
    ITEM_WOB,
)


class FrameWriter(Protocol):
    def write(self, data: bytes) -> int: ...
    def flush(self) -> None: ...
    def close(self) -> None: ...


class SerialWriter:
    def __init__(self, port: str, baud: int) -> None:
        self._ser = serial.Serial(port, baudrate=baud, timeout=0)

    def write(self, data: bytes) -> int:
        return int(self._ser.write(data) or 0)

    def flush(self) -> None:
        self._ser.flush()

    def close(self) -> None:
        self._ser.close()


class PtyWriter:
    """Write WITS frames to the master side of a PTY; edge opens the slave path."""

    def __init__(self) -> None:
        self.master_fd, self.slave_fd = os.openpty()
        self.slave_name = os.ttyname(self.slave_fd)
        # Non-blocking master avoids hang if no reader yet.
        os.set_blocking(self.master_fd, False)

    def write(self, data: bytes) -> int:
        try:
            return os.write(self.master_fd, data)
        except BlockingIOError:
            return 0

    def flush(self) -> None:
        return

    def close(self) -> None:
        for fd in (self.master_fd, self.slave_fd):
            try:
                os.close(fd)
            except OSError:
                pass


def synthesize(t: float, well_id: str) -> dict[str, float | str]:
    bit_depth = 1000.0 + t * 0.15 + random.uniform(-0.05, 0.05)
    rop = 25.0 + 5.0 * math.sin(t / 30.0) + random.uniform(-1.0, 1.0)
    wob = 15.0 + 2.0 * math.sin(t / 17.0) + random.uniform(-0.5, 0.5)
    gamma_ray = 80.0 + 20.0 * math.sin(t / 45.0) + random.uniform(-3.0, 3.0)
    return {
        "well_id": well_id,
        "bit_depth": round(bit_depth, 2),
        "rop": round(max(rop, 0.0), 2),
        "wob": round(max(wob, 0.0), 2),
        "gamma_ray": round(max(gamma_ray, 0.0), 2),
    }


def to_wits_frame(sample: dict[str, float | str]) -> str:
    return (
        f"{FRAME_START}\n"
        f"{ITEM_BIT_DEPTH}{sample['bit_depth']:.2f}\n"
        f"{ITEM_ROP}{sample['rop']:.2f}\n"
        f"{ITEM_WOB}{sample['wob']:.2f}\n"
        f"{ITEM_GAMMA_RAY}{sample['gamma_ray']:.2f}\n"
        f"{FRAME_END}\n"
    )


def sample_table(sample: dict[str, float | str], frame: str, port: str) -> Table:
    table = Table(title="WITS Generator — Live Telemetry")
    table.add_column("Field", style="cyan")
    table.add_column("Value", style="green")
    table.add_row("time (UTC)", datetime.now(timezone.utc).isoformat())
    table.add_row("port", port)
    table.add_row("well_id", str(sample["well_id"]))
    table.add_row("bit_depth", f"{sample['bit_depth']}")
    table.add_row("rop", f"{sample['rop']}")
    table.add_row("wob", f"{sample['wob']}")
    table.add_row("gamma_ray", f"{sample['gamma_ray']}")
    table.add_row("frame", frame.replace("\n", "\\n"))
    return table


def open_writer(port: str | None, baud: int) -> tuple[FrameWriter, str]:
    if port:
        return SerialWriter(port, baud), port
    pty = PtyWriter()
    return pty, pty.slave_name


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="WITS Level 0 telemetry generator")
    p.add_argument("--interval", type=float, default=1.0, help="seconds between frames")
    p.add_argument("--well-id", default="WELL-DEMO-01", help="well id (stamped by edge)")
    p.add_argument("--port", default=None, help="serial/PTY path; omit to create a new PTY")
    p.add_argument("--baud", type=int, default=9600, help="baud rate")
    return p.parse_args()


def main() -> None:
    args = parse_args()
    console = Console()

    try:
        writer, port_path = open_writer(args.port, args.baud)
    except Exception as exc:  # noqa: BLE001
        console.print(f"[red]Failed to open serial/PTY:[/red] {exc}")
        sys.exit(1)

    console.print("[bold]WITS generator started[/bold] (Ctrl+C to stop)")
    console.print(f"[bold yellow]SERIAL_PORT={port_path}[/bold yellow]")
    console.print(f"Run edge with: SERIAL_PORT={port_path} make run-edge")

    start = time.monotonic()
    try:
        with Live(console=console, refresh_per_second=4) as live:
            while True:
                t = time.monotonic() - start
                sample = synthesize(t, args.well_id)
                frame = to_wits_frame(sample)
                writer.write(frame.encode("ascii"))
                writer.flush()
                live.update(sample_table(sample, frame, port_path))
                time.sleep(args.interval)
    except KeyboardInterrupt:
        console.print("\n[yellow]Stopped.[/yellow]")
    finally:
        writer.close()


if __name__ == "__main__":
    main()
