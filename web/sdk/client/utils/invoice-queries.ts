import { create } from '@bufbuild/protobuf';
import { RQLFilterSchema } from '@raystack/proton/frontier';
import { INVOICE_STATES } from './constants';

// The one definition of "an invoice the customer still has to pay": open
// state with a non-zero amount. The server refuses an organization delete
// while any exist, and the billing page's payment-issue banner keys off the
// same set — both build their queries from these filters so the two can not
// drift apart.
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
