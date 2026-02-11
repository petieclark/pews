#!/bin/bash
# Test Stripe webhook locally
# Usage: ./scripts/test-stripe-webhook.sh [event-type]

set -e

EVENT_TYPE="${1:-checkout.session.completed}"
WEBHOOK_URL="http://localhost:8190/api/billing/webhook"

echo "Testing Stripe webhook: $EVENT_TYPE"
echo "Webhook URL: $WEBHOOK_URL"
echo ""

case "$EVENT_TYPE" in
  "checkout.session.completed")
    PAYLOAD='{
      "id": "evt_test_webhook",
      "object": "event",
      "type": "checkout.session.completed",
      "data": {
        "object": {
          "id": "cs_test_123",
          "object": "checkout.session",
          "customer": "cus_test_123",
          "metadata": {
            "tenant_id": "test-tenant-id"
          },
          "subscription": {
            "id": "sub_test_123",
            "current_period_start": '$(date +%s)',
            "current_period_end": '$(date -v+1m +%s 2>/dev/null || date -d "+1 month" +%s)'
          }
        }
      }
    }'
    ;;
    
  "customer.subscription.updated")
    PAYLOAD='{
      "id": "evt_test_webhook",
      "object": "event",
      "type": "customer.subscription.updated",
      "data": {
        "object": {
          "id": "sub_test_123",
          "object": "subscription",
          "status": "active",
          "metadata": {
            "tenant_id": "test-tenant-id"
          },
          "current_period_start": '$(date +%s)',
          "current_period_end": '$(date -v+1m +%s 2>/dev/null || date -d "+1 month" +%s)'
        }
      }
    }'
    ;;
    
  "customer.subscription.deleted")
    PAYLOAD='{
      "id": "evt_test_webhook",
      "object": "event",
      "type": "customer.subscription.deleted",
      "data": {
        "object": {
          "id": "sub_test_123",
          "object": "subscription",
          "metadata": {
            "tenant_id": "test-tenant-id"
          }
        }
      }
    }'
    ;;
    
  "invoice.payment_failed")
    PAYLOAD='{
      "id": "evt_test_webhook",
      "object": "event",
      "type": "invoice.payment_failed",
      "data": {
        "object": {
          "id": "in_test_123",
          "object": "invoice",
          "subscription": {
            "id": "sub_test_123",
            "metadata": {
              "tenant_id": "test-tenant-id"
            },
            "current_period_start": '$(date +%s)',
            "current_period_end": '$(date -v+1m +%s 2>/dev/null || date -d "+1 month" +%s)'
          }
        }
      }
    }'
    ;;
    
  *)
    echo "Unknown event type: $EVENT_TYPE"
    echo "Supported types: checkout.session.completed, customer.subscription.updated, customer.subscription.deleted, invoice.payment_failed"
    exit 1
    ;;
esac

echo "Payload:"
echo "$PAYLOAD" | jq .
echo ""

# Note: This will fail signature verification since we don't have the webhook secret
# To test properly, use the Stripe CLI: stripe trigger <event-type>
echo "Sending webhook (will fail signature verification without proper Stripe signature)..."
curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD"

echo ""
echo ""
echo "Note: For proper testing with signature verification, use Stripe CLI:"
echo "  stripe listen --forward-to $WEBHOOK_URL"
echo "  stripe trigger $EVENT_TYPE"
