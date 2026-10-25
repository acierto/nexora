# Nexora documentation

**[Open the Nexora docs website](https://acierto.github.io/nexora/)**

The website includes the user guides, searchable documentation, and an interactive
resolution playground. Start with [Getting started online](https://acierto.github.io/nexora/#start).

Prefer reading on GitHub? Open the [quickstart](getting-started.md),
[knowledge format](knowledge-format.md), [resolution guide](resolution.md),
[CLI reference](cli.md), or [troubleshooting](troubleshooting.md).

## Publishing and previewing

GitHub Pages publishes the `docs/` directory from the `main` branch. The supplied
Nexora Docs HTML is stored as `docs/index.html`; `.nojekyll` keeps it a plain
static site. Push changes to `main` to publish an update. Deployment status is
visible in the repository's **Actions** tab and **Settings → Pages**.

The HTML includes copies of the seven Markdown guides. After changing them, run
this from the repository root before committing:

```sh
python3 scripts/sync_docs.py
python3 scripts/sync_docs.py --check
```

To preview locally, run `python3 -m http.server 8000 --directory docs` from the
repository root, then open **http://localhost:8000/**. Stop the server with Ctrl-C.
Google Fonts, Three.js, and the Markdown renderer load from external CDNs, so
the website needs an internet connection for those assets. The Nexora CLI
continues to work offline.
