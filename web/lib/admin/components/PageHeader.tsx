import React from "react";
import { Flex, Text } from "@raystack/apsara";
import styles from "./page-header.module.css";

export type PageHeaderTypes = {
  title: string;
  breadcrumb: { name: string; href?: string }[];
  // eslint-disable-next-line no-unused-vars -- callback param name is for type documentation
  onBreadcrumbClick?: (item: { name: string; href?: string }) => void;
  className?: string;
  style?: React.CSSProperties;
};

export function PageHeader({
  title,
  breadcrumb,
  onBreadcrumbClick,
  children,
  className,
  style = {},
  ...props
}: React.PropsWithChildren<PageHeaderTypes>) {
  return (
    <Flex
      align="center"
      justify="between"
      className={className}
      style={{ padding: "16px 24px", ...style }}
      {...props}
    >
      <Flex align="center" gap="medium">
        <Flex align="center" gap="small" className={styles.breadcrumb}>
          <Text style={{ fontSize: "14px", fontWeight: "500" }}>{title}</Text>
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
      <Flex gap="small">{children}</Flex>
    </Flex>
  );
}
