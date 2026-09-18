# SEO and GEO audit: thermal docs site

Date: 2026-09-18
Scope: `docs/pages/` (one page, `index.html`, mirrored as `404.html`)
Target URL for canonical tags: `https://jadmadi.github.io/thermal/`

## Status after this pass

| Area | Before | After |
|------------------------|--------|-------|
| Title tag | 50 chars | 50 chars, keyword-led |
| Meta description | 169 chars, truncates | 157 chars |
| H1 count | 1 | 1 |
| Schema markup | none | SoftwareApplication, Person, FAQPage |
| FAQ section | none | 6 questions, visible and marked up |
| Order line in H2 | no | yes ("Supported tools") |
| robots.txt | missing | present, AI bots named explicitly |
| sitemap.xml | missing | present, one canonical URL |
| Social image | none | `og.png`, 1200x630, 47KB |
| External link rel | none | `noopener noreferrer` on all six |
| Load time (local) | 0.04s | 0.01s |
| Horizontal overflow | none 344-1600px | none 360-1600px |

## What the audit script reported

Baseline run found four gaps: no schema, no robots.txt, no sitemap, and a
169-character description. All four are closed. The remaining script output is
clean.

## Deliberate decisions

Two findings from the checklist are intentionally not "fixed":

- The H1 is the tagline, "Don't break the streak," not a keyword phrase.
  The brand voice wins; the keyword work is carried by the title, the
  description, the lede, and the FAQ.
- `--weeks`, `--db`, and `--verbose` are discussed in prose rather than a
  flag table. The reference table covers the flags a reader is most likely
  to reach for.

## Competitive landscape

The category is crowded and growing. Search for the obvious terms surfaces:

| Tool | Stack | Coverage | Positioning |
|-------------|--------------|---------|----------------------------------------|
| ccusage | TypeScript | 18 tools | The reference tool, largest mindshare |
| tokscale | Rust + TS | 22 tools | Leaderboards, 2D/3D graphs |
| Token Tracker | Node/React | 36 tools | Desktop pet, widgets, achievements |
| tku | Go | 9 tools | Live spend monitor, subscription view |
| toktrack | Rust | 5 tools | Speed, survives log deletion |
| goccc | Go | 1 tool | Claude Code only, statusline provider |
| VibeUsage | Node | 7 tools | Cloud sync, public profiles |
| aiusage | TypeScript | 20+ tools | Sync across machines, leaderboard |
| thermal | Go | 12 tools | Contribution heatmap and streaks first |

Where thermal differs, and where the content should push:

1. Heatmap and streaks as the primary view. Every competitor is an accounting
   table first; thermal's default output is the GitHub-style grid.
2. Twelve tools that most trackers skip: Devin, Muse, Droid, command-code,
   Agy, codewhale.
3. A single static binary with no runtime. ccusage needs bun or node,
   tokscale needs a Rust toolchain or a binary, the Python tools need Python.
4. Read-only by construction, with the delta cache inside the tool's own
   directory rather than the source.

## Prioritized actions

1. Deploy. Pages is enabled with the source set to GitHub Actions, and
   `.github/workflows/pages.yml` publishes `docs/pages`. Branch-based Pages
   cannot serve a subfolder other than `/docs`, which is why the workflow
   exists. The workflow runs when this branch merges to `main`. After that,
   submit the sitemap in Search Console and Bing Webmaster Tools.
2. Add comparison pages: "thermal vs ccusage", "thermal vs tokscale", and a
   ccusage-alternatives page. In this category, comparison intent converts and
   the terms are not yet owned by anyone.
3. Meet the category where it searches: cost, tokens, and per-tool queries
   ("open code token usage", "claude code cost tracker"). The FAQ already
   answers the shape of these; a longer form on the blog at jadmadi.net,
   linked from the docs, is the natural next step.
4. Add one real screenshot per view (leaderboard, weekly report, projects) with
   descriptive alt text once the page has a build step for images. The current
   `og.png` is the only raster asset, which keeps the page light.
5. Backlinks. The GitHub README already links here. A Show HN or r/ClaudeAI
   post, plus the ccusage and tokscale issue trackers where multi-tool support
   is discussed, are the realistic first links.

## Notes on the canonical URL

Every canonical reference points at `https://jadmadi.github.io/thermal/`. If
the site moves to a custom domain, update the canonical link, `og:url`,
`og:image`, `twitter:image`, the `Sitemap:` line in `robots.txt`, and the
`<loc>` in `sitemap.xml`. The comment in `index.html` marks the canonical line.
