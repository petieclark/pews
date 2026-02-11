# SEO & Performance Improvements

## Branch: `feat/seo-performance`

This document outlines all SEO and performance optimizations implemented for the Pews project.

---

## 🔍 SEO Improvements

### Main Web Application (`~/Projects/pews/web/`)

#### 1. **Meta Tags & Social Sharing**
- ✅ Created reusable `SEO.svelte` component (`src/lib/components/SEO.svelte`)
- ✅ Supports dynamic title, description, Open Graph, Twitter Cards
- ✅ Canonical URLs for all pages
- ✅ Integrated into watch page with server-side data

#### 2. **Structured Data (JSON-LD)**
- ✅ Organization schema in root layout (`+layout.server.js`)
- ✅ VideoObject schema for live stream pages (`watch/[id]/+page.server.js`)
- ✅ BroadcastEvent schema for live streams
- ✅ Search engines can properly index church and streaming content

#### 3. **Sitemap & Robots**
- ✅ Created `static/sitemap.xml` with public pages
- ✅ Created `static/robots.txt` blocking dashboard, allowing public pages
- ✅ Note: Dynamic watch URLs should be generated via build script or server-side

#### 4. **Server-Side Rendering**
- ✅ Added `+page.server.js` for watch pages to generate SEO meta tags server-side
- ✅ Ensures crawlers see proper meta tags before JavaScript loads

---

### Landing Page (`~/Projects/pews-landing/`)

#### 1. **Meta Tags for All Pages**
- ✅ `index.html`: Added Open Graph, Twitter Cards, canonical URL
- ✅ `pricing.html`: Added Open Graph, Twitter Cards, canonical URL
- ✅ `about.html`: Added Open Graph, Twitter Cards, canonical URL
- ✅ `faq.html`: Added Open Graph, Twitter Cards, canonical URL

#### 2. **Structured Data**
- ✅ Added SoftwareApplication schema to `index.html`
- ✅ Includes pricing information, app category, operating system

#### 3. **Sitemap & Robots**
- ✅ Created `sitemap.xml` with all landing pages
- ✅ Created `robots.txt` allowing all pages

---

## ⚡ Performance Improvements

### Main Web Application

#### 1. **Image Optimization**
- ✅ All images in watch page use lazy loading (already implemented)
- ✅ Iframe embeds load on-demand

#### 2. **Font Optimization**
- ✅ Google Fonts already use `display=swap` parameter
- ✅ Added preconnect hints in `app.html` for faster font loading

#### 3. **API Preconnect**
- ✅ Added preconnect and dns-prefetch for API domain in `app.html`
- ✅ Reduces latency for API calls

#### 4. **Build Optimization** (`vite.config.js`)
- ✅ Enabled Terser minification
- ✅ Configured manual chunk splitting for vendor code
- ✅ Separated Svelte core into separate chunk for better caching
- ✅ Target ES2020 for modern browsers (smaller bundles)

#### 5. **Tree Shaking**
- ✅ Vite + SvelteKit automatically tree-shake unused code
- ✅ Manual chunking ensures vendor code is cached separately

---

### Landing Page

#### 1. **Image Optimization**
- ✅ All screenshots use `loading="lazy"` attribute
- ✅ Logo/icons load immediately (above the fold)

#### 2. **Font Optimization**
- ✅ Google Fonts use `display=swap` parameter
- ✅ Preconnect hints for fonts.googleapis.com and fonts.gstatic.com

#### 3. **Compression**
- Note: Static HTML files should be served with gzip/brotli compression
- This should be configured at the web server level (Nginx, Caddy, etc.)

---

## 🧪 Testing Recommendations

### Main Application
1. **Lighthouse Audit**
   ```bash
   npm run build
   npm run preview
   # Open Chrome DevTools > Lighthouse
   # Test: http://localhost:4173/watch/[test-stream-id]
   ```

2. **Meta Tag Validation**
   - Use [Facebook Sharing Debugger](https://developers.facebook.com/tools/debug/)
   - Use [Twitter Card Validator](https://cards-dev.twitter.com/validator)

3. **Structured Data**
   - Use [Google Rich Results Test](https://search.google.com/test/rich-results)
   - Test both organization and video schemas

### Landing Page
1. **Lighthouse Audit**
   - Open any page in Chrome
   - DevTools > Lighthouse > Run audit
   - Target: 90+ Performance, 100 SEO

2. **Meta Tag Validation**
   - Test all pages with Facebook/Twitter validators

3. **Sitemap Validation**
   - Visit `https://pews.church/sitemap.xml` (when deployed)
   - Submit to Google Search Console

---

## 📋 Deployment Checklist

### Before Deploying
- [ ] Update `robots.txt` domain if different from `pews.church`
- [ ] Update `sitemap.xml` domain if different
- [ ] Verify API_URL environment variable for preconnect
- [ ] Test server-side rendering is working (view page source, not inspect element)

### After Deploying
- [ ] Submit sitemap to Google Search Console
- [ ] Submit sitemap to Bing Webmaster Tools
- [ ] Test all meta tags with Facebook/Twitter validators
- [ ] Run Lighthouse audit on production URLs
- [ ] Verify gzip/brotli compression is enabled (check response headers)

### Web Server Configuration (Nginx/Caddy)
```nginx
# Example Nginx compression config
gzip on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript;
gzip_min_length 1000;
```

---

## 🎯 Additional SEO Opportunities

### Not Implemented (Future Work)
1. **Dynamic Sitemap Generation**
   - Generate sitemap entries for all active streams
   - Update automatically when new streams are created

2. **Blog/Content Pages**
   - Add `/blog` route with SEO-optimized articles
   - Target church management keywords

3. **Local SEO**
   - Add LocalBusiness schema for church instances
   - Implement location-based landing pages

4. **Social Proof**
   - Add customer testimonials with Review schema
   - Implement case studies with Article schema

5. **Performance Monitoring**
   - Set up Core Web Vitals tracking
   - Monitor Lighthouse scores over time

---

## 📊 Expected Impact

### SEO Improvements
- ✅ Proper indexing of public pages (watch, landing)
- ✅ Rich results in Google Search (video cards, organization info)
- ✅ Better social media sharing (proper previews on Facebook, Twitter, Discord, WhatsApp)
- ✅ Improved click-through rate from search results

### Performance Improvements
- ✅ Faster initial page load (preconnect, font-display)
- ✅ Reduced bundle size (tree shaking, code splitting)
- ✅ Better caching (vendor chunks separate)
- ✅ Improved perceived performance (lazy loading)

---

## 🔗 Resources

- [Google Search Central - SEO Starter Guide](https://developers.google.com/search/docs/beginner/seo-starter-guide)
- [Schema.org Documentation](https://schema.org/)
- [Open Graph Protocol](https://ogp.me/)
- [Web.dev Performance](https://web.dev/learn/#performance)
- [SvelteKit SEO Best Practices](https://kit.svelte.dev/docs/seo)
