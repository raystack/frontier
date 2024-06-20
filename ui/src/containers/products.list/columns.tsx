import { Pencil2Icon } from "@radix-ui/react-icons";
import { ApsaraColumnDef, Flex, Image } from "@raystack/apsara";
import { V1Beta1Product } from "@raystack/frontier";
import { Link, NavLink } from "react-router-dom";
import { Price } from "~/components/Price";

export const getColumns: () => ApsaraColumnDef<V1Beta1Product>[] = () => {
  return [
    {
      header: " ",
      cell: ({ row, getValue }) => {
        return (
          <Link to={`/products/${row.getValue("id")}`}>
            <Image
              src="/product.svg"
              alt="product-icon"
              width={20}
              style={{
                backgroundColor: "var(--background-inset)",
                padding: "var(--pd-6)",
                borderRadius: "var(--pd-6)",
                border: "1px solid var(--border-base)",
              }}
            />
          </Link>
        );
      },
    },
    {
      header: "Name",
      accessorKey: "title",
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        const prices = row?.original?.prices;

        const priceComp =
          prices?.length == 1 ? (
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
      footer: (props) => props.column.id,
    },

    {
      header: "Created on",
      accessorKey: "created_at",
      meta: {
        headerFilter: false,
      },
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
      filterVariant: "date",
      footer: (props) => props.column.id,
    },
    {
      header: "Updated on",
      accessorKey: "updated_at",
      meta: {
        headerFilter: false,
      },
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
      filterVariant: "date",
      footer: (props) => props.column.id,
    },
    {
      header: "Actions",
      cell: ({ row, getValue }) => {
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
