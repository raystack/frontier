import { Flex, IconButton, Text } from "@raystack/apsara/v1";
import { V1Beta1Domain } from "~/api/frontier";
import styles from "./security.module.css";
import { CheckCircledIcon, TrashIcon } from "@radix-ui/react-icons";
import Skeleton from "react-loading-skeleton";

interface DomainItemProps {
  domain: V1Beta1Domain;
}

const DomainItem = ({ domain }: DomainItemProps) => {
  return (
    <Flex className={styles["domains-list-item"]} justify="between">
      <Flex gap={3}>
        <Text size={3}>{domain?.name}</Text>
        {domain.state === "verified" ? (
          <CheckCircledIcon
            color={"var(--rs-color-foreground-success-primary)"}
          />
        ) : null}
      </Flex>
      <IconButton size={3} data-test-id="delete-domain-button">
        <TrashIcon />
      </IconButton>
    </Flex>
  );
};

interface DomainsListProps {
  isLoading: boolean;
  domains: V1Beta1Domain[];
  fetchDomains: () => void;
}

export const DomainsList = ({ isLoading, domains = [] }: DomainsListProps) => {
  return isLoading ? (
    <Flex direction="column" className={styles["domains-list"]}>
      {[...Array(3)].map((_, index) => (
        <Skeleton
          key={index}
          height={20}
          style={{ margin: "var(--rs-space-5) 0" }}
        />
      ))}
    </Flex>
  ) : domains.length ? (
    <Flex direction="column" className={styles["domains-list"]}>
      {domains.map((domain) => (
        <DomainItem key={domain?.id} domain={domain} />
      ))}
    </Flex>
  ) : null;
};
