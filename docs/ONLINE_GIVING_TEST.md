# Online Giving Testing Guide

## Overview
This guide covers testing the online donation flow with Stripe Checkout.

## Setup

1. **Start the application:**
   ```bash
   docker compose up -d
   ```

2. **Create a test church and fund:**
   - Register a new church at http://localhost:5273/register
   - Log in and enable the "Giving" module in Settings
   - Go to Giving → Funds and create at least one fund (e.g., "General Fund")
   - Note your church's slug (visible in the URL or settings)

3. **Set up Stripe Connect (optional for testing):**
   - Go to Giving → Settings
   - Click "Connect with Stripe" (this will use test mode if STRIPE_SECRET_KEY is test key)
   - Complete the Express onboarding flow

## Testing the Public Give Page

### Access the page
Navigate to: `http://localhost:5273/give?tenant=YOUR-CHURCH-SLUG`

### Test the flow

1. **Enter church slug** (if not in URL)
   - Should auto-load available funds

2. **Select a fund** from the dropdown

3. **Enter donation details:**
   - Amount: `25.00`
   - Name: `Test Donor`
   - Email: `donor@example.com`

4. **Click "Continue to Payment"**
   - Should redirect to Stripe Checkout

5. **Complete test payment:**
   - Use test card: `4242 4242 4242 4242`
   - Expiry: Any future date
   - CVC: Any 3 digits
   - ZIP: Any 5 digits

6. **Verify success:**
   - Should redirect back to `/give?tenant=YOUR-SLUG&status=success`
   - Success message should display

## Backend Verification

### Check donation was recorded
```bash
docker compose exec postgres psql -U pews -d pews -c "SELECT id, amount_cents, status, donated_at FROM donations ORDER BY created_at DESC LIMIT 5;"
```

### Check donor was created
```bash
docker compose exec postgres psql -U pews -d pews -c "SELECT id, first_name, last_name, email FROM people WHERE email = 'donor@example.com';"
```

### View webhook events (backend logs)
```bash
docker compose logs -f backend | grep webhook
```

## Test Scenarios

### Scenario 1: New Donor
- Use a new email address
- Should create new person record automatically
- Should split "First Last" into first_name and last_name

### Scenario 2: Existing Donor
- Use an email that already exists in People
- Should link donation to existing person
- Should NOT create duplicate person record

### Scenario 3: Multiple Funds
- Create multiple funds (General, Building, Missions, etc.)
- Default fund should be pre-selected
- User can switch between funds

### Scenario 4: Canceled Payment
- Start checkout flow
- Click "Back" in Stripe Checkout
- Should return to `/give?tenant=SLUG&status=canceled`
- Should show cancelation message

## API Testing (curl)

### List public funds:
```bash
curl http://localhost:8190/api/giving/public/funds?tenant=YOUR-SLUG
```

### Create checkout session:
```bash
curl -X POST http://localhost:8190/api/giving/public/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_slug": "YOUR-SLUG",
    "fund_id": "FUND-UUID",
    "amount_cents": 5000,
    "donor_name": "Jane Doe",
    "donor_email": "jane@example.com"
  }'
```

## Webhook Testing

### Using Stripe CLI:
```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Forward webhooks to local backend
stripe listen --forward-to localhost:8190/api/giving/webhook

# Trigger test event
stripe trigger checkout.session.completed
```

## Troubleshooting

### Funds not loading
- Check church slug is correct
- Verify at least one fund exists and is_active = true
- Check browser console for API errors

### Checkout fails with "church has not completed onboarding"
- Church needs to complete Stripe Connect onboarding first
- OR set a test Stripe account ID directly in database:
  ```sql
  UPDATE tenants SET stripe_account_id = 'acct_test123' WHERE slug = 'YOUR-SLUG';
  ```

### Donation not appearing after payment
- Check webhook endpoint is accessible
- Verify STRIPE_WEBHOOK_SECRET matches
- Check backend logs for webhook processing errors

### Person not being created
- Check email format is valid
- Verify people table has no unique constraint violations
- Check backend logs for "Failed to find/create donor" messages

## Production Checklist

Before deploying to production:

- [ ] Replace test Stripe keys with live keys
- [ ] Update FRONTEND_URL to production domain
- [ ] Configure webhook endpoint at https://dashboard.stripe.com/webhooks
- [ ] Test with real bank card (small amount)
- [ ] Verify email receipts are sent (when email service is wired)
- [ ] Set up monitoring for failed webhooks
- [ ] Document give page URL for churches: `https://yourdomain.com/give?tenant=SLUG`
