# Pews Presenter — ProPresenter Killer

## The Opportunity

ProPresenter (Renewed Vision) dominates church presentation software at **$499 one-time** (single seat) or **$179/year** subscription. Most churches also pay separately for:
- Planning Center ($100-300/mo) for service planning
- ProPresenter ($499+) for lyrics/media projection
- A separate media asset manager

**Pews Presenter** would be the first church management platform with **built-in presentation/projection** — eliminating ProPresenter entirely. This is a category-defining feature that NO competitor (PCO, Breeze, Tithe.ly, Subsplash, ChurchTrac) offers.

## Core Features (MVP)

### 1. Lyric Projection
- Song database (already in Pews Song Library)
- Live lyric display with verse/chorus/bridge navigation
- Multiple font styles, sizes, backgrounds
- Lower-thirds mode for lyric overlay on video/image backgrounds

### 2. Sermon Support
- Scripture display (pull from Bible API)
- Sermon notes/outlines with click-to-advance
- Timer display for speaker

### 3. Media Playback
- Image backgrounds (church branding, seasonal themes)
- Video playback (announcements, worship backgrounds)
- Audio-only playback (backing tracks, walk-in music)

### 4. Multi-Screen / Multi-Output
- **Audience view**: What the congregation sees (lyrics, media)
- **Stage display**: What the worship team sees (next verse, chord charts, notes)
- **Operator view**: Control interface with preview + program + next slide

### 5. Live Service Integration
- Pull directly from Pews service plans (already built in Services module)
- Song order → automatic lyric queue
- Seamless handoff between worship, sermon, announcements

## Architecture Options

### Option A: Electron Desktop App (Recommended for MVP)
- **Pros**: Full screen control, multi-monitor support, local video playback, NDI output potential
- **Cons**: Requires install, platform-specific builds
- **Tech**: Electron + SvelteKit (reuse Pews frontend), WebRTC for stage display sync
- **Timeline**: 2-4 weeks for MVP with Sonnet agents

### Option B: Web-Based (Progressive)
- **Pros**: Zero install, works anywhere, remote control from phone
- **Cons**: Limited multi-monitor (Presentation API is spotty), video performance concerns
- **Tech**: SvelteKit + Presentation API + WebSocket sync
- **Timeline**: 1-2 weeks for basic, but limited by browser capabilities

### Option C: Hybrid (Best of Both)
- Electron for the operator station (multi-monitor, local media)
- Web app for stage display (any screen with a browser + URL)
- Phone/tablet remote control via WebSocket
- **This is the ProPresenter architecture** but integrated with Pews

### Recommendation: Option C (Hybrid)
- Electron operator app controls everything
- Stage displays are just browser windows (Raspberry Pi + TV = $50 stage display)
- Any phone becomes a remote clicker
- Cloud sync through Pews API (songs, service plans, media library)

## Competitive Destruction

| Feature | ProPresenter | Pews Presenter |
|---------|-------------|----------------|
| Lyric projection | ✅ | ✅ |
| Stage display | ✅ ($99 add-on) | ✅ (included) |
| Service planning | ❌ (need PCO) | ✅ (built-in) |
| Member management | ❌ | ✅ (built-in) |
| Giving/donations | ❌ | ✅ (built-in) |
| Check-ins | ❌ | ✅ (built-in) |
| Communication | ❌ | ✅ (built-in) |
| Price | $499 + $300/mo PCO | $100/mo total |

A church currently paying **$499 + $300/mo = ~$4,100/year** could switch to **$1,200/year** with MORE functionality. That's a 70% cost reduction.

## Market Size
- ~380,000 churches in the US
- ~60% use some form of presentation software
- ProPresenter has ~40% of that market (~91,000 churches)
- Even capturing 1% = 910 churches × $100/mo = $91K MRR

## Development Phases

### Phase 1: Song Display MVP (1-2 weeks)
- Electron app with full-screen output
- Pull songs from Pews Song Library API
- Basic text-on-background display
- Simple slide advance (keyboard/click)

### Phase 2: Multi-Screen (1 week)
- Stage display as separate browser window
- WebSocket sync between operator and stage
- Different layouts per screen

### Phase 3: Service Plan Integration (1 week)
- Pull service plan from Pews Services module
- Auto-queue songs, scriptures, media
- One-click "Go Live" from service plan

### Phase 4: Media & Polish (2 weeks)
- Video backgrounds
- Image slideshows
- Transitions and animations
- Bible verse lookup + display
- Timer/countdown

## Key Insight
The hardest part of ProPresenter is the real-time rendering engine. But we don't need GPU-accelerated 3D text or complex motion graphics. **90% of churches just need lyrics on a screen with a nice background.** That's a solved problem with HTML/CSS.

## Notes
- Consider CCLI SongSelect integration for lyrics import (API available)
- NDI output would be game-changer for streaming integration (Electron can do this via native modules)
- Mobile remote control is a killer feature that ProPresenter charges extra for
- Raspberry Pi as a $50 stage display computer is a massive selling point
