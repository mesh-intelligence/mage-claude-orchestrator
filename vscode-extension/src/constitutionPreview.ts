// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// prd: prd006-vscode-extension

import * as fs from "fs";
import * as path from "path";
import * as yaml from "js-yaml";
import * as vscode from "vscode";

/** A single tagged section from a constitution YAML file. */
export interface ConstitutionSection {
  tag: string;
  title: string;
  content: string;
}

/**
 * Converts a slice of ConstitutionSection values into a markdown string.
 * Each section becomes a level-2 heading (## Title), followed by a blank
 * line, the section content, and a trailing blank line.
 *
 * Mirrors ConstitutionToMarkdown in pkg/orchestrator/constitution_md.go.
 */
export function constitutionToMarkdown(sections: ConstitutionSection[]): string {
  return sections
    .map((s) => `## ${s.title}\n\n${s.content.trimEnd()}\n\n`)
    .join("");
}

/**
 * Renders a list of ConstitutionSection values as a styled HTML document
 * using VS Code CSS variables so it matches the editor theme. Each section
 * appears as a level-2 heading followed by its prose content.
 */
export function renderConstitutionHtml(
  title: string,
  sections: ConstitutionSection[]
): string {
  const sectionHtml = sections
    .map(
      (s) =>
        `  <div class="section">\n` +
        `    <h2>${escapeHtml(s.title)}</h2>\n` +
        `    <pre>${escapeHtml(s.content.trimEnd())}</pre>\n` +
        `  </div>`
    )
    .join("\n");

  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${escapeHtml(title)}</title>
  <style>
    body {
      font-family: var(--vscode-font-family);
      color: var(--vscode-foreground);
      background: var(--vscode-editor-background);
      padding: 16px;
      font-size: var(--vscode-font-size);
      max-width: 900px;
    }
    h1 { font-size: 1.4em; margin-bottom: 24px; }
    h2 { font-size: 1.1em; margin-top: 24px; margin-bottom: 8px; }
    pre {
      font-family: var(--vscode-font-family);
      font-size: var(--vscode-font-size);
      white-space: pre-wrap;
      margin: 0 0 16px 0;
      line-height: 1.5;
    }
    .section { margin-bottom: 16px; }
  </style>
</head>
<body>
  <h1>${escapeHtml(title)}</h1>
${sectionHtml}
</body>
</html>`;
}

/** Escapes HTML special characters to prevent XSS in rendered content. */
function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

/**
 * Manages a singleton WebviewPanel that displays a constitution YAML file
 * as rendered markdown. Calling show() with a new URI replaces the panel
 * content in place; a second call while the panel is visible brings it to
 * the foreground.
 */
export class ConstitutionPreview {
  private panel: vscode.WebviewPanel | undefined;

  /**
   * Opens (or refreshes) the preview panel for the given URI. Reads the
   * YAML file, extracts the sections field, and renders it as HTML. Shows
   * an error message when the file is unreadable or has no sections.
   */
  show(uri: vscode.Uri): void {
    const filePath = uri.fsPath;
    const fileName = path.basename(filePath);

    let sections: ConstitutionSection[];
    try {
      const raw = fs.readFileSync(filePath, "utf-8");
      const parsed = yaml.load(raw) as { sections?: ConstitutionSection[] };
      sections = parsed?.sections ?? [];
    } catch (err) {
      vscode.window.showErrorMessage(
        `Cobbler: failed to read ${fileName}: ${err}`
      );
      return;
    }

    if (sections.length === 0) {
      vscode.window.showErrorMessage(
        `Cobbler: ${fileName} has no sections field`
      );
      return;
    }

    if (this.panel) {
      this.panel.reveal(vscode.ViewColumn.Beside);
    } else {
      this.panel = vscode.window.createWebviewPanel(
        "mageOrchestrator.constitutionPreview",
        fileName,
        vscode.ViewColumn.Beside,
        { enableScripts: false }
      );
      this.panel.onDidDispose(() => {
        this.panel = undefined;
      });
    }

    this.panel.title = fileName;
    this.panel.webview.html = renderConstitutionHtml(fileName, sections);
  }

  /** Disposes the preview panel if it is currently open. */
  dispose(): void {
    this.panel?.dispose();
  }
}
