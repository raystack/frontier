import { Button, List } from "@raystack/apsara";
import { ReactNode } from "react";
import styles from "./audit-logs.module.css";

type SidepanelListItemLinkProps = {
  isLink: boolean;
  children: ReactNode;
  href: string;
  label: string;
  onNavigate?: (path: string) => void;
  "data-test-id"?: string;
};

export default function SidepanelListItemLink({
  isLink,
  children,
  href,
  label,
  onNavigate,
  "data-test-id": dataTestId,
}: SidepanelListItemLinkProps) {
  if (isLink && onNavigate) {
    return (
      <List.Item className={styles["sidepanel-list-link"]}>
        <List.Label minWidth="112px">{label}</List.Label>
        <List.Value className={styles["text-overflow"]}>
          <Button
            variant="text"
            color="neutral"
            data-test-id={dataTestId}
            className={styles["sidepanel-link-trigger"]}
            onClick={() => onNavigate(href)}>
            {children}
          </Button>
        </List.Value>
      </List.Item>
    );
  }
  return (
    <List.Item>
      <List.Label minWidth="120px">{label}</List.Label>
      <List.Value>{children}</List.Value>
    </List.Item>
  );
}
