import { create } from '@bufbuild/protobuf';
import { RQLFilterSchema } from '@raystack/proton/frontier';
import { INVOICE_STATES } from './constants';

// An invoice the customer still has to pay: open state with a non-zero
// amount. This is defined once here so every caller means the same thing
// by it.
export function openInvoiceFilters() {
  return [
    create(RQLFilterSchema, {
      name: 'state',
      operator: 'eq',
      value: { case: 'stringValue', value: INVOICE_STATES.OPEN }
    }),
    create(RQLFilterSchema, {
      name: 'amount',
      operator: 'gt',
      value: { case: 'numberValue', value: 0 }
    })
  ];
}
