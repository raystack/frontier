import { Flex } from "~/primitives/flex/Flex";
import SidebarContent from "./SidebarContent";
import SidebarFooter from "./SidebarFooter";
import SidebarHeader from "./SidebarHeader";

export default function Sidebar({ children }: { children: React.ReactNode }) {
  return (
    <Flex justify="between" as="nav" css={sidebarContainerStyle}>
      {children}
    </Flex>
  );
}

Sidebar.Header = SidebarHeader;
Sidebar.Content = SidebarContent;
Sidebar.Footer = SidebarFooter;

const sidebarContainerStyle = {
  width: "100%",
  padding: "1.6rem",
  "@media (max-width: 1200px)": {
    display: "none",
  },
};
