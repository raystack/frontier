ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS current_period_start_at timestamptz;
ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS current_period_end_at timestamptz;
ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS billing_cycle_anchor_at timestamptz;

CREATE TABLE billing_invoices (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  provider_id text NOT NULL UNIQUE,
  customer_id uuid NOT NULL REFERENCES billing_customers(id),

  amount int8 NOT NULL,
  currency text NOT NULL,
  hosted_url text,
  state text NOT NULL DEFAULT 'pending',
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,

  due_at timestamptz,
  effective_at timestamptz,
  period_start_at timestamptz,
  period_end_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS billing_invoices_customer_id_idx ON billing_invoices(customer_id);