import * as z from "zod";

export const intervals = [
  { value: "day", title: "Day" },
  { value: "week", title: "Week" },
  { value: "month", title: "Month" },
  { value: "year", title: "year" },
];

export const behaviors = [
  { value: "basic", title: "Basic" },
  { value: "per_seat", title: "Per seat" },
  { value: "credits", title: "Credits" },
];

export const defaultFormValues = {
  title: "",
  name: "",
  description: "",
  behavior: "basic",
  prices: [
    {
      name: "",
      interval: "",
    },
  ],
  behavior_config: {
    credit_amount: "",
    seat_limit: "",
  },
  metadata: [
    {
      key: "",
      value: "",
    },
  ],
};

export type ProductForm = z.infer<typeof productSchema>;
export const productSchema = z.object({
  title: z
    .string()
    .trim()
    .min(3, { message: "Must be 3 or more characters long" }),
  name: z
    .string()
    .trim()
    .toLowerCase()
    .min(3, { message: "Must be 3 or more characters long" }),
  description: z.string().optional().default(""),
  behavior: z.string().optional().default("basic"),

  prices: z
    .object({
      name: z.string(),
      interval: z.string(),
      amount: z
        .string()
        .refine((num) => parseInt(num))
        .transform(Number),
    })
    .array()
    .default([]),
  metadata: z
    .object({ key: z.string(), value: z.string() })
    .array()
    .default([]),
  behavior_config: z.any().default({}),
  features: z.object({ name: z.string() }).array().default([]),
  newfeatures: z.string().default(""),
});
