# Thermal docs site

Static, dependency-free HTML documentation for thermal. One file, no build
step, no JavaScript framework, no external requests.

## What is here

```
docs/pages/
├── index.html       The whole site: styles, content, and a little vanilla JS
├── 404.html         A copy of index.html, used by GitHub Pages for unknown paths
├── og.png           Social share card, 1200x630
├── robots.txt       Crawler policy, AI bots named explicitly
├── sitemap.xml      One URL, updated by hand
├── SEO-GEO-AUDIT.md Audit notes, competitors, and the next actions
└── README.md        This file
```

Everything the page needs is inline: CSS in a `<style>` block, JavaScript in a
`<script>` block, the wordmark and favicon as inline SVG or data URIs, and the
system font stack. The page makes no network requests, so it renders the same
offline as it does on a CDN. The only binary asset is `og.png`.

## Deploy to GitHub Pages

Repository settings, Pages, then:

- Source: Deploy from a branch
- Branch: `main`
- Folder: `/docs/pages`

The site lands at `https://<owner>.github.io/<repo>/`. No workflow file is
needed. If you later publish at the repository root, move the two HTML files up
one level and leave the README behind.

## Deploy to Cloudflare Pages

Connect the repository and set:

- Framework preset: None
- Build command: leave empty
- Build output directory: `docs/pages`

Cloudflare serves `index.html` at the root and, because 404.html mirrors it,
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
