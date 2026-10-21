# Terminal demo

The [README animation](git-why-demo.gif) renders stdout captured from the published git-why v0.1.0 Windows amd64 binary against this repository. It animates command typing and output display; it is not a screen recording or a latency benchmark.

The three scenes show line attribution, function history, and `--llm` context. Long lines wrap for readability. Scenes display up to 11 wrapped output lines, with an explicit excerpt label when content is omitted. Full captured stdout, commands, binary version, and repository commit are preserved in [demo-session.json](demo-session.json).

## Reproduce

Install Pillow in your Python environment, then run from the repository root:

```bash
python -m pip install Pillow
python scripts/render_demo.py --binary /absolute/path/to/git-why
```

Use `git-why.exe` on Windows. The renderer looks for Consolas on Windows, DejaVu Sans Mono on Linux, or Menlo on macOS. Supply `--font /path/to/monospace.ttf` to choose another font.

The script executes the binary, captures new output, and writes `assets/git-why-demo.gif` and `assets/demo-session.json`. A preview frame is saved under the ignored `tmp/` directory. Regenerating against another revision or binary can change the output; the session file records those inputs. If the binary version changes, update the README caption too.
