import { EmptyState, Flex } from "@raystack/apsara/v1";
import styles from "./tokens.module.css";
import { CoinIcon } from "@raystack/apsara/icons";
import { useContext } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";

const NoTokens = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No tokens present"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<CoinIcon />}
    />
  );
};

export function OrganizationTokensPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const title = `Tokens | ${organization?.title} | Organizations`;

  return (
    <Flex justify="center">
      <PageTitle title={title} />
      <NoTokens />
    </Flex>
  );
}
