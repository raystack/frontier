import type { CSSProperties, PropsWithChildren } from "react";
import { Flex, Text } from "@raystack/apsara";
import styles from "./page-header.module.css";

export type PageHeaderTypes = {
  title: string;
  breadcrumb: { name: string; href?: string }[];
  className?: string;
  style?: CSSProperties;
};

export function PageHeader({
  title,
  breadcrumb,
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
      style={{ padding: "16px 24px", ...style }}
      {...props}
    >
      <Flex align="center" gap="medium">
        <Flex align="center" gap="small" className={styles.breadcrumb}>
          <Text style={{ fontSize: "14px", fontWeight: "500" }}>{title}</Text>
          {breadcrumb.map((item) => (
            <span key={item.name} className={styles.breadcrumbItem}>
              {item.name}
            </span>
          ))}
        </Flex>
      </Flex>
      <Flex gap="small">{children}</Flex>
    </Flex>
  );
}
