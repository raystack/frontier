import { styled } from "~/stitches";

export const BaseContainer = styled("div", {
  // Reset
  boxSizing: "border-box",
  flexShrink: 0,
  variants: {
    size: {
      "1": {
        maxWidth: "430px",
      },
      "2": {
        maxWidth: "715px",
      },
      "3": {
        maxWidth: "1145px",
      },
      "4": {
        maxWidth: "none",
      },
    },
    background: {
      gradient: {
        height: "100vh",
        width: "100vw",
      },
      none: {},
    },
  },
  defaultVariants: {
    background: "gradient",
  },
});
