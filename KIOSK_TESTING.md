# Giving Kiosk Testing Guide

## Setup

1. **Start the application:**
   ```bash
   docker compose up -d
   ```

2. **Access the application:**
   - Backend: http://localhost:8190
   - Frontend: http://localhost:5273

## Testing Steps

### 1. Configure Stripe Connect (if not already done)

1. Log into the admin dashboard
2. Navigate to `/dashboard/giving/settings`
3. Complete Stripe Connect onboarding
4. Wait for approval (or use Stripe test mode)

### 2. Configure Kiosk Settings

1. Navigate to `/dashboard/settings/kiosk`
2. Enable the giving kiosk (toggle switch)
3. Configure quick amounts (default: $10, $25, $50, $100, $250)
4. (Optional) Set a default fund
5. (Optional) Customize thank you message
6. Click "Save Settings"

### 3. Test the Kiosk

1. Open kiosk URL: `/giving-kiosk?tenant=<TENANT_ID>`
   - For local testing, get tenant ID from database or API response
   - Example: `http://localhost:5273/giving-kiosk?tenant=123e4567-e89b-12d3-a456-426614174000`

2. **Welcome Screen:** Click "Give Now"

3. **Amount Selection:**
   - Click a quick amount button ($10, $25, etc.)
   - OR enter a custom amount and click "Continue"

4. **Fund Selection:** Click a fund (General, Missions, Building, etc.)

5. **Optional Info:**
   - Enter name and email for receipt
   - OR click "Skip (No Receipt)"

6. **Payment:** Redirected to Stripe Checkout
   - Enter test card: `4242 4242 4242 4242`
   - Any future expiration date
   - Any 3-digit CVC
   - Any 5-digit ZIP

7. **Thank You:** After payment, see thank you screen
   - Auto-redirects to start after 10 seconds

### 4. Verify Inactivity Reset

1. Start a donation flow
2. Wait 60 seconds without interaction
3. Screen should automatically reset to welcome

### 5. Check Admin Dashboard

1. Navigate to `/dashboard/giving`
2. Verify the donation appears in the donations list
3. Check that it's marked as "kiosk" donation (metadata)

## API Endpoints

### Public Endpoints (No Auth)

- `GET /api/giving/kiosk/config?tenant_id=<ID>` - Get kiosk configuration
- `POST /api/giving/public/checkout` - Create checkout session

### Protected Endpoints (Auth Required)

- `GET /api/giving/kiosk` - Get kiosk config (admin)
- `PUT /api/giving/kiosk` - Update kiosk config (admin only)

## Deployment Notes

For production deployment:

1. **iPad Setup:**
   - Set orientation to landscape (1366x1024)
   - Enable browser full-screen mode
   - Enable Guided Access to prevent exiting
   - Disable sleep mode
   - Keep device plugged in

2. **URL Configuration:**
   - Update frontend URL in environment variables
   - Configure proper tenant ID detection (subdomain or query param)
   - Set up proper success/cancel URLs in Stripe service

3. **Security:**
   - Kiosk config requires admin role
   - Public checkout validates tenant has kiosk enabled
   - Stripe webhooks handle payment verification

## Troubleshooting

- **Kiosk not loading:** Check that kiosk is enabled in settings
- **Payment failing:** Verify Stripe Connect is fully configured
- **Wrong tenant:** Make sure tenant ID is passed correctly in URL
- **Donations not appearing:** Check Stripe webhook configuration

## Features Implemented

✅ Full-screen kiosk UI with touch-optimized buttons (80px+)  
✅ Step-by-step donation flow (6 steps)  
✅ Quick amount buttons (configurable)  
✅ Custom amount input  
✅ Fund selection  
✅ Optional donor name/email  
✅ Stripe Checkout integration  
✅ Thank you screen with auto-reset  
✅ 60-second inactivity timer  
✅ Admin configuration page  
✅ Database migration for kiosk settings  
✅ Public API endpoints (no auth required)  

## Next Steps (Future Enhancements)

- [ ] QR code for direct kiosk access
- [ ] Multi-language support
- [ ] Receipt printing capability
- [ ] Recurring donation option from kiosk
- [ ] Analytics dashboard for kiosk donations
- [ ] Customizable branding (logo, colors)
