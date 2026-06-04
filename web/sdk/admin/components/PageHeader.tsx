import type { CSSProperties, PropsWithChildren, ReactNode } from "react";
import { Flex, Text } from "@raystack/apsara";
import styles from "./page-header.module.css";

export type PageHeaderTypes = {
  title: string;
  icon?: ReactNode;
  breadcrumb: { name: string; href?: string }[];
  // eslint-disable-next-line no-unused-vars -- callback param name is for type documentation
  onBreadcrumbClick?: (item: { name: string; href?: string }) => void;
  className?: string;
  style?: CSSProperties;
};

export function PageHeader({
  title,
  icon,
  breadcrumb,
  onBreadcrumbClick,
  children,
  className,
  style = {},
  ...props
}: PropsWithChildren<PageHeaderTypes>) {
  return (
    <Flex
      align="center"
      justify="between"
      className={className}
      style={{ padding: "var(--rs-space-5) var(--rs-space-7)", minHeight: "var(--rs-space-12)", ...style }}
      {...props}
    >
      <Flex align="center" gap={5}>
        <Flex align="center" gap={2} className={styles.breadcrumb}>
          {icon}
          <Text size="regular" weight="medium">{title}</Text>
          {breadcrumb.map((item) =>
            item.href && onBreadcrumbClick ? (
              <span
                key={item.name}
                role="button"
                tabIndex={0}
                className={styles.breadcrumbItem}
                data-test-id="page-header-breadcrumb-link"
                onClick={() => onBreadcrumbClick?.(item)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    onBreadcrumbClick?.(item);
                  }
                }}
                style={{ cursor: "pointer" }}
              >
                {item.name}
              </span>
            ) : (
              <span key={item.name} className={styles.breadcrumbItem}>
                {item.name}
              </span>
            )
          )}
        </Flex>
      </Flex>
      <Flex gap={3}>{children}</Flex>
    </Flex>
  );
}
