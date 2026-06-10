import type { DataTableQuery, DataTableSort } from '@raystack/apsara';
import type { RQLRequest, RQLFilter, RQLSort } from '@raystack/proton/frontier';
import {
  RQLRequestSchema,
  RQLFilterSchema,
  RQLSortSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import dayjs from 'dayjs';

// Extract DataTableFilter type from DataTableQuery since it's not exported
type DataTableFilter = NonNullable<DataTableQuery['filters']>[number];

export interface TransformOptions {
  /**
   * Default limit if not specified
   * @default 50
   */
  defaultLimit?: number;
  /**
   * Maps DataTable field names to RQL field names.
   * E.g. `{ createdAt: "created_at" }` transforms the DataTable column name to the backend field.
   */
  fieldNameMapping?: Record<string, string>;
}

/**
 * Converts a filter value to the appropriate RQLFilter value format
 */
function convertFilterValue(value: unknown): RQLFilter['value'] {
  switch (typeof value) {
    case 'boolean':
      return { case: 'boolValue', value };
    case 'number':
      return { case: 'numberValue', value };
    case 'string':
      return { case: 'stringValue', value };
    default:
      return { case: 'stringValue', value: value == null ? '' : String(value) };
  }
}

/**
 * Expands a date filter into RQLFilter(s) anchored to the user's local day.
 * The date picker emits local midnight, and `eq` on a timestamp column can
 * never match a calendar day — so expand it to a [start of day, next day)
 * range. Other operators compare against the local start of day.
 */
function transformDateFilter(
  filter: DataTableFilter,
  fieldName: string
): RQLFilter[] {
  const startOfDay = dayjs(filter.value as Date).startOf('day');

  if (filter.operator === 'eq') {
    return [
      create(RQLFilterSchema, {
        name: fieldName,
        operator: 'gte',
        value: { case: 'stringValue', value: startOfDay.toISOString() }
      }),
      create(RQLFilterSchema, {
        name: fieldName,
        operator: 'lt',
        value: {
          case: 'stringValue',
          value: startOfDay.add(1, 'day').toISOString()
        }
      })
    ];
  }

  return [
    create(RQLFilterSchema, {
      name: fieldName,
      operator: filter.operator,
      value: { case: 'stringValue', value: startOfDay.toISOString() }
    })
  ];
}

function transformFilter(
  filter: DataTableFilter,
  fieldNameMapping?: Record<string, string>
): RQLFilter[] {
  const fieldName = fieldNameMapping?.[filter.name] ?? filter.name;

  if (filter.value instanceof Date) {
    return transformDateFilter(filter, fieldName);
  }

  // Priority: typed values > generic value field
  let value: RQLFilter['value'];

  if (filter.boolValue !== undefined) {
    value = { case: 'boolValue', value: filter.boolValue };
  } else if (filter.numberValue !== undefined) {
    value = { case: 'numberValue', value: Number(filter.numberValue) };
  } else if (filter.stringValue !== undefined) {
    value = { case: 'stringValue', value: filter.stringValue };
  } else {
    value = convertFilterValue(filter.value);
  }

  return [
    create(RQLFilterSchema, {
      name: fieldName,
      operator: filter.operator,
      value
    })
  ];
}

/**
 * Transforms DataTableSort to RQLSort (if RQLSort exists in the API)
 */
function transformSort(
  sort: DataTableSort[],
  fieldNameMapping?: Record<string, string>
): RQLSort[] | undefined {
  if (!sort || sort.length === 0) {
    return undefined;
  }

  return sort.map(s =>
    create(RQLSortSchema, {
      ...s,
      name: fieldNameMapping?.[s.name] ?? s.name
    })
  );
}

/**
 * Converts an Apsara `DataTableQuery` (search, filters, sort, pagination) into
 * a Frontier `RQLRequest` that can be passed to Connect RPC query hooks.
 */
export function transformDataTableQueryToRQLRequest(
  query: DataTableQuery,
  options: TransformOptions = {}
): RQLRequest {
  const { defaultLimit = 50, fieldNameMapping } = options;

  // Transform DataTable filters
  const filters: RQLFilter[] = query.filters?.length
    ? query.filters.flatMap(filter => transformFilter(filter, fieldNameMapping))
    : [];

  // Build the RQLRequest with snake_case properties
  const rqlRequest = create(RQLRequestSchema, {
    filters,
    groupBy: (query.group_by || []).map(
      field => fieldNameMapping?.[field] ?? field
    ),
    offset: query.offset || 0,
    limit: query.limit || defaultLimit,
    sort: transformSort(query.sort || [], fieldNameMapping) || [],
    search: query.search || ''
  });

  return rqlRequest;
}
