"""Capture real CLI output and render a readable README terminal GIF.

Requires Python 3.10+ and Pillow. Run from the repository root:
python scripts/render_demo.py --binary /path/to/git-why
"""

import argparse
import json
import os
from pathlib import Path
import subprocess
import textwrap

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / "assets"
WIDTH, HEIGHT = 1200, 680
BG, PANEL = "#0e1922", "#14232e"
TEXT, MUTED, ACCENT = "#edf0e9", "#9fadb8", "#99e2bd"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", required=True, type=Path)
    parser.add_argument("--font", type=Path)
    args = parser.parse_args()
    binary = args.binary.resolve()
    candidates = [
        args.font,
        Path("C:/Windows/Fonts/consola.ttf"),
        Path("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"),
        Path("/System/Library/Fonts/Menlo.ttc"),
    ]
    font_path = next((p for p in candidates if p and p.exists()), None)
    if font_path is None:
        parser.error("Pass --font with the path to a monospace TrueType font")
    font = ImageFont.truetype(str(font_path), 22)
    label_font = ImageFont.truetype(str(font_path), 18)
    title_font = ImageFont.truetype(str(font_path), 29)

    env = os.environ.copy()
    # Trust only this checkout, without changing the user's Git configuration.
    config_index = int(env.get("GIT_CONFIG_COUNT", "0"))
    env[f"GIT_CONFIG_KEY_{config_index}"] = "safe.directory"
    env[f"GIT_CONFIG_VALUE_{config_index}"] = ROOT.as_posix()
    env["GIT_CONFIG_COUNT"] = str(config_index + 1)

    def capture(command):
        return subprocess.run(
            command, cwd=ROOT, env=env, check=True, capture_output=True,
            text=True, encoding="utf-8", timeout=30,
        ).stdout.strip()

    target = "internal/render/ai_renderer.go"
    scenarios = [
        ("Trace a line", [f"{target}:20", "--plain"]),
        ("Follow a function", [target, "--fn", "GetContext", "--short"]),
        ("Give agents context", [f"{target}:20", "--llm"]),
    ]
    version = capture([str(binary), "--version"])
    revision = capture(["git", "rev-parse", "HEAD"])
    scenes = [
        {"title": title, "argv": argv, "stdout": capture([str(binary), *argv])}
        for title, argv in scenarios
    ]
    ASSETS.mkdir(exist_ok=True)
    (ASSETS / "demo-session.json").write_text(
        json.dumps({"version": version, "repository_commit": revision,
                    "scenes": scenes}, indent=2) + "\n", encoding="utf-8",
    )

    frames, durations = [], []

    def draw_frame(index, command, output, excerpt=False):
        canvas = Image.new("RGB", (WIDTH, HEIGHT), BG)
        draw = ImageDraw.Draw(canvas)
        draw.text((44, 29), "git-why", fill=TEXT, font=title_font)
        draw.text((WIDTH - 265, 39), "CLI / REAL OUTPUT", fill=MUTED, font=label_font)
        draw.text((44, 88), f"0{index + 1} / {scenes[index]['title']}",
                  fill=ACCENT, font=title_font)
        draw.rounded_rectangle((30, 146, WIDTH - 30, 604), radius=16, fill=PANEL)
        for x, color in [(57, "#df827c"), (82, "#dbbf84"), (107, ACCENT)]:
            draw.ellipse((x, 164, x + 10, 174), fill=color)
        draw.text((WIDTH - 224, 159), "~/git-why", fill=MUTED, font=label_font)
        draw.line((48, 190, WIDTH - 48, 190), fill="#293a45")
        draw.text((53, 211), "$ " + command, fill=ACCENT, font=font)
        for row, line in enumerate(output):
            draw.text((53, 260 + row * 29), line, fill=TEXT, font=font)
        note = "Output excerpt; the full result includes the commit diff." if excerpt else f"Captured CLI output / {version.split(' (')[0]}"
        draw.text((44, 626), note, fill=MUTED, font=label_font)
        for dot in range(3):
            x = WIDTH - 115 + dot * 25
            draw.ellipse((x, 631, x + 8, 639), fill=ACCENT if dot == index else "#344650")
        # A shared palette avoids flashing between terminal frames.
        return canvas.quantize(palette=palette, dither=Image.Dither.NONE)

    colors = [BG, PANEL, TEXT, MUTED, ACCENT, "#df827c", "#dbbf84", "#293a45", "#344650"]
    palette = Image.new("P", (1, 1))
    rgb = [int(color[i:i + 2], 16) for color in colors for i in (1, 3, 5)]
    palette.putpalette(rgb + [0] * (768 - len(rgb)))

    for index, scene in enumerate(scenes):
        command = "git-why " + " ".join(scene["argv"])
        if font.getlength("$ " + command) > WIDTH - 106:
            raise ValueError("Command exceeds the terminal width")
        lines = []
        for line in scene["stdout"].splitlines():
            lines.extend(textwrap.wrap(line, width=82, subsequent_indent="  ",
                                       replace_whitespace=False) or [""])
        excerpt = len(lines) > 11
        lines = lines[:11]
        for chars in range(0, len(command), 4):
            frames.append(draw_frame(index, command[:chars] + "_", []))
            durations.append(65)
        frames.append(draw_frame(index, command, []))
        durations.append(250)
        for count in range(1, len(lines) + 1):
            frames.append(draw_frame(index, command, lines[:count], excerpt))
            durations.append(90)
        frames.append(draw_frame(index, command, lines, excerpt))
        durations.append(4700 if index == 1 else 3700)

    gif_path = ASSETS / "git-why-demo.gif"
    frames[0].save(gif_path, save_all=True, append_images=frames[1:],
                   duration=durations, loop=0, optimize=True, disposal=1)
    # Keep a final frame for quick visual QA, outside the published assets.
    preview = ROOT / "tmp" / "demo-preview.png"
    preview.parent.mkdir(exist_ok=True)
    frames[-1].convert("RGB").save(preview)
    with Image.open(gif_path) as gif:
        print(f"Saved {gif_path.name}: {gif.n_frames} frames, {gif_path.stat().st_size:,} bytes")
        print(f"Duration: {sum(durations) / 1000:.1f}s; source: {revision[:7]}")


if __name__ == "__main__":
    main()
