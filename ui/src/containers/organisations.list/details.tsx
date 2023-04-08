import { Flex, styled } from "@odpf/apsara";
import { useOrganisation } from ".";

export default function OrganisationDetails() {
  const { organizationMapByName } = useOrganisation();
  return <></>;
}

const TeamContainer = styled(Flex, {
  width: "3rem",
});
