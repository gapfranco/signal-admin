#!/usr/bin/env node
/**
 * Insere sumário HTML a partir dos headings do Markdown.
 * Uso: node preprocess.mjs input.md > output.md
 */
import fs from "node:fs";
import matter from "gray-matter";

const inputPath = process.argv[2];
if (!inputPath) {
  console.error("uso: node preprocess.mjs <arquivo.md>");
  process.exit(1);
}

const escapeHtml = (text) =>
  text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");

const slug = (text) =>
  text
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^\w\s-]/g, "")
    .trim()
    .replace(/\s+/g, "-");

const source = fs.readFileSync(inputPath, "utf8");
const { content, data } = matter(source);

const headings = [];
for (const line of content.split("\n")) {
  const match = /^(#{2,3})\s+(.+)$/.exec(line);
  if (!match) continue;
  const level = match[1].length;
  const title = match[2].replace(/\*\*/g, "").trim();
  headings.push({ level, title, id: slug(title) });
}

let toc = "";
if (headings.length > 0 && data.toc !== false) {
  const entries = headings
    .map((h) => {
      const cls = h.level === 3 ? "toc-h3" : "toc-h2";
      return `    <a class="toc-entry ${cls}" href="#${h.id}"><span class="toc-label">${escapeHtml(h.title)}</span></a>`;
    })
    .join("\n");

  toc = [
    '<div class="toc-section">',
    "",
    "  <h2>Sumário</h2>",
    "",
    '  <nav class="toc-list">',
    entries,
    "  </nav>",
    "",
    "</div>",
    "",
  ].join("\n");
}

let body = content.replace(/^#\s+.+\n+/m, "");
const leadMatch = /^([\s\S]*?)(?=\n## )/.exec(body);
if (leadMatch?.[1]?.trim()) {
  const lead = leadMatch[1].trim();
  const rest = body.slice(leadMatch[0].length);
  body = `<p class="lead">${escapeHtml(lead)}</p>\n\n${rest}`;
}

const output = matter.stringify(toc + body, data);
process.stdout.write(output);
