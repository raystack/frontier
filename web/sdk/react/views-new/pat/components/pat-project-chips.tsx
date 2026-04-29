'use client';

import { useLayoutEffect, useRef, useState } from 'react';
import { Chip } from '@raystack/apsara-v1';
import styles from './pat-project-chips.module.css';

interface ProjectChipItem {
  id: string;
  title: string;
}

interface PATProjectChipsProps {
  projects: ProjectChipItem[];
}

const MAX_WIDTH = 500;
const COUNT_CHIP_WIDTH_RESERVE = 56;
const CHIP_GAP = 4;

export function PATProjectChips({ projects }: PATProjectChipsProps) {
  const [expanded, setExpanded] = useState(false);
  const [visibleCount, setVisibleCount] = useState<number | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useLayoutEffect(() => {
    if (expanded || visibleCount !== null) return;
    const container = containerRef.current;
    if (!container) return;

    const children = Array.from(container.children) as HTMLElement[];
    if (children.length === 0) return;

    let used = 0;
    let count = 0;
    for (let i = 0; i < children.length; i++) {
      const w = children[i].offsetWidth;
      const remaining = children.length - i - 1;
      const reserve = remaining > 0 ? CHIP_GAP + COUNT_CHIP_WIDTH_RESERVE : 0;
      const next = used + (count > 0 ? CHIP_GAP : 0) + w;
      if (next + reserve > MAX_WIDTH) break;
      used = next;
      count++;
    }

    setVisibleCount(count > 0 ? count : 1);
  }, [expanded, visibleCount, projects]);

  if (expanded) {
    return (
      <div className={styles.expanded}>
        {projects.map(p => (
          <Chip key={p.id}>{p.title}</Chip>
        ))}
      </div>
    );
  }

  const visible =
    visibleCount === null ? projects : projects.slice(0, visibleCount);
  const hidden =
    visibleCount === null ? 0 : projects.length - visibleCount;

  return (
    <div ref={containerRef} className={styles.compact}>
      {visible.map(p => (
        <Chip key={p.id}>{p.title}</Chip>
      ))}
      {hidden > 0 && (
        <Chip
          className={styles.countChip}
          onClick={() => setExpanded(true)}
          ariaLabel={`Show ${hidden} more projects`}
          data-test-id="frontier-sdk-pat-project-chips-expand-btn"
        >
          +{hidden}
        </Chip>
      )}
    </div>
  );
}
