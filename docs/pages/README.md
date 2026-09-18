# Thermal docs site

Static, dependency-free HTML documentation for thermal. One file, no build
step, no JavaScript framework, no external requests.

## What is here

```
docs/pages/
├── index.html       The whole site: styles, content, and a little vanilla JS
├── 404.html         A copy of index.html, served for unknown paths
├── og.png           Social share card, 1200x630
├── shot-*.webp      Terminal captures: leaderboard, weekly, projects
├── robots.txt       Crawler policy, AI bots named explicitly
├── sitemap.xml      One URL, updated by hand
├── SEO-GEO-AUDIT.md Audit notes, competitors, and the next actions
└── README.md        This file
```

Everything the page needs is inline: CSS in a `<style>` block, JavaScript in a
`<script>` block, the wordmark and favicon as inline SVG or data URIs, and the
system font stack. The page makes no network requests, so it renders the same
offline as it does on a CDN apart from three screenshots, which are lazy
loaded.

## Deploy to GitHub Pages

Pages is enabled on this repository with the source set to GitHub Actions,
which is what lets the site live in `docs/pages`. Branch-based Pages only
offers `/` or `/docs`, so an Actions workflow in
`.github/workflows/pages.yml` publishes `docs/pages` instead, on every push to
`main` that touches the site.

Nothing else to configure. Merging to `main` deploys; the workflow also has a
manual trigger in the Actions tab. The site lands at
`https://<owner>.github.io/<repo>/`.

If you would rather not use Actions, move the site files to `docs/` at the
repository root and switch Settings, Pages, Source back to "Deploy from a
branch" with folder `/docs`.

## Deploy to Cloudflare Pages

Connect the repository and set the build output directory to `docs/pages`.
Leave the framework preset at None and the build command empty. The GitHub
Actions workflow is not used there, so no extra configuration is needed.
Cloudflare serves `index.html` at the root and, because `404.html` mirrors it,
also serves the documentation for unknown paths.

## Editing

Edit `index.html` directly. Two things to know before you do:

- The heatmap is static markup, generated once and pasted in. The hero panel
  holds 53 weeks by 7 days of `<span class="hm-cell hm-lv0..4">` elements.
  Changing the shape means regenerating that block rather than editing it by
  hand.
- Colour and spacing come from CSS custom properties at the top of the
  `<style>` block. Change a token there rather than a value inside a rule.
  The heat ramp is the one exception worth care: its five steps are the same
  colours the terminal uses for heat levels 1 to 4.

## Checks worth running after an edit

```bash
# Serve locally
python3 -m http.server 8899 --directory docs/pages

# No horizontal overflow at any width
# (open the page, then in the console)
#   document.documentElement.scrollWidth > document.documentElement.clientWidth

# Terminal output keeps its columns and scrolls, tables scroll inside .tbl-wrap
```

Contrast, focus rings, hit targets on the Copy buttons, and
`prefers-reduced-motion` are already handled. Keep them that way.

When the page content changes, update `<lastmod>` in `sitemap.xml`.

The three `shot-*.webp` captures were rendered at ray.so and converted from
PNG. Two things to know if you re-shoot them: ray.so's editor clips lines
wider than 80 columns, so keep samples under that, and the projects capture
uses invented project names on real numbers.
