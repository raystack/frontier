import { Dialog, IconButton, CodeBlock } from "@raystack/apsara";
import styles from "./list.module.css";
import { AuditRecord } from "@raystack/proton/frontier";
import { auditLogToJson } from "../util";
import JsonIcon from "~/assets/icons/json.svg?react";

export default function SidePanelLogDialog(props: Partial<AuditRecord>) {
  return (
    <Dialog key="show-audit-json-dialog">
      <Dialog.Trigger asChild>
        <IconButton
          size={3}
          key="show-audit-json-icon"
          data-test-id="show-audit-json-icon">
          <JsonIcon />
        </IconButton>
      </Dialog.Trigger>
      <Dialog.Content
        width={600}
        ariaLabel="Basic Dialog"
        ariaDescription="A simple dialog example">
        <Dialog.Header>
          <Dialog.Title>Log entry</Dialog.Title>
          <Dialog.CloseButton data-test-id="close-audit-json-dialog-icon" />
        </Dialog.Header>
        <Dialog.Body className={styles["code-block-container"]}>
          <CodeBlock>
            <CodeBlock.Content>
              <CodeBlock.Code language="jsx" className={styles["code-block"]}>
                {auditLogToJson(props as AuditRecord)}
              </CodeBlock.Code>
              <CodeBlock.CopyButton
                variant="floating"
                data-test-id="copy-audit-json-button"
              />
            </CodeBlock.Content>
          </CodeBlock>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
}
