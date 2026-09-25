/**
 * Cloudflare Pages Advanced Mode Worker
 * High-SEO 301 Permanent Redirect Engine for thermal.jadmadi.net -> jadmadi.net/projects/thermal/
 */
export default {
  async fetch(request, env) {
    const url = new URL(request.url);

    // 1. Dedicated installer redirect
    if (url.pathname === "/install.sh") {
      return Response.redirect("https://jadmadi.net/thermal/install.sh", 301);
    }

    // 2. Target URL construction preserving subpaths when relevant
    let target = "https://jadmadi.net/projects/thermal/";
    if (url.pathname === "/share" || url.pathname === "/share.html") {
      target = "https://jadmadi.net/projects/thermal/share" + url.search;
    } else if (url.pathname === "/llms.txt") {
      target = "https://jadmadi.net/projects/thermal/llms.txt";
    } else if (url.pathname === "/llms-full.txt") {
      target = "https://jadmadi.net/projects/thermal/llms-full.txt";
    }

    // 3. High-SEO 301 Permanent Redirect Response
    const headers = new Headers();
    headers.set("Location", target);
    headers.set("Link", `<${target}>; rel="canonical"`);
    headers.set("Cache-Control", "public, max-age=604800, immutable");
    headers.set("Content-Type", "text/html; charset=utf-8");
    headers.set("X-Robots-Tag", "noindex, follow");

    const html = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>301 Moved Permanently - Thermal</title>
  <link rel="canonical" href="${target}">
  <meta http-equiv="refresh" content="0; url=${target}">
  <meta name="robots" content="noindex, follow">
</head>
<body style="font-family:system-ui,-apple-system,sans-serif;background:#0d1117;color:#e6edf3;padding:40px;text-align:center;">
  <h1>Redirecting to <a href="${target}" style="color:#58a6ff;">${target}</a>...</h1>
  <p>Thermal has moved to <a href="${target}" style="color:#58a6ff;">https://jadmadi.net/projects/thermal/</a>.</p>
</body>
</html>`;

    return new Response(html, {
      status: 301,
      statusText: "Moved Permanently",
      headers,
    });
  },
};
