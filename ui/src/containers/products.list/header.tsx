import { PlusIcon } from "@radix-ui/react-icons";
import { Button, Flex, DataTable } from "@raystack/apsara/v1";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

const defaultPageHeader = {
  title: "Products",
  breadcrumb: [] as {
    href: string;
    name: string;
  }[],
};

export const ProductsHeader = ({ header = defaultPageHeader }) => {
  const navigate = useNavigate();

  return (
    <PageHeader title={header.title} breadcrumb={header.breadcrumb}>
      <DataTable.Search placeholder="Search products..." size="small" />
      <Button
        size={"small"}
        variant="outline"
        color="neutral"
        onClick={() => navigate("/products/create")}
        style={{ width: "100%" }}
        data-test-id="admin-ui-create-product-btn"
      >
        <Flex
          direction="column"
          align="center"
          style={{ paddingRight: "var(--pd-4)" }}
        >
          <PlusIcon />
        </Flex>
        new product
      </Button>
    </PageHeader>
  );
};
