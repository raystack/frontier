import { Flex } from "~/primitives/flex/Flex";

const BRAND_NAME = "Shield UI";
export default function SidebarHeader() {
  return (
    <Flex
      align="center"
      justify="between"
      direction="row"
      css={sidebarHeaderContainerStyle}
    >
      {BRAND_NAME}
    </Flex>
  );
}

const sidebarHeaderContainerStyle = {
  height: "3.2rem",
  marginBottom: "2.4rem",
  fontSize: "1.6rem",
  fontWeight: "bold",
};
