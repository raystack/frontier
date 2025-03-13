import { OrganizationsDetailsNavabar } from "./navbar";
import styles from "./details.module.css";
import { Flex } from "@raystack/apsara/v1";

export const OrganizationDetails = () => {
  return (
    <Flex direction="column" className={styles.page}>
      <OrganizationsDetailsNavabar />
      <p>This is the details page for an organization.</p>
    </Flex>
  );
};
