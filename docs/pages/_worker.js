/**
 * Cloudflare Pages Advanced Mode Worker
 * Provides Content Negotiation for "Markdown for Agents" and agent discovery headers.
 */
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const accept = request.headers.get("Accept") || "";

    // Content negotiation for Markdown for Agents
    if (
      (accept.includes("text/markdown") || accept.includes("text/x-markdown")) &&
      (url.pathname === "/" || url.pathname === "/index.html")
    ) {
      const mdUrl = new URL("/index.md", request.url);
      const res = await env.ASSETS.fetch(mdUrl);
      if (res.ok) {
        const text = await res.text();
        const headers = new Headers(res.headers);
        headers.set("Content-Type", "text/markdown; charset=utf-8");
        headers.set("Vary", "Accept");
        
        // Approximate token count based on whitespace-delimited words * 1.33
        const wordCount = text.trim().split(/\s+/).length;
        const estTokens = Math.round(wordCount * 1.33);
        headers.set("x-markdown-tokens", estTokens.toString());
        
        return new Response(text, {
          status: 200,
          headers,
        });
      }
    }

    // Default static asset handling
    const response = await env.ASSETS.fetch(request);
    
    // Inject Link header for markdown alternative when serving HTML
    if (url.pathname === "/" || url.pathname === "/index.html") {
      const headers = new Headers(response.headers);
      headers.append(
        "Link",
        '</index.md>; rel="alternate"; type="text/markdown"'
      );
      return new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers,
      });
    }

    return response;
  },
};
