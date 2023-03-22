import { styled } from "~/stitches";

export const Text = styled("span", {
  // Reset
  lineHeight: "1",
  margin: "0",
  fontWeight: 400,
  fontVariantNumeric: "tabular-nums",
  display: "block",

  variants: {
    type: {
      small: {
        fontSize: "$12",
        lineHeight: "$16",
        letterSpacing: "$40",
      },
      medium: {
        fontSize: "$14",
        lineHeight: "$20",
        letterSpacing: "$25",
      },
      large: {
        fontSize: "$16",
        lineHeight: "$24",
        letterSpacing: "$50",
      },
    },
    size: {
      "11": {
        fontSize: "$11",
        lineHeight: "$16",
        letterSpacing: "$50",
      },
      "12": {
        fontSize: "$12",
        lineHeight: "$16",
        letterSpacing: "$40",
      },
      "13": {
        fontSize: "$13",
        lineHeight: "$18",
        letterSpacing: "$40",
      },
      "14": {
        fontSize: "$14",
        lineHeight: "$20",
        letterSpacing: "$25",
      },
      "16": {
        fontSize: "$16",
        lineHeight: "$24",
        letterSpacing: "$10",
      },
      "18": {
        fontSize: "$18",
        lineHeight: "$24",
        letterSpacing: "-$1.5",
      },
      "20": {
        fontSize: "$20",
        lineHeight: "$24",
      },
      "22": {
        fontSize: "$24",
        lineHeight: "$28",
      },
      "24": {
        fontSize: "$24",
        lineHeight: "$32",
      },
      "28": {
        fontSize: "$28",
        lineHeight: "$36",
      },
      "32": {
        fontSize: "$32",
        lineHeight: "$40",
      },
      "36": {
        fontSize: "$36",
        lineHeight: "$44",
      },
      "40": {
        fontSize: "$40",
        lineHeight: "$48",
      },
      "45": {
        fontSize: "$45",
        lineHeight: "$52",
      },
      "57": {
        fontSize: "$57",
        lineHeight: "$64",
      },
    },

    weight: {
      normal: {
        fontWeight: "$default",
      },
      500: {
        fontWeight: "$500",
      },
      600: {
        fontWeight: "$600",
      },
      700: {
        fontWeight: "$700",
      },
    },

    variant: {
      red: {
        color: "$red11",
      },
      crimson: {
        color: "$crimson11",
      },
      pink: {
        color: "$pink11",
      },
      purple: {
        color: "$purple11",
      },
      violet: {
        color: "$violet11",
      },
      indigo: {
        color: "$indigo11",
      },
      blue: {
        color: "$blue11",
      },
      cyan: {
        color: "$cyan11",
      },
      teal: {
        color: "$teal11",
      },
      green: {
        color: "$green11",
      },
      lime: {
        color: "$lime11",
      },
      yellow: {
        color: "$yellow11",
      },
      orange: {
        color: "$orange11",
      },
      gold: {
        color: "$gold11",
      },
      bronze: {
        color: "$bronze11",
      },
      gray: {
        color: "$gray11",
      },
      contrast: {
        color: "$hiContrast",
      },
    },
    gradient: {
      true: {
        WebkitBackgroundClip: "text",
        WebkitTextFillColor: "transparent",
      },
    },
  },
  compoundVariants: [
    {
      variant: "red",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $red11, $crimson11)",
      },
    },
    {
      variant: "crimson",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $crimson11, $pink11)",
      },
    },
    {
      variant: "pink",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $pink11, $purple11)",
      },
    },
    {
      variant: "purple",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $purple11, $violet11)",
      },
    },
    {
      variant: "violet",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $violet11, $indigo11)",
      },
    },
    {
      variant: "indigo",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $indigo11, $blue11)",
      },
    },
    {
      variant: "blue",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $blue11, $cyan11)",
      },
    },
    {
      variant: "cyan",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $cyan11, $teal11)",
      },
    },
    {
      variant: "teal",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $teal11, $green11)",
      },
    },
    {
      variant: "green",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $green11, $lime11)",
      },
    },
    {
      variant: "lime",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $lime11, $yellow11)",
      },
    },
    {
      variant: "yellow",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $yellow11, $orange11)",
      },
    },
    {
      variant: "orange",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $orange11, $red11)",
      },
    },
    {
      variant: "gold",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $gold11, $gold9)",
      },
    },
    {
      variant: "bronze",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $bronze11, $bronze9)",
      },
    },
    {
      variant: "gray",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $gray11, $gray12)",
      },
    },
    {
      variant: "contrast",
      gradient: "true",
      css: {
        background: "linear-gradient(to right, $hiContrast, $gray12)",
      },
    },
    {
      variant: "base",
      css: {},
    },
  ],
  defaultVariants: {
    size: "small",
    weight: "normal",
    variant: "base",
  },
});
