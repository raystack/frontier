import { Pencil2Icon } from "@radix-ui/react-icons";
import { Flex, Image, type DataTableColumnDef } from "@raystack/apsara";
import type { V1Beta1Product } from "@raystack/frontier";
import { Link, NavLink } from "react-router-dom";
import { Price } from "~/components/Price";

export const getColumns: () => DataTableColumnDef<
  V1Beta1Product,
  unknown
>[] = () => {
  return [
    {
      accessorKey: "id",
      header: "",
      cell: (info) => {
        return (
          <Link to={`/products/${info.getValue() as string}`}>
            <Image
              src="/product.svg"
              alt="product-icon"
              width={20}
              style={{
                backgroundColor: "var(--rs-color-background-neutral-secondary)",
                padding: "var(--rs-space-3)",
                borderRadius: "var(--rs-radius-3)",
                border: "1px solid var(--rs-color-border-base-primary)",
              }}
            />
          </Link>
        );
      },
    },
    {
      header: "Name",
      accessorKey: "title",
      cell: ({ row }) => {
        const prices = row?.original?.prices;

        const priceComp =
          prices?.length === 1 ? (
            <Price value={prices[0].amount} currency={prices[0].currency} />
          ) : (
            <NavLink to={`/products/${row?.original?.id}/prices`}>
              {prices?.length} prices
            </NavLink>
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
      accessorKey: "created_at",
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
      filterVariant: "date",
    },
    {
      header: "Updated on",
      accessorKey: "updated_at",
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
      filterVariant: "date",
    },
    {
      accessorKey: "id",
      header: "Actions",
      cell: ({ row }) => {
        return (
          <Flex align="center" justify="center" gap="small">
            <NavLink to={`/products/${row?.original?.id}/edit`}>
              <Pencil2Icon />
            </NavLink>
          </Flex>
        );
      },
    },
  ];
};
