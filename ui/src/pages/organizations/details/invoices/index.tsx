import { EmptyState, Flex } from "@raystack/apsara/v1";
import styles from "./invoices.module.css";
import { FileTextIcon } from "@radix-ui/react-icons";
import { useContext } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";

const NoInvoices = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Invoice found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<FileTextIcon />}
    />
  );
};

export function OrganizationInvoicesPage() {
  const { organization } = useContext(OrganizationContext);

  const title = `Invoices | ${organization?.title} | Organizations`;

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
    </Flex>
  );
}
