#!/usr/bin/env node
/**
 * Converte Markdown preprocessado em PDF via Chromium (puppeteer-core).
 * Uso: node render-pdf.mjs input.md output.pdf
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import puppeteer from "puppeteer-core";
import { marked } from "marked";
import matter from "gray-matter";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const mdPath = process.argv[2];
const outPath = process.argv[3];
if (!mdPath || !outPath) {
  console.error("uso: node render-pdf.mjs <input.md> <output.pdf>");
  process.exit(1);
}

const escapeHtml = (text) =>
  String(text)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");

const chromeCandidates = [
  process.env.PUPPETEER_EXECUTABLE_PATH,
  "/usr/bin/google-chrome",
  "/usr/bin/google-chrome-stable",
  "/opt/google/chrome/chrome",
  "/usr/bin/chromium",
  "/usr/bin/chromium-browser",
].filter(Boolean);

const chromePath = chromeCandidates.find((p) => fs.existsSync(p));
if (!chromePath) {
  console.error(
    "Chrome/Chromium não encontrado. Defina PUPPETEER_EXECUTABLE_PATH ou instale google-chrome.",
  );
  process.exit(1);
}

const css = fs.readFileSync(path.join(__dirname, "manual.css"), "utf8");
const raw = fs.readFileSync(mdPath, "utf8");
const { content: md, data } = matter(raw);
const docTitle = escapeHtml(data.title || "Manual do Signal Admin");
const docSubtitle = escapeHtml(
  data.subtitle || "Back-office de clientes e instalações Signal",
);
const headerLabel = escapeHtml(data.header || data.title || "Signal Admin");

marked.setOptions({ headerIds: true, mangle: false });
const body = marked.parse(md);

const coverMeta = [
  data.author ? `<span class="cover-meta">${escapeHtml(data.author)}</span>` : "",
  data.date ? `<span class="cover-meta">${escapeHtml(data.date)}</span>` : "",
]
  .filter(Boolean)
  .join("\n        ");

const coverMetaBlock = coverMeta
  ? `<div class="cover-meta-group">\n        ${coverMeta}\n      </div>`
  : "";

const cover = `<section class="cover">
  <div class="cover-accent"></div>
  <div class="cover-content">
    <p class="cover-kicker">Documentação técnica</p>
    <h1>${docTitle}</h1>
    <p class="cover-subtitle">${docSubtitle}</p>
    ${coverMetaBlock}
  </div>
</section>`;

const html = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <title>${docTitle}</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
  <style>${css}</style>
</head>
<body>${cover}${body}</body>
</html>`;

const browser = await puppeteer.launch({
  executablePath: chromePath,
  headless: true,
  args: ["--no-sandbox", "--disable-setuid-sandbox", "--disable-dev-shm-usage"],
});

try {
  const page = await browser.newPage();
  await page.setContent(html, { waitUntil: "networkidle0" });
  await page.evaluate(() => document.fonts.ready);
  await page.emulateMediaType("print");
  await page.pdf({
    path: outPath,
    format: "A4",
    printBackground: true,
    preferCSSPageSize: false,
    margin: { top: "24mm", bottom: "26mm", left: "20mm", right: "20mm" },
    displayHeaderFooter: true,
    headerTemplate: `<div style="width:100%;padding:0 20mm;font-family:Inter,sans-serif;">
      <div style="border-top:2px solid #4f46e5;width:32px;margin-bottom:6px;"></div>
      <span style="font-size:7pt;color:#94a3b8;letter-spacing:0.04em;text-transform:uppercase;">${headerLabel}</span>
    </div>`,
    footerTemplate: `<div style="width:100%;padding:0 20mm;font-family:Inter,sans-serif;display:flex;justify-content:space-between;align-items:center;border-top:1px solid #e2e8f0;padding-top:6px;">
      <span style="font-size:7pt;color:#94a3b8;">Signal Admin</span>
      <span style="font-size:7pt;color:#64748b;"><span class="pageNumber"></span> / <span class="totalPages"></span></span>
    </div>`,
  });
} finally {
  await browser.close();
}

console.log(`PDF gerado: ${outPath}`);
