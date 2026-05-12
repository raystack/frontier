import React from "react";
import { Flex, Text } from "@raystack/apsara-v1";
import { Link } from "react-router-dom";
import styles from "./page-header.module.css";
export type PageHeaderTypes = {
  title: string;
  breadcrumb: { name: string; href?: string }[];
  className?: string;
  style?: React.CSSProperties;
};

export default function PageHeader({
  title,
  breadcrumb,
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
      <Flex align="center" gap={5}>
        <Flex align="center" gap={3} className={styles.breadcrumb}>
          <Text style={{ fontSize: "14px", fontWeight: "500" }}>{title}</Text>
          {breadcrumb.map((item) => (
            <Link
              key={item.name}
              to={item?.href ?? ""}
              style={{ display: "flex", flexDirection: "row", gap: "8px" }}
            >
              <Flex align="center">
                <Text size="mini">{item.name}</Text>
              </Flex>
            </Link>
          ))}
        </Flex>
      </Flex>
      <Flex gap={3}>{children}</Flex>
    </Flex>
  );
}
