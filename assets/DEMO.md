# Terminal demo

## README walkthrough

The README features a supplied terminal walkthrough using the Cobra repository. It shows installation, inspecting `command.go`, tracing `SetUsageTemplate`, and requesting `--llm` output.

- [Animated GIF](git-why-terminal-demo.gif): 1280 by 720, approximately 54 seconds, 495 KB.
- [MP4 video](git-why-terminal-demo.mp4): H.264, 1280 by 720, approximately 54 seconds, 484 KB.

The supplied files are preserved unchanged. The walkthrough does not identify a pinned Cobra commit or git-why binary version; the versioned example below provides a separate reproducible reference.

## Reproducible short demo

The [short animation](git-why-demo.gif) renders stdout captured from the published git-why v0.1.0 Windows amd64 binary against this repository. It animates command typing and output display; it is not a screen recording or a latency benchmark.

The three scenes show line attribution, function history, and `--llm` context. Long lines wrap for readability. Scenes display up to 11 wrapped output lines, with an explicit excerpt label when content is omitted. Full captured stdout, commands, binary version, and repository commit are preserved in [demo-session.json](demo-session.json).

### Reproduce

Install Pillow in your Python environment, then run from the repository root:

```bash
python -m pip install Pillow
python scripts/render_demo.py --binary /absolute/path/to/git-why
```

Use `git-why.exe` on Windows. The renderer looks for Consolas on Windows, DejaVu Sans Mono on Linux, or Menlo on macOS. Supply `--font /path/to/monospace.ttf` to choose another font.

The script executes the binary, captures new output, and writes `assets/git-why-demo.gif` and `assets/demo-session.json`. It does not overwrite the supplied README walkthrough. A preview frame is saved under the ignored `tmp/` directory. Regenerating against another revision or binary can change the output; the session file records those inputs.
