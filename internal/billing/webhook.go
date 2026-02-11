package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

func (h *Handler) HandleWebhook(webhookSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading webhook body: %v", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(body, signature, webhookSecret)
		if err != nil {
			log.Printf("Error verifying webhook signature: %v", err)
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		log.Printf("Received Stripe webhook event: %s", event.Type)

		switch event.Type {
		case "checkout.session.completed":
			var session stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
				log.Printf("Error parsing checkout.session.completed: %v", err)
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			if err := h.handleCheckoutCompleted(&session); err != nil {
				log.Printf("Error handling checkout.session.completed: %v", err)
				http.Error(w, "Failed to process event", http.StatusInternalServerError)
				return
			}

		case "customer.subscription.updated":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Printf("Error parsing customer.subscription.updated: %v", err)
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			if err := h.handleSubscriptionUpdated(&sub); err != nil {
				log.Printf("Error handling customer.subscription.updated: %v", err)
				http.Error(w, "Failed to process event", http.StatusInternalServerError)
				return
			}

		case "customer.subscription.deleted":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
				log.Printf("Error parsing customer.subscription.deleted: %v", err)
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			if err := h.handleSubscriptionDeleted(&sub); err != nil {
				log.Printf("Error handling customer.subscription.deleted: %v", err)
				http.Error(w, "Failed to process event", http.StatusInternalServerError)
				return
			}

		case "invoice.payment_failed":
			var invoice stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
				log.Printf("Error parsing invoice.payment_failed: %v", err)
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}
			if err := h.handlePaymentFailed(&invoice); err != nil {
				log.Printf("Error handling invoice.payment_failed: %v", err)
				http.Error(w, "Failed to process event", http.StatusInternalServerError)
				return
			}

		default:
			log.Printf("Unhandled webhook event type: %s", event.Type)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) handleCheckoutCompleted(session *stripe.CheckoutSession) error {
	tenantID := session.Metadata["tenant_id"]
	if tenantID == "" {
		return fmt.Errorf("no tenant_id in checkout session metadata")
	}

	if session.Subscription == nil {
		return fmt.Errorf("no subscription in checkout session")
	}

	ctx := context.Background()

	// Fetch the full subscription details to get all necessary info
	subscriptionID := session.Subscription.ID
	log.Printf("Checkout completed for tenant %s, subscription %s", tenantID, subscriptionID)

	// Update subscription with Stripe details
	err := h.service.UpdateSubscriptionStatus(
		ctx,
		tenantID,
		subscriptionID,
		"pro",
		"active",
		time.Unix(session.Subscription.CurrentPeriodStart, 0),
		time.Unix(session.Subscription.CurrentPeriodEnd, 0),
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	log.Printf("Successfully activated subscription for tenant %s", tenantID)
	return nil
}

func (h *Handler) handleSubscriptionUpdated(sub *stripe.Subscription) error {
	// Try to get tenant_id from subscription metadata
	tenantID := sub.Metadata["tenant_id"]
	if tenantID == "" {
		// Try to get tenant_id from customer metadata
		if sub.Customer != nil && sub.Customer.Metadata != nil {
			tenantID = sub.Customer.Metadata["tenant_id"]
		}
		if tenantID == "" {
			return fmt.Errorf("no tenant_id in subscription or customer metadata")
		}
	}

	ctx := context.Background()

	// Determine plan based on subscription status and items
	plan := "free"
	status := string(sub.Status)
	
	if sub.Status == stripe.SubscriptionStatusActive || 
	   sub.Status == stripe.SubscriptionStatusTrialing {
		plan = "pro"
	} else if sub.Status == stripe.SubscriptionStatusCanceled ||
	          sub.Status == stripe.SubscriptionStatusUnpaid {
		plan = "free"
	}

	log.Printf("Subscription updated for tenant %s: plan=%s, status=%s", tenantID, plan, status)

	err := h.service.UpdateSubscriptionStatus(
		ctx,
		tenantID,
		sub.ID,
		plan,
		status,
		time.Unix(sub.CurrentPeriodStart, 0),
		time.Unix(sub.CurrentPeriodEnd, 0),
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (h *Handler) handleSubscriptionDeleted(sub *stripe.Subscription) error {
	// Try to get tenant_id from subscription metadata
	tenantID := sub.Metadata["tenant_id"]
	if tenantID == "" {
		// Try to get tenant_id from customer metadata
		if sub.Customer != nil && sub.Customer.Metadata != nil {
			tenantID = sub.Customer.Metadata["tenant_id"]
		}
		if tenantID == "" {
			return fmt.Errorf("no tenant_id in subscription or customer metadata")
		}
	}

	ctx := context.Background()
	
	log.Printf("Subscription deleted for tenant %s, downgrading to free", tenantID)
	
	err := h.service.CancelSubscription(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

func (h *Handler) handlePaymentFailed(invoice *stripe.Invoice) error {
	if invoice.Subscription == nil {
		return fmt.Errorf("no subscription in invoice")
	}

	// Try to get tenant_id from subscription metadata
	tenantID := invoice.Subscription.Metadata["tenant_id"]
	if tenantID == "" {
		// Try to get tenant_id from customer metadata
		if invoice.Customer != nil && invoice.Customer.Metadata != nil {
			tenantID = invoice.Customer.Metadata["tenant_id"]
		}
		if tenantID == "" {
			return fmt.Errorf("no tenant_id in subscription or customer metadata")
		}
	}

	ctx := context.Background()
	
	log.Printf("Payment failed for tenant %s, marking as past_due", tenantID)
	
	err := h.service.UpdateSubscriptionStatus(
		ctx,
		tenantID,
		invoice.Subscription.ID,
		"pro", // Keep plan as pro but mark status as past_due
		"past_due",
		time.Unix(invoice.Subscription.CurrentPeriodStart, 0),
		time.Unix(invoice.Subscription.CurrentPeriodEnd, 0),
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// TODO: Send notification to tenant about failed payment
	log.Printf("Payment failed notification needed for tenant %s", tenantID)
	
	return nil
}
