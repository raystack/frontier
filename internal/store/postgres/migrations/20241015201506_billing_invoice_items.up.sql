ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS items jsonb DEFAULT '{}';
