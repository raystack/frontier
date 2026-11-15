import {
  buildFeatureRows,
  normalizeFeatureValue,
  parseFeatureInfo
} from '~/client/views/plans/helpers/feature-rows';

describe('buildFeatureRows', () => {
  it('sorts by weightage', () => {
    const rows = buildFeatureRows([
      { title: 'C', metadata: { weightage: 3 } },
      { title: 'A', metadata: { weightage: 1 } },
      { title: 'B', metadata: { weightage: 2 } }
    ]);
    expect(rows.map(r => r.title)).toEqual(['A', 'B', 'C']);
  });

  it('falls back to the feature count when weightage is missing', () => {
    const rows = buildFeatureRows([
      { title: 'NoWeight', metadata: {} },
      { title: 'NoMeta' },
      { title: 'Last', metadata: { weightage: 5 } },
      { title: 'First', metadata: { weightage: 1 } }
    ]);
    expect(rows.map(r => r.title)).toEqual([
      'First',
      'NoWeight',
      'NoMeta',
      'Last'
    ]);
  });

  it('drops empty titles', () => {
    const rows = buildFeatureRows([
      { title: '', metadata: { weightage: 1 } },
      { metadata: { weightage: 1 } },
      { title: 'A', metadata: { weightage: 2 } }
    ]);
    expect(rows).toEqual([{ title: 'A' }]);
  });

  it('collapses duplicate titles and keeps the first non-empty info', () => {
    const rows = buildFeatureRows([
      { title: 'A', metadata: { weightage: 1, info: '  ' } },
      { title: 'B', metadata: { weightage: 2 } },
      { title: 'A', metadata: { weightage: 3, info: 'first' } },
      { title: 'A', metadata: { weightage: 3, info: 'second' } }
    ]);
    expect(rows).toEqual([{ title: 'B' }, { title: 'A', info: 'first' }]);
  });

  it('ignores missing, blank and non-string info', () => {
    const rows = buildFeatureRows([
      { title: 'A', metadata: { weightage: 1 } },
      { title: 'B', metadata: { weightage: 2, info: '' } },
      { title: 'C', metadata: { weightage: 3, info: '   ' } },
      { title: 'D', metadata: { weightage: 4, info: 42 } },
      { title: 'E', metadata: { weightage: 5, info: { text: 'x' } } },
      { title: 'F', metadata: { weightage: 6, info: 'Help' } }
    ]);
    expect(rows).toEqual([
      { title: 'A' },
      { title: 'B' },
      { title: 'C' },
      { title: 'D' },
      { title: 'E' },
      { title: 'F', info: 'Help' }
    ]);
  });

  it('trims surrounding whitespace from info', () => {
    const rows = buildFeatureRows([
      { title: 'A', metadata: { weightage: 1, info: '\n  Help text\n' } }
    ]);
    expect(rows[0].info).toBe('Help text');
  });

  it('keeps the old ordering for tied integer-like titles', () => {
    const rows = buildFeatureRows([
      { title: 'Storage', metadata: {} },
      { title: '2', metadata: {} },
      { title: '1', metadata: {} }
    ]);
    expect(rows.map(r => r.title)).toEqual(['1', '2', 'Storage']);
  });
});

/*
 * - Renders parts the way the popover shows them: links as <url|label>.
 * - Asserts on visible output, not on how text is split into parts.
 */
const show = (info: string) =>
  parseFeatureInfo(info)
    .map(p => (p.type === 'text' ? p.text : `<${p.url}|${p.label}>`))
    .join('');

describe('parseFeatureInfo', () => {
  it('shows plain text unchanged', () => {
    expect(show('Just text.')).toBe('Just text.');
  });

  it('turns one link into a link', () => {
    expect(show('See [Details](https://example.com).')).toBe(
      'See <https://example.com|Details>.'
    );
  });

  it('turns two links into links', () => {
    expect(
      show(
        '[Docs](http://example.com/docs) or [support@example.com](mailto:support@example.com)'
      )
    ).toBe(
      '<http://example.com/docs|Docs> or <mailto:support@example.com|support@example.com>'
    );
  });

  it('keeps links with a disallowed scheme as plain text', () => {
    [
      'Click [here](javascript:alert(1)) now',
      'Go [there](/settings/billing)',
      'Get [file](ftp://example.com/file)',
      'Shout [docs](HTTPS://example.com)'
    ].forEach(input => {
      expect(show(input)).toBe(input);
    });
  });

  it('keeps malformed markdown as plain text', () => {
    [
      '[Details](https://example.com',
      '[Details] (https://example.com)',
      '[](https://example.com)',
      'Details](https://example.com)'
    ].forEach(input => {
      expect(show(input)).toBe(input);
    });
  });

  it('keeps a blank label as plain text', () => {
    const input = 'Hidden [ ](https://example.com) link';
    expect(show(input)).toBe(input);
  });
});

describe('normalizeFeatureValue', () => {
  it('returns strings unchanged', () => {
    expect(normalizeFeatureValue('10 GB')).toBe('10 GB');
  });

  it('converts true and non-zero numbers to strings', () => {
    expect(normalizeFeatureValue(true)).toBe('true');
    expect(normalizeFeatureValue(5)).toBe('5');
  });

  it('returns "-" for falsy values', () => {
    [undefined, null, '', false, 0].forEach(value => {
      expect(normalizeFeatureValue(value)).toBe('-');
    });
  });
});
