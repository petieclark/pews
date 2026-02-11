# Progressive Web App (PWA) Configuration

## Overview

Pews is now configured as a Progressive Web App, allowing churches to install it on their mobile devices for a native app-like experience.

## Features

### 1. **Web App Manifest**
- **Location:** `web/static/manifest.json`
- **Theme Color:** #1B3A4B (navy)
- **Background:** #F7FAFA
- **Display Mode:** Standalone
- **Start URL:** `/dashboard`
- **Icons:** 8 sizes from 72x72 to 512x512

### 2. **Service Worker**
- **Location:** `web/static/sw.js`
- **Strategies:**
  - **Static Assets:** Cache-first with network fallback
  - **API Calls:** Network-first with cache fallback
  - **Offline Pages:** Fallback to `/offline` route
- **Auto-Update:** Prompts user when new version available
- **Skip Waiting:** Updates activate immediately with user consent

### 3. **Offline Support**
- **Page:** `web/src/routes/offline/+page.svelte`
- **Features:**
  - Branded offline message
  - Retry connection button
  - Church-themed design
  - User-friendly messaging

### 4. **Install Prompt**
- **Component:** `web/src/lib/InstallPrompt.svelte`
- **Behavior:**
  - Detects `beforeinstallprompt` event
  - Shows custom branded install banner
  - Session-based (shows once per session)
  - Dismiss or Install options
  - Auto-hides after install

### 5. **PWA Icons**
Generated from `landing/icon.png` in the following sizes:
- 72x72, 96x96, 128x128, 144x144
- 152x152, 192x192, 384x384, 512x512

All icons support both regular and maskable display.

### 6. **Meta Tags**
Added to `web/src/app.html`:
- PWA manifest link
- Theme color
- Viewport with safe area support
- Apple Web App capable
- Apple status bar style
- Apple touch icon

## Testing

### Local Development
```bash
cd ~/Projects/pews
docker compose up -d
```

Access at: http://localhost:5173 (or configured port)

### Mobile Testing (Chrome/Edge)
1. Open the app in Chrome on Android or Edge
2. Look for "Install" prompt in browser
3. Or use the custom install banner at bottom of screen
4. Check: Menu → "Add to Home screen"

### iOS Testing (Safari)
1. Open the app in Safari on iOS
2. Tap Share button
3. Select "Add to Home Screen"
4. App icon will appear on home screen

### Lighthouse Audit
```bash
# Run Lighthouse PWA audit
npx lighthouse http://localhost:5173 --view --preset=desktop
```

Check for:
- ✅ Installable
- ✅ PWA optimized
- ✅ Service worker registered
- ✅ Offline fallback
- ✅ Themed
- ✅ Viewport configured

## Implementation Details

### Cache Strategy
- **Static assets** (JS, CSS, images): Cached on first load, served from cache
- **API requests** (`/api/*`): Always try network first, fallback to cache
- **Navigation**: Offline page shown when network unavailable

### Update Flow
1. New service worker detected
2. User prompted: "New version available! Reload to update?"
3. If accepted: Skip waiting → activate new SW → reload page
4. If dismissed: Update on next page load

### Session Persistence
- Install prompt shows once per session
- Stored in `sessionStorage`
- Resets on browser close/new tab

## Browser Support
- ✅ Chrome/Edge (Android): Full support
- ✅ Safari (iOS): Manual install via Share sheet
- ✅ Firefox: Basic PWA support
- ❌ Safari (macOS): Limited PWA support

## Troubleshooting

### Install prompt not showing
- Check DevTools → Application → Manifest
- Ensure HTTPS (required in production)
- Clear cache and reload
- Check browser console for errors

### Service worker not registering
- Check `sw.js` is accessible at `/sw.js`
- Verify no console errors
- Check DevTools → Application → Service Workers
- Ensure not in private/incognito mode

### Offline page not showing
- Check service worker is active
- Verify `/offline` route exists
- Test by disabling network in DevTools

## Production Deployment

Ensure these headers are set for production:
```
# Service Worker (sw.js)
Cache-Control: no-cache

# Manifest
Cache-Control: public, max-age=3600

# Static assets
Cache-Control: public, max-age=31536000, immutable
```

HTTPS is required for PWA features to work in production.
