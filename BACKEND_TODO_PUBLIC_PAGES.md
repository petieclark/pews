# Backend TODO: Public Pages Support

This document outlines the backend changes needed to support the new public-facing pages.

## ✅ Already Working (No Changes Needed)

### Watch Page (`/watch/[id]`)
- ✅ `GET /api/streaming/watch/{id}` - Public endpoint exists
- ✅ `GET /api/streaming/{id}/chat` - Public endpoint exists
- ✅ `POST /api/streaming/{id}/chat` - Public guest chat exists
- ✅ `POST /api/streaming/{id}/join` - Public join exists
- ✅ `POST /api/streaming/{id}/leave` - Public leave exists

### Connection Card (`/connect`)
- ✅ `POST /api/communication/cards` - Public endpoint exists (line 63 in router.go)

## 🔴 Required Backend Changes

### 1. Public Giving Page (`/give`)

Currently, the `/give` page needs a public checkout endpoint for guest donations.

#### Required: New Public Endpoint

**Endpoint:** `POST /api/giving/public/checkout`

**No authentication required** - This is for guest giving.

**Request Body:**
```json
{
  "fund_id": "uuid",
  "amount_cents": 5000,
  "guest_name": "John Doe",
  "guest_email": "john@example.com"
}
```

**Response:**
```json
{
  "url": "https://checkout.stripe.com/c/pay/..."
}
```

**Implementation Notes:**
1. Add to `internal/router/router.go` in the public routes section (after line 63):
   ```go
   r.Post("/api/giving/public/checkout", givingHandler.CreatePublicCheckout)
   ```

2. Add to `internal/giving/handler.go`:
   ```go
   func (h *Handler) CreatePublicCheckout(w http.ResponseWriter, r *http.Request) {
       var req struct {
           FundID      string `json:"fund_id"`
           AmountCents int    `json:"amount_cents"`
           GuestName   string `json:"guest_name"`
           GuestEmail  string `json:"guest_email"`
       }
       
       if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
           http.Error(w, "Invalid request", http.StatusBadRequest)
           return
       }
       
       // Extract tenant_id from context (set by TenantExtractor middleware)
       tenantID := middleware.GetTenantIDFromContext(r.Context())
       if tenantID == "" {
           http.Error(w, "Tenant not found", http.StatusBadRequest)
           return
       }
       
       // Create checkout session (similar to existing CreateCheckout but for guests)
       url, err := h.stripeService.CreatePublicCheckoutSession(
           r.Context(),
           tenantID,
           req.FundID,
           req.AmountCents,
           req.GuestName,
           req.GuestEmail,
       )
       
       if err != nil {
           http.Error(w, "Failed to create checkout: "+err.Error(), http.StatusInternalServerError)
           return
       }
       
       w.Header().Set("Content-Type", "application/json")
       json.NewEncoder(w).Encode(map[string]string{"url": url})
   }
   ```

3. Add to `internal/giving/stripe.go`:
   ```go
   func (s *StripeService) CreatePublicCheckoutSession(
       ctx context.Context,
       tenantID, fundID string,
       amountCents int,
       guestName, guestEmail string,
   ) (string, error) {
       // Get tenant's Stripe account ID
       var stripeAccountID *string
       err := s.db.QueryRow(ctx,
           `SELECT stripe_account_id FROM tenants WHERE id = $1`,
           tenantID,
       ).Scan(&stripeAccountID)
       if err != nil {
           return "", err
       }
       
       if stripeAccountID == nil || *stripeAccountID == "" {
           return "", fmt.Errorf("church has not completed Stripe Connect onboarding")
       }
       
       // Get fund name
       var fundName string
       err = s.db.QueryRow(ctx, `SELECT name FROM funds WHERE id = $1`, fundID).Scan(&fundName)
       if err != nil {
           return "", fmt.Errorf("fund not found")
       }
       
       // Calculate application fee (1%)
       appFeeCents := amountCents / 100
       if appFeeCents < 30 {
           appFeeCents = 30 // Minimum 30 cents
       }
       
       // Create checkout session
       params := &stripe.CheckoutSessionParams{
           Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
           CustomerEmail:     stripe.String(guestEmail),
           LineItems: []*stripe.CheckoutSessionLineItemParams{
               {
                   PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                       Currency: stripe.String("usd"),
                       ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                           Name:        stripe.String("Donation to " + fundName),
                           Description: stripe.String("Online giving"),
                       },
                       UnitAmount: stripe.Int64(int64(amountCents)),
                   },
                   Quantity: stripe.Int64(1),
               },
           },
           PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
               ApplicationFeeAmount: stripe.Int64(int64(appFeeCents)),
           },
           // Update success/cancel URLs for public page
           SuccessURL: stripe.String(s.frontendURL + "/give?status=success"),
           CancelURL:  stripe.String(s.frontendURL + "/give?status=canceled"),
           Metadata: map[string]string{
               "tenant_id":  tenantID,
               "fund_id":    fundID,
               "guest_name": guestName,
               "is_guest":   "true", // Flag to indicate guest donation
           },
       }
       params.SetStripeAccount(*stripeAccountID)
       
       sess, err := session.New(params)
       if err != nil {
           return "", fmt.Errorf("failed to create checkout session: %w", err)
       }
       
       return sess.URL, nil
   }
   ```

4. **Update webhook handler** in `HandlePaymentIntentSucceeded` to handle guest donations:
   - Check for `is_guest` metadata flag
   - If `is_guest == "true"`, create donation record with `person_id = NULL`
   - Store `guest_name` in donation metadata for receipts

#### Optional: Public Funds List

**Endpoint:** `GET /api/giving/public/funds`

**No authentication required**

Currently the frontend falls back to a hardcoded "General Fund" if the authenticated endpoint fails. For a better experience, create a public funds endpoint:

```go
// In router.go (public section)
r.Get("/api/giving/public/funds", givingHandler.ListPublicFunds)

// In handler.go
func (h *Handler) ListPublicFunds(w http.ResponseWriter, r *http.Request) {
    tenantID := middleware.GetTenantIDFromContext(r.Context())
    if tenantID == "" {
        http.Error(w, "Tenant not found", http.StatusBadRequest)
        return
    }
    
    funds, err := h.service.ListActiveFunds(r.Context(), tenantID)
    if err != nil {
        http.Error(w, "Failed to list funds: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(funds)
}

// In service.go
func (s *Service) ListActiveFunds(ctx context.Context, tenantID string) ([]Fund, error) {
    rows, err := s.db.Query(ctx,
        `SELECT id, name, description, is_default, is_active, created_at, updated_at
         FROM funds
         WHERE tenant_id = $1 AND is_active = true
         ORDER BY is_default DESC, name ASC`,
        tenantID,
    )
    // ... rest of implementation
}
```

Then update the frontend to use the public endpoint:
```javascript
const response = await fetch(`${API_URL}/api/giving/public/funds`);
```

## Testing

1. **Watch Page:** Visit `http://localhost:5273/watch/test` (create a test stream first)
2. **Connection Card:** Visit `http://localhost:5273/connect` - should submit successfully
3. **Give Page:** Visit `http://localhost:5273/give` - after backend changes, should redirect to Stripe

## Summary

**Must Have:**
- Public checkout endpoint for guest donations

**Nice to Have:**
- Public funds list endpoint (currently has fallback)

All frontend code is ready and waiting for these backend changes.
