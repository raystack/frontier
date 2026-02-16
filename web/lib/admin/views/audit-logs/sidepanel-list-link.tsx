import { Button, List } from "@raystack/apsara";
import { ReactNode } from "react";
import { Link } from "react-router-dom";
import styles from "./audit-logs.module.css";

type SidepanelListItemLinkProps = {
  isLink: boolean;
  children: ReactNode;
  href: string;
  label: string;
};

export default function SidepanelListItemLink({
  isLink,
  children,
  href,
  label,
}: SidepanelListItemLinkProps) {
  if (isLink) {
    return (
      <List.Item className={styles["sidepanel-list-link"]}>
        <List.Label minWidth="112px">{label}</List.Label>
        <List.Value className={styles["text-overflow"]}>
          <Link to={href}>
            <Button
              variant="text"
              color="neutral"
              data-test-id="organization-link"
              className={styles["sidepanel-link-trigger"]}>
              {children}
            </Button>
          </Link>
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
