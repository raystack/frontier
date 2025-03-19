import { Button, Flex, IconButton, Separator, Text } from "@raystack/apsara/v1";

import styles from "./security.module.css";
import { CheckCircledIcon, PlusIcon, TrashIcon } from "@radix-ui/react-icons";

const DomainItem = () => {
  return (
    <Flex className={styles["domains-list-item"]} justify="between">
      <Flex gap={3}>
        <Text size={3}>raystack.org</Text>
        <CheckCircledIcon
          color={"var(--rs-color-foreground-success-primary)"}
        />
      </Flex>
      <IconButton size={3} data-test-id="delete-domain-button">
        <TrashIcon />
      </IconButton>
    </Flex>
  );
};

export const OrganizationSecurity = () => {
  return (
    <Flex justify="center" className={styles["container"]}>
      <Flex className={styles["content"]} direction="column" gap={9}>
        <Flex gap={5} justify="between">
          <Flex direction="column" gap={3}>
            <Text size={5}>Allowed email domains</Text>
            <Text size={3} variant="secondary">
              Anyone with an email address at these domains is allowed to sign
              up for this workspace.
            </Text>
          </Flex>
          <Button
            variant="outline"
            color="neutral"
            leadingIcon={<PlusIcon />}
            size="small"
            data-test-id="add-domain-button"
          >
            <Text size={1}>Add domain</Text>
          </Button>
        </Flex>
        <Flex direction="column" className={styles["domains-list"]}>
          <DomainItem />
        </Flex>
      </Flex>
    </Flex>
  );
};
