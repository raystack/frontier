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
 * - Sorted by metadata.weightage; missing/falsy weightage falls back to the feature count.
 * - Duplicate titles collapse into one row: last weightage wins, first non-empty info wins.
 * - Empty titles are dropped.
 */
export function buildFeatureRows(features: FeatureLike[]): FeatureRow[] {
  const totalFeatures = features.length;
  const rows = new Map<string, { weightage: unknown; info?: string }>();

  features.forEach(feature => {
    const metadata = (feature.metadata || {}) as Record<string, unknown>;
    const title = feature.title || '';
    const weightage = metadata.weightage || totalFeatures;
    const info =
      typeof metadata.info === 'string' && metadata.info.trim()
        ? metadata.info
        : undefined;

    const existing = rows.get(title);
    rows.set(title, { weightage, info: existing?.info ?? info });
  });

  return Array.from(rows.entries())
    .sort((a, b) => Number(a[1].weightage) - Number(b[1].weightage))
    .filter(([title]) => Boolean(title))
    .map(([title, { info }]) => (info ? { title, info } : { title }));
}

const LINK_PATTERN = /\[([^\]]+)\]\(([^)\s]+)\)/g;
const ALLOWED_URL_PATTERN = /^(https?|mailto):/i;

/*
 * Splits info text into text and link parts.
 * - Only [label](url) is parsed; the url must start with http:, https: or mailto:.
 * - Everything else is kept as plain text exactly as written.
 */
export function parseFeatureInfo(info: string): FeatureInfoPart[] {
  const parts: FeatureInfoPart[] = [];
  const pushText = (text: string) => {
    if (!text) return;
    const last = parts[parts.length - 1];
    if (last?.type === 'text') last.text += text;
    else parts.push({ type: 'text', text });
  };

  let cursor = 0;
  for (const match of Array.from(info.matchAll(LINK_PATTERN))) {
    const [raw, label, url] = match;
    const start = match.index ?? 0;
    pushText(info.slice(cursor, start));
    if (ALLOWED_URL_PATTERN.test(url)) parts.push({ type: 'link', label, url });
    else pushText(raw);
    cursor = start + raw.length;
  }
  pushText(info.slice(cursor));

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
