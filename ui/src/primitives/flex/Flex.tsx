import { styled } from "~/stitches";

export const Flex = styled("div", {
  boxSizing: "border-box",
  display: "flex",

  variants: {
    direction: {
      row: {
        flexDirection: "row",
      },
      column: {
        flexDirection: "column",
      },
      rowReverse: {
        flexDirection: "row-reverse",
      },
      columnReverse: {
        flexDirection: "column-reverse",
      },
    },
    align: {
      start: {
        alignItems: "flex-start",
      },
      center: {
        alignItems: "center",
      },
      end: {
        alignItems: "flex-end",
      },
      stretch: {
        alignItems: "stretch",
      },
      baseline: {
        alignItems: "baseline",
      },
    },
    justify: {
      start: {
        justifyContent: "flex-start",
      },
      center: {
        justifyContent: "center",
      },
      end: {
        justifyContent: "flex-end",
      },
      between: {
        justifyContent: "space-between",
      },
    },
    wrap: {
      noWrap: {
        flexWrap: "nowrap",
      },
      wrap: {
        flexWrap: "wrap",
      },
      wrapReverse: {
        flexWrap: "wrap-reverse",
      },
    },
    gap: {
      4: {
        gap: "$4",
      },
      6: {
        gap: "$6",
      },
      8: {
        gap: "$8",
      },
      16: {
        gap: "$16",
      },
      20: {
        gap: "$20",
      },
      32: {
        gap: "$32",
      },
    },
  },
  defaultVariants: {
    direction: "column",
    align: "stretch",
    justify: "start",
    wrap: "noWrap",
  },
});
