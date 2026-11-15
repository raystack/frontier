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
});

describe('parseFeatureInfo', () => {
  it('returns plain text as one part', () => {
    expect(parseFeatureInfo('Just text.')).toEqual([
      { type: 'text', text: 'Just text.' }
    ]);
  });

  it('parses one link', () => {
    expect(parseFeatureInfo('See [Details](https://example.com).')).toEqual([
      { type: 'text', text: 'See ' },
      { type: 'link', label: 'Details', url: 'https://example.com' },
      { type: 'text', text: '.' }
    ]);
  });

  it('parses two links', () => {
    expect(
      parseFeatureInfo(
        '[Docs](http://example.com/docs) or [support@example.com](mailto:support@example.com)'
      )
    ).toEqual([
      { type: 'link', label: 'Docs', url: 'http://example.com/docs' },
      { type: 'text', text: ' or ' },
      {
        type: 'link',
        label: 'support@example.com',
        url: 'mailto:support@example.com'
      }
    ]);
  });

  it('keeps links with a disallowed scheme as plain text', () => {
    const input = 'Click [here](javascript:alert(1)) now';
    expect(parseFeatureInfo(input)).toEqual([{ type: 'text', text: input }]);
  });

  it('keeps malformed markdown as plain text', () => {
    const inputs = [
      '[Details](https://example.com',
      '[Details] (https://example.com)',
      '[](https://example.com)',
      'Details](https://example.com)'
    ];
    inputs.forEach(input => {
      expect(parseFeatureInfo(input)).toEqual([{ type: 'text', text: input }]);
    });
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
