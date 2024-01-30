import { Flex, Link, Text } from "@raystack/apsara";
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
      <Flex align="center" gap="medium">
        <Flex align="center" gap="small" className={styles.breadcrumb}>
          <Text style={{ fontSize: "14px", fontWeight: "500" }}>{title}</Text>
          {breadcrumb.map((item) => (
            <Link
              key={item.name}
              href={item?.href}
              style={{ display: "flex", flexDirection: "row", gap: "8px" }}
            >
              <Text size={1}>{item.name}</Text>
            </Link>
          ))}
        </Flex>
      </Flex>
      <Flex gap="small">{children}</Flex>
    </Flex>
  );
}
