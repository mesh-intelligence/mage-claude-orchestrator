// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

import { describe, it, expect } from "vitest";
import {
  ConstitutionSection,
  constitutionToMarkdown,
  renderConstitutionHtml,
} from "../constitutionPreview";

// ---- constitutionToMarkdown ----

describe("constitutionToMarkdown", () => {
  it("returns empty string for no sections", () => {
    expect(constitutionToMarkdown([])).toBe("");
  });

  it("renders a single section with trailing newline in content", () => {
    const sections: ConstitutionSection[] = [
      { tag: "articles", title: "Core Principles", content: "Five principles govern.\n" },
    ];
    expect(constitutionToMarkdown(sections)).toBe(
      "## Core Principles\n\nFive principles govern.\n\n"
    );
  });

  it("renders a single section without trailing newline in content", () => {
    const sections: ConstitutionSection[] = [
      { tag: "x", title: "Title", content: "No trailing newline" },
    ];
    expect(constitutionToMarkdown(sections)).toBe(
      "## Title\n\nNo trailing newline\n\n"
    );
  });

  it("renders multiple sections in order", () => {
    const sections: ConstitutionSection[] = [
      { tag: "a", title: "First", content: "First content.\n" },
      { tag: "b", title: "Second", content: "Second content.\n" },
    ];
    expect(constitutionToMarkdown(sections)).toBe(
      "## First\n\nFirst content.\n\n## Second\n\nSecond content.\n\n"
    );
  });

  it("collapses extra trailing newlines in content", () => {
    const sections: ConstitutionSection[] = [
      { tag: "s", title: "Heading", content: "Body text.\n\n\n" },
    ];
    expect(constitutionToMarkdown(sections)).toBe(
      "## Heading\n\nBody text.\n\n"
    );
  });

  it("preserves internal newlines in multi-line content", () => {
    const sections: ConstitutionSection[] = [
      { tag: "s", title: "Multi", content: "Line one.\nLine two.\n" },
    ];
    expect(constitutionToMarkdown(sections)).toBe(
      "## Multi\n\nLine one.\nLine two.\n\n"
    );
  });
});

// ---- renderConstitutionHtml ----

describe("renderConstitutionHtml", () => {
  it("produces a DOCTYPE html document", () => {
    const html = renderConstitutionHtml("exec.yaml", []);
    expect(html).toContain("<!DOCTYPE html>");
    expect(html).toContain("<html");
  });

  it("includes the title in an h1 tag", () => {
    const html = renderConstitutionHtml("execution.yaml", []);
    expect(html).toContain("<h1>execution.yaml</h1>");
  });

  it("renders each section as h2 + pre", () => {
    const sections: ConstitutionSection[] = [
      { tag: "a", title: "Articles", content: "Five principles.\n" },
      { tag: "b", title: "Standards", content: "Go code follows...\n" },
    ];
    const html = renderConstitutionHtml("exec.yaml", sections);
    expect(html).toContain("<h2>Articles</h2>");
    expect(html).toContain("<h2>Standards</h2>");
    expect(html).toContain("<pre>Five principles.</pre>");
    expect(html).toContain("<pre>Go code follows...</pre>");
  });

  it("escapes HTML special characters in title and content", () => {
    const sections: ConstitutionSection[] = [
      { tag: "s", title: "<script>", content: "<b>bold</b>\n" },
    ];
    const html = renderConstitutionHtml("<title>", sections);
    expect(html).not.toContain("<script>");
    expect(html).toContain("&lt;script&gt;");
    expect(html).toContain("&lt;title&gt;");
    expect(html).toContain("&lt;b&gt;bold&lt;/b&gt;");
  });

  it("uses VS Code CSS variables", () => {
    const html = renderConstitutionHtml("f.yaml", []);
    expect(html).toContain("var(--vscode-font-family)");
    expect(html).toContain("var(--vscode-foreground)");
    expect(html).toContain("var(--vscode-editor-background)");
  });
});
