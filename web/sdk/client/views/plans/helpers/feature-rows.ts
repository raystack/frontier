export interface FeatureRow {
  title: string;
  info?: string;
}

interface FeatureLike {
  title?: string;
  metadata?: unknown;
}

export type FeatureInfoPart =
  | { type: 'text'; text: string }
  | { type: 'link'; label: string; url: string };

/*
 * Builds the feature table rows.
 * - Plain object keyed by title, so ordering matches the old inline sort exactly.
 * - Missing/falsy weightage falls back to the feature count; last weightage wins.
 * - First non-blank info wins; empty titles are dropped.
 */
export function buildFeatureRows(features: FeatureLike[]): FeatureRow[] {
  const weightages: Record<string, unknown> = {};
  const infos = new Map<string, string>();

  features.forEach(feature => {
    const metadata = (feature.metadata || {}) as Record<string, unknown>;
    const title = feature.title || '';
    weightages[title] = metadata.weightage || features.length;
    const info = typeof metadata.info === 'string' ? metadata.info.trim() : '';
    if (info && !infos.has(title)) infos.set(title, info);
  });

  return Object.entries(weightages)
    .sort((a, b) => Number(a[1]) - Number(b[1]))
    .filter(([title]) => Boolean(title))
    .map(([title]) => ({ title, info: infos.get(title) }));
}

const LINK_PATTERN = /\[([^\]]+)\]\(([^)\s]+)\)/g;
const ALLOWED_URL_PATTERN = /^(https?|mailto):/;

/*
 * Splits info text into text and link parts.
 * - [label](url) becomes a link only with a non-blank label and an http:, https: or mailto: url.
 * - Everything else is kept as plain text exactly as written.
 */
export function parseFeatureInfo(info: string): FeatureInfoPart[] {
  const parts: FeatureInfoPart[] = [];
  let cursor = 0;
  for (const match of info.matchAll(LINK_PATTERN)) {
    const [raw, label, url] = match;
    const start = match.index ?? 0;
    const isLink = label.trim() !== '' && ALLOWED_URL_PATTERN.test(url);
    parts.push({ type: 'text', text: info.slice(cursor, start) });
    parts.push(isLink ? { type: 'link', label, url } : { type: 'text', text: raw });
    cursor = start + raw.length;
  }
  parts.push({ type: 'text', text: info.slice(cursor) });
  return parts;
}

/*
 * Normalises a feature cell value to a string.
 * - Falsy values (missing, '', false, 0) become '-'.
 * - Everything else goes through String(), so lowercasing never throws.
 */
export function normalizeFeatureValue(value: unknown): string {
  return value ? String(value) : '-';
}
