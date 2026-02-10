package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

func (h *Handler) HandleWebhook(webhookSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(body, signature, webhookSecret)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		switch event.Type {
		case "checkout.session.completed":
			var session stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			h.handleCheckoutCompleted(&session)

		case "invoice.paid":
			var invoice stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			h.handleInvoicePaid(&invoice)

		case "invoice.payment_failed":
			var invoice stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			h.handlePaymentFailed(&invoice)

		case "customer.subscription.deleted":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			h.handleSubscriptionDeleted(&sub)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) handleCheckoutCompleted(session *stripe.CheckoutSession) {
	tenantID := session.Metadata["tenant_id"]
	if tenantID == "" {
		fmt.Println("No tenant_id in metadata")
		return
	}

	// Note: context.Background() used since we're in an async webhook handler
	ctx := context.Background()

	// Update subscription with Stripe details
	err := h.service.UpdateSubscriptionStatus(
		ctx,
		tenantID,
		session.Subscription.ID,
		"pro",
		"active",
		time.Unix(session.Subscription.CurrentPeriodStart, 0),
		time.Unix(session.Subscription.CurrentPeriodEnd, 0),
	)
	if err != nil {
		fmt.Printf("Failed to update subscription: %v\n", err)
	}
}

func (h *Handler) handleInvoicePaid(invoice *stripe.Invoice) {
	// Update subscription status to active
	tenantID := invoice.Subscription.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	ctx := context.Background()
	err := h.service.UpdateSubscriptionStatus(
		ctx,
		tenantID,
		invoice.Subscription.ID,
		"pro",
		"active",
		time.Unix(invoice.Subscription.CurrentPeriodStart, 0),
		time.Unix(invoice.Subscription.CurrentPeriodEnd, 0),
	)
	if err != nil {
		fmt.Printf("Failed to update subscription: %v\n", err)
	}
}

func (h *Handler) handlePaymentFailed(invoice *stripe.Invoice) {
	// Mark subscription as past_due
	tenantID := invoice.Subscription.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	ctx := context.Background()
	err := h.service.UpdateSubscriptionStatus(
		ctx,
		tenantID,
		invoice.Subscription.ID,
		"pro",
		"past_due",
		time.Unix(invoice.Subscription.CurrentPeriodStart, 0),
		time.Unix(invoice.Subscription.CurrentPeriodEnd, 0),
	)
	if err != nil {
		fmt.Printf("Failed to update subscription: %v\n", err)
	}
}

func (h *Handler) handleSubscriptionDeleted(sub *stripe.Subscription) {
	tenantID := sub.Metadata["tenant_id"]
	if tenantID == "" {
		return
	}

	ctx := context.Background()
	err := h.service.CancelSubscription(ctx, tenantID)
	if err != nil {
		fmt.Printf("Failed to cancel subscription: %v\n", err)
	}
}
