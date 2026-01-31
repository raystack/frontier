import { Flex, Image, Amount, type DataTableColumnDef } from "@raystack/apsara";
import type { Product } from "@raystack/proton/frontier";
import { timestampToDate, TimeStamp } from "../../utils/connect-timestamp";

export const getColumns = (
  onNavigateToPrices?: (productId: string) => void
): DataTableColumnDef<Product, unknown>[] => {
  return [
    {
      accessorKey: "id",
      header: "",
      cell: (info) => {
        return (
              <Image
                src="/product.svg"
                alt="product-icon"
                width={24}
                height={24}
                style={{
                  backgroundColor: "var(--rs-color-background-neutral-secondary)",
                  borderRadius: "var(--rs-radius-3)",
                  border: "1px solid var(--rs-color-border-base-primary)",
                  margin: "var(--rs-space-3)",
                }}
              />
        );
      },
    },
    {
      header: "Name",
      accessorKey: "title",
      cell: ({ row }) => {
        const prices = row?.original?.prices;
        const productId = row?.original?.id;

        const priceComp =
          prices?.length === 1 ? (
            <Amount
              value={Number(prices[0].amount ?? 0)}
              currency={prices[0].currency}
            />
          ) : onNavigateToPrices && productId ? (
            <button
              data-test-id="products-table-prices-link"
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onNavigateToPrices(productId);
              }}
              style={{
                background: "none",
                border: "none",
                padding: 0,
                cursor: "pointer",
                textDecoration: "underline",
                font: "inherit",
                color: "inherit",
              }}
            >
              {prices?.length} prices
            </button>
          ) : (
            <span>{prices?.length ?? 0} prices</span>
          );

        return (
          <Flex direction="column" gap="extra-small">
            <Flex>{row?.original?.title}</Flex>
            <Flex>{priceComp}</Flex>
          </Flex>
        );
      },
    },
    {
      header: "Product Id",
      accessorKey: "id",
      filterVariant: "text",
      cell: (info) => info.getValue(),
    },
    {
      header: "Created on",
      accessorKey: "createdAt",
      cell: ({ getValue }) => {
        const timestamp = getValue() as TimeStamp | undefined;
        const date = timestampToDate(timestamp);
        if (!date) return "-";
        return date.toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        });
      },
      filterVariant: "date",
    },
    {
      header: "Updated on",
      accessorKey: "updatedAt",
      cell: ({ getValue }) => {
        const timestamp = getValue() as TimeStamp | undefined;
        const date = timestampToDate(timestamp);
        if (!date) return "-";
        return date.toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        });
      },
      filterVariant: "date",
    },
  ];
};
