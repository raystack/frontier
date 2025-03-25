import { useOutletContext } from "react-router-dom";
import { OutletContext } from "../types";
import { Flex } from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import styles from "./members.module.css";

export function OrganizationMembersPage() {
  const { organization } = useOutletContext<OutletContext>();

  const title = `Members | ${organization.title} | Organizations`;

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <Flex className={styles["content"]} direction="column" gap={9}></Flex>
    </Flex>
  );
}
