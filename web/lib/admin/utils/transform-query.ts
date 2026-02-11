import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import type { RQLRequest, RQLFilter, RQLSort } from "@raystack/proton/frontier";
import { RQLRequestSchema, RQLFilterSchema, RQLSortSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

// Extract DataTableFilter type from DataTableQuery since it's not exported
type DataTableFilter = NonNullable<DataTableQuery["filters"]>[number];

export interface TransformOptions {
  /**
   * Default limit if not specified
   * @default 50
   */
  defaultLimit?: number;
  // TODO: add support for more cases in RQL
  fieldNameMapping?: Record<string, string>;
}

/**
 * Converts a filter value to the appropriate RQLFilter value format
 */
function convertFilterValue(
  value: string | number | boolean | null | undefined,
): RQLFilter["value"] {
  switch (typeof value) {
    case "boolean":
      return { case: "boolValue", value };
    case "number":
      return { case: "numberValue", value };
    case "string":
      return { case: "stringValue", value };
    default:
      return { case: "stringValue", value: value ?? "" };
  }
}

/**
 * Transforms a DataTableFilter to RQLFilter
 */
function transformFilter(
  filter: DataTableFilter,
  fieldNameMapping?: Record<string, string>,
): RQLFilter {
  // Priority: typed values > generic value field
  let value: RQLFilter["value"];

  if (filter.boolValue !== undefined) {
    value = { case: "boolValue", value: filter.boolValue };
  } else if (filter.numberValue !== undefined) {
    value = { case: "numberValue", value: filter.numberValue };
  } else if (filter.stringValue !== undefined) {
    value = { case: "stringValue", value: filter.stringValue };
  } else {
    value = convertFilterValue(filter.value);
  }

  // TODO: add support for ilike in RQL and backend
  // Transform ilike operator to like only for string values
  const operator =
    filter.operator === "ilike" && value.case === "stringValue"
      ? "like"
      : filter.operator;

  const fieldName = fieldNameMapping?.[filter.name] ?? filter.name;

  return create(RQLFilterSchema, {
    name: fieldName,
    operator,
    value,
  });
}

/**
 * Transforms DataTableSort to RQLSort (if RQLSort exists in the API)
 */
function transformSort(
  sort: DataTableSort[],
  fieldNameMapping?: Record<string, string>,
): RQLSort[] | undefined {
  if (!sort || sort.length === 0) {
    return undefined;
  }

  return sort.map((s) => create(RQLSortSchema, {
    ...s,
    name: fieldNameMapping?.[s.name] ?? s.name,
  }));
}

export function transformDataTableQueryToRQLRequest(
  query: DataTableQuery,
  options: TransformOptions = {},
): RQLRequest {
  const { defaultLimit = 50, fieldNameMapping } = options;

  // Transform DataTable filters
  const filters: RQLFilter[] = query.filters?.length
    ? query.filters.map((filter) => transformFilter(filter, fieldNameMapping))
    : [];

  // Build the RQLRequest with snake_case properties
  const rqlRequest = create(RQLRequestSchema, {
    filters,
    groupBy: query.group_by || [],
    offset: query.offset || 0,
    limit: query.limit || defaultLimit,
    sort: transformSort(query.sort || [], fieldNameMapping) || [],
    search: query.search || "",
  });

  return rqlRequest;
}
