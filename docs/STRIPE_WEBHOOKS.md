# Stripe Webhooks Integration

## Overview

Pews uses Stripe webhooks to handle subscription lifecycle events. This ensures subscription status stays in sync between Stripe and the Pews database.

## Architecture

- **Webhook Endpoint**: `POST /api/billing/webhook`
- **Handler**: `internal/billing/webhook.go`
- **Service**: `internal/billing/service.go`
- **No Authentication**: Webhook endpoint is public but validates Stripe signatures

## Environment Variables

Required environment variables in `docker-compose.yml`:

```yaml
STRIPE_SECRET_KEY: sk_test_51SzTAJJDZY7DAhKH...
STRIPE_WEBHOOK_SECRET: whsec_...
STRIPE_PRICE_ID: price_...
```

## Supported Events

### 1. `checkout.session.completed`
- **Trigger**: Customer completes checkout and creates a subscription
- **Action**: 
  - Updates tenant subscription status to `active`
  - Stores Stripe subscription ID
  - Sets plan to `pro`
  - Records billing period dates

### 2. `customer.subscription.updated`
- **Trigger**: Subscription is modified (plan change, billing cycle, etc.)
- **Action**:
  - Syncs plan changes
  - Updates subscription status
  - Adjusts plan tier based on subscription status

### 3. `customer.subscription.deleted`
- **Trigger**: Subscription is canceled
- **Action**:
  - Downgrades tenant to `free` tier
  - Marks subscription as `canceled`

### 4. `invoice.payment_failed`
- **Trigger**: Payment attempt fails
- **Action**:
  - Marks subscription as `past_due`
  - Logs event for notification
  - Keeps plan as `pro` (grace period)

## Metadata Requirements

All Stripe objects must include `tenant_id` in metadata:

```go
Metadata: map[string]string{
    "tenant_id": tenantID,
}
```

This is automatically added during:
- Checkout session creation
- Subscription creation (via `SubscriptionData.Metadata`)

## Checkout Flow

1. **Frontend**: User clicks "Upgrade to Pro"
2. **Backend**: `POST /api/billing/checkout`
   - Creates or retrieves Stripe customer
   - Creates checkout session with `tenant_id` in metadata
   - Returns checkout URL
3. **Frontend**: Redirects to Stripe Checkout
4. **User**: Completes payment on Stripe
5. **Stripe**: Sends `checkout.session.completed` webhook
6. **Backend**: Activates subscription in database
7. **Frontend**: Success redirect to `/settings?session_id={CHECKOUT_SESSION_ID}`

## Creating the Stripe Product & Price

The Pro plan is $100/month. To create it:

### Option 1: Via Stripe Dashboard
1. Go to https://dashboard.stripe.com/test/products
2. Create product "Pews Pro"
3. Add price: $100.00/month, recurring
4. Copy price ID (starts with `price_`)
5. Set `STRIPE_PRICE_ID` environment variable

### Option 2: Via Code
```go
priceID, err := billingService.EnsureStripeProduct()
```

This creates:
- Product: "Pews Pro"
- Price: $100.00/month USD recurring

## Testing

### Local Testing (without signature verification)

```bash
chmod +x scripts/test-stripe-webhook.sh
./scripts/test-stripe-webhook.sh checkout.session.completed
./scripts/test-stripe-webhook.sh customer.subscription.updated
./scripts/test-stripe-webhook.sh customer.subscription.deleted
./scripts/test-stripe-webhook.sh invoice.payment_failed
```

**Note**: These requests will fail signature verification. Use for payload structure testing only.

### Proper Testing with Stripe CLI

1. **Install Stripe CLI**:
   ```bash
   brew install stripe/stripe-cli/stripe
   # or
   # https://stripe.com/docs/stripe-cli
   ```

2. **Login**:
   ```bash
   stripe login
   ```

3. **Forward webhooks to local**:
   ```bash
   stripe listen --forward-to http://localhost:8190/api/billing/webhook
   ```
   
   This will output a webhook signing secret like `whsec_...`
   
4. **Update environment**:
   ```bash
   export STRIPE_WEBHOOK_SECRET=whsec_...
   # Restart backend to pick up new secret
   ```

5. **Trigger test events**:
   ```bash
   stripe trigger checkout.session.completed
   stripe trigger customer.subscription.updated
   stripe trigger customer.subscription.deleted
   stripe trigger invoice.payment_failed
   ```

### Production Webhook Setup

1. Go to https://dashboard.stripe.com/webhooks
2. Click "Add endpoint"
3. URL: `https://api.pews.app/api/billing/webhook` (or your production URL)
4. Select events:
   - `checkout.session.completed`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_failed`
5. Copy the webhook signing secret
6. Set `STRIPE_WEBHOOK_SECRET` in production environment

## Database Schema

Subscriptions table:
```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    plan TEXT NOT NULL DEFAULT 'free',
    status TEXT NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Error Handling

- All webhook handlers log errors via `log.Printf`
- Webhook endpoint returns appropriate HTTP status codes:
  - `200 OK`: Event processed successfully
  - `400 Bad Request`: Invalid signature or malformed payload
  - `500 Internal Server Error`: Processing failed
- Failed webhook events can be retried from Stripe Dashboard

## Security

- **Signature Verification**: Every webhook request is verified using `stripe.webhook.ConstructEvent`
- **No Auth Middleware**: Webhook endpoint is public (Stripe needs to call it)
- **Metadata Validation**: tenant_id is validated before processing
- **Idempotency**: Webhook handlers are idempotent (safe to retry)

## Monitoring

Check webhook delivery in:
- Stripe Dashboard → Developers → Webhooks → View logs
- Backend logs: `docker compose logs -f backend | grep "Stripe webhook"`

## Troubleshooting

### Webhook fails signature verification
- Verify `STRIPE_WEBHOOK_SECRET` matches the secret from Stripe dashboard
- Check that webhook endpoint is publicly accessible
- Ensure no middleware modifies the request body

### Subscription not updating
- Check that `tenant_id` is in metadata
- Verify subscription ID matches database
- Check backend logs for error messages

### Testing locally with ngrok
```bash
ngrok http 8190
# Copy ngrok URL
stripe listen --forward-to https://YOUR-NGROK-URL.ngrok.io/api/billing/webhook
```
