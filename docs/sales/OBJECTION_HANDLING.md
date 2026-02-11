# Pews Objection Handling Guide

## Philosophy

Objections aren't rejection—they're **requests for more information**. Your job isn't to "overcome" objections; it's to **understand the concern and provide clarity**.

### Framework: LAER
1. **Listen**: Let them finish. Don't interrupt.
2. **Acknowledge**: "I understand why that's a concern."
3. **Explore**: Ask clarifying questions. "Tell me more about..."
4. **Respond**: Address the root concern, not just the surface objection.

---

## Common Objections & Responses

### 1. "We already use Planning Center. It's been fine."

**Root Concern**: Change is risky. Why fix what isn't broken?

**Response**:

"I totally get it—Planning Center is the industry standard for a reason. They've been around forever, and if it's working, why change?

Here's what I'd ask: **How much are you paying per month?** Most churches we talk to are spending $200-500 on Planning Center alone—and that's just for people, check-in, and groups. You're likely paying separately for giving (Tithe.ly or Pushpay) and email (Mailchimp), right?

Pews does everything Planning Center does—**plus** giving, SMS, email campaigns, and drip automation—for $100/month flat.

So the question isn't 'Is Planning Center fine?'—it's **'Is it worth paying $200-400 extra per month for the same features?'**

And here's the best part: **We'll migrate your data for free.** You don't lose anything. If Pews doesn't feel right after a month, you can go back. But most churches who switch never look back—because they're saving $2,400-4,800 per year."

**Follow-Up Question**:  
"Would you be open to seeing a side-by-side comparison of what you're paying now vs. what you'd pay with Pews?"

---

### 2. "Is it secure? We handle sensitive data."

**Root Concern**: Data breaches, privacy violations, liability.

**Response**:

"Absolutely critical question—I'd be concerned too.

Here's how we handle security:

1. **Row-Level Security (RLS)**: Every user only sees data they're authorized for. A volunteer checking in kids can't see giving records. A small group leader can't access admin dashboards. This is enforced at the database level, not just in the UI—so even if someone tries to hack the API, they hit a wall.

2. **Encryption**:
   - **Data at rest**: AES-256 encryption (same standard as banks)
   - **Data in transit**: TLS 1.3 (HTTPS everywhere)

3. **Daily automated backups**: Stored in geographically redundant locations. If something goes wrong, we can restore to any point in the last 30 days.

4. **GDPR-compliant**: Even though most US churches don't *have* to comply with GDPR, we do anyway—because it's the gold standard for data privacy.

5. **No third-party data sharing**: Your data is yours. We don't sell it, we don't share it, we don't train AI models on it.

6. **Audit logs**: Every action is logged (who accessed what, when). If you ever need to investigate something, you have a full trail.

We can also send you our **security whitepaper** if your board or IT committee needs more detail. And if you have specific compliance requirements (HIPAA, PCI, etc.), let's talk through those."

**Follow-Up Question**:  
"What specific security concerns does your board or leadership have? Happy to address those directly."

---

### 3. "Can we try it first before committing?"

**Root Concern**: Fear of buyer's remorse. Need to test before investing.

**Response**:

"100%. In fact, I insist on it.

Here's what I can do:

1. **Spin up a demo instance** with your church's name and sample data. You can log in, click around, invite your team, test check-in on your iPads—whatever you need to see.

2. **No credit card required**. No commitment. If you hate it, just say so and we part as friends.

3. **If you like it**, we move forward with data migration. We'll pull your real data from Planning Center (or wherever you're at now), import it into Pews, and let you run it in parallel for a week or two. You can compare side-by-side.

4. **Go live when you're ready**—not when we pressure you.

Most churches know within a week whether Pews is a fit. Some take a month. That's fine. We're not going anywhere.

Sound good? I can have your demo instance ready by end of day today."

**Follow-Up**:  
"Who else on your team should have access to the demo? Pastors, admin staff, volunteers?"

---

### 4. "What about support? What if something breaks on Sunday morning?"

**Root Concern**: Downtime during critical moments (Sunday service).

**Response**:

"Great question—because if check-in goes down at 9:45 AM on Sunday, you're in a tough spot.

Here's our approach:

1. **Email and chat support included** in your $100/mo. No tiered support plans, no extra fees. Typical response time: under 2 hours (usually faster).

2. **For your first Sunday**, we can have someone on **standby via Zoom or phone**. You go live, we're watching. If anything hiccups, we're there in real time to troubleshoot.

3. **Uptime**: Pews is hosted on **Vercel + Supabase**, which have 99.9%+ uptime SLAs. These are enterprise-grade platforms used by companies like OpenAI, GitHub, and Notion. If they go down, half the internet goes down—and they rarely do.

4. **Offline fallback**: If the internet goes out at your church (it happens), you can use our **offline check-in mode**. It syncs once you're back online.

5. **Knowledge base & video tutorials**: Most common questions are answered in our help center. You can search 24/7.

And here's the thing: Planning Center and Breeze go down too. The difference is, **we're small enough to care deeply about every church**. You're not ticket #47293 in a queue. You're a real person we want to help.

That said—**we can't promise zero issues**. Software is software. But we *can* promise we'll bust our butts to fix it fast."

**Follow-Up**:  
"What time does your service start on Sundays? We'll make sure someone's available during that window for your first few weeks."

---

### 5. "We're not very technical. Is this too complicated?"

**Root Concern**: Learning curve, staff frustration, training burden.

**Response**:

"I love that you're thinking about this—because if your staff hates the tool, it doesn't matter how good it is.

Here's our design philosophy: **If you can use Gmail and Facebook, you can use Pews.**

No complex setup. No IT department needed. It's just a web browser. Log in, click around, it makes sense.

But don't take my word for it—**let me show you**. [Offer demo or trial.]

As for training:

1. **Onboarding call**: 2-hour Zoom session where we walk your team through setup. People management, check-in, giving, groups—everything.

2. **Video tutorials**: Short (5-10 min) videos for each feature. Your volunteers can watch on their own time.

3. **We do the heavy lifting**: Data migration, setup, configuration—**we handle it**. You just show up and start using it.

4. **Ongoing support**: Questions pop up? Email or chat us. We're here.

And here's the kicker: **Planning Center is way more complicated.** They have seven separate products, each with its own login, its own interface. Pews is one tool, one place. Simpler, not harder."

**Follow-Up**:  
"Who on your team would be the main admin? Let's make sure they're comfortable with it before you commit."

---

### 6. "What if we outgrow it?"

**Root Concern**: Scalability. Will Pews handle growth?

**Response**:

"Love that you're thinking about growth—that's a healthy problem to have.

Pews scales from **50 members to 5,000+**. Same interface, same price.

We have churches using Pews with:
- 50 members (small church plant)
- 500 members (mid-size suburban church)
- 2,000+ members (large multi-site church)

The system doesn't slow down. The UI doesn't change. You don't hit some arbitrary limit where we force you onto an 'enterprise plan.'

Now, if you grow to **10,000 members**—amazing! At that point, you might need custom integrations or dedicated infrastructure. We'll have that conversation when you get there. But for 99% of churches, Pews will handle everything you need for the next 10+ years.

And if you ever *do* outgrow us (which would be incredible), **we'll export your data cleanly**. No lock-in. You own your data, and you can take it wherever you need to go."

**Follow-Up**:  
"How fast are you growing? That helps me understand your scaling timeline."

---

### 7. "What's the catch? This seems too cheap."

**Root Concern**: Suspicion. If it's too good to be true, it probably is.

**Response**:

"Fair question. I'd be skeptical too.

Here's the truth: **There's no catch.**

We're not a venture-backed startup burning money to acquire customers and then jacking up prices later. We're a **small, sustainable business** building tools churches actually need—without the bloat and overhead of legacy players.

Why are we cheaper?

1. **Modern tech stack**: We built Pews in 2024 using serverless infrastructure (Supabase, Vercel). Our hosting costs are 10x lower than Planning Center's 2006-era servers. We pass that savings to you.

2. **No sales team**: Planning Center has dozens of salespeople, account managers, and enterprise reps. We don't. You're talking directly to the people who build the product.

3. **No enterprise bloat**: We don't spend millions on Super Bowl ads or sponsor conferences. We focus on making a great product and letting word-of-mouth do the work.

4. **One product, not seven**: Planning Center maintains seven separate products (People, Check-Ins, Giving, etc.). We built *one* unified platform. Less overhead, less complexity.

The result? **We can charge $100/month and still run a healthy, profitable business.**

So the catch is... there isn't one. We're just building what Planning Center *would* be if they started today."

**Follow-Up**:  
"What would make you feel confident this is legit? Happy to share customer references, financials, whatever you need."

---

### 8. "We're locked into a contract with [current tool]."

**Root Concern**: Sunk cost. Already paid for the year.

**Response**:

"I get it—sunk costs hurt. Let's do some math.

Let's say you prepaid Breeze for the year: **$720**. That money's gone, whether you use it or not.

Now you have two choices:

**Option 1: Wait it out.**  
Finish the year with Breeze. Then switch. You save $0 this year, start saving next year.

**Option 2: Switch now.**  
Yes, you lose the remaining months of Breeze (painful). But you start saving immediately.

Let's say you have 6 months left on Breeze. You lose $360. Ouch.

But over the next **5 years**, switching to Pews saves you **$X,XXX** compared to staying on Breeze.

So the question is: Do you want to delay those savings by 6 months to avoid losing $360 today?

Most churches say, 'You're right—let's rip the band-aid off and switch now.' A few say, 'We'll switch when the contract ends.' Both are valid. But waiting doesn't save you money—it just delays the savings.

And here's the other thing: **Your current tool isn't free just because you prepaid.** You're paying in staff time, frustration, and inefficiency. If Pews saves your admin 5 hours a week, that's 20 hours a month—what's that worth?"

**Follow-Up**:  
"When does your contract expire? Let's plan the transition so you're ready to switch the day it ends."

---

### 9. "What about integrations? We use [tool X]."

**Root Concern**: Workflow disruption. Losing existing integrations.

**Response**:

"Great question. What tool are you using?"

*[Listen. Then respond based on their answer.]*

**Common scenarios**:

**Accounting (QuickBooks, Xero)**:  
"Pews can export giving data in formats compatible with QuickBooks and Xero. You can import transactions monthly or weekly—takes about 5 minutes. We're also building native integrations for both (coming Q2 2026)."

**Email marketing (Mailchimp, Constant Contact)**:  
"You won't need them anymore—Pews has built-in email and SMS. But if you want to keep using Mailchimp for newsletters, you can export contact lists from Pews and import to Mailchimp. Or use Zapier to sync automatically."

**Worship planning (Planning Center Services)**:  
"If you're just using PCO Services for worship team scheduling, Pews has volunteer scheduling built-in. If you need chord charts, lyric projection, and setlists, you'd keep PCO Services (it's $20/mo standalone). Most churches are fine with just Pews, but we get it if you're attached to PCO Services."

**Website (WordPress, Squarespace)**:  
"Pews gives you embeddable widgets for your website—online giving, event registration, group signups, etc. Just copy/paste a snippet of code and it works. No need to rebuild your site."

**Zapier**:  
"If you have custom workflows (e.g., 'When someone fills out a Google Form, add them to Pews'), Zapier integration is on our roadmap for Q2 2026. In the meantime, you can use our API or CSV imports."

**Bottom line**: Most integrations aren't deal-breakers. But if you have a must-have integration, let's talk through it. We might already support it, or we can prioritize building it."

**Follow-Up**:  
"What integrations are absolute must-haves for you? Let's make sure we can support your workflow."

---

### 10. "We've been burned by software before. How do we know this won't be the same?"

**Root Concern**: Past trauma. Trust issues.

**Response**:

"I'm really sorry that happened. Getting burned by software—especially something your whole church depends on—is brutal.

Can I ask what went wrong?"

*[Listen. Let them vent. Then respond.]*

"That's awful. Here's what we do differently:

1. **Transparent pricing**: No surprise fees, no bait-and-switch. $100/mo, period.

2. **No long-term contracts**: You can cancel anytime. We're not locking you in—we're earning your business every month.

3. **We own our mistakes**: If something breaks, we'll say so. We'll fix it. We'll tell you what we're doing to prevent it next time. No corporate runaround.

4. **You own your data**: If Pews doesn't work out, we'll export your data cleanly. No hostage-taking, no 'contact sales to unlock your data' nonsense.

5. **Direct access to the team**: You're not talking to a call center in another country. You're talking to the people who actually build Pews.

That said—**I can't promise we'll never make mistakes.** Software is hard. But I can promise we'll care deeply about fixing it fast and making it right.

Would a trial help? You can kick the tires with zero commitment and see if we're different from the tool that burned you."

**Follow-Up**:  
"What would make you feel confident enough to try Pews? Customer references? Demo? Something else?"

---

### 11. "Our pastor/board needs to approve this. How do I sell them?"

**Root Concern**: Internal buy-in. Decision-maker isn't in the room.

**Response**:

"Totally get it—this isn't a unilateral decision. Let's make it easy for you to sell them.

Here's what I can provide:

1. **One-page summary** (print or PDF): Problem → Solution → Pricing → ROI. Something you can hand to your pastor or board at the next meeting.

2. **Pricing comparison spreadsheet**: Shows exactly how much you're spending now vs. what Pews costs. Hard numbers, easy sell.

3. **Demo recording**: I can record a 10-minute walkthrough of Pews so they can watch on their own time.

4. **Live demo for decision-makers**: If they want to see it themselves, I'm happy to hop on a call with your pastor, board chair, or whoever needs to say yes.

5. **Customer references**: I can connect you with another church (similar size, similar situation) who switched to Pews. Let them hear from someone who's been there.

What's the approval process look like? Who needs to say yes, and when's the next meeting?"

**Follow-Up**:  
"What objections do you think your pastor/board will have? Let's address those upfront so you're prepared."

---

### 12. "We're launching in 6 months. Can we wait to decide?"

**Root Concern**: Timing. Not ready yet.

**Response**:

"Totally fair. But let me ask: **What are you using in the meantime?**

If you're using spreadsheets and paper forms—fine, you can survive 6 months. But if you're paying for Breeze, PCO, or another tool right now, you're spending $XX/month that you could save by switching to Pews earlier.

Let's do the math:

- **Wait 6 months**: Pay $XX/mo for current tool = $XXX total
- **Switch now**: Pay $100/mo for Pews = $600 total
- **Savings**: $XXX - $600 = $XXX saved

Plus, switching now gives you **6 months to learn the system** before your big launch. You'll be way more confident on day one.

That said—if you're truly not ready, I get it. Let's stay in touch. I'll check in with you in 3 months and see where you're at. Sound good?"

**Follow-Up**:  
"What's happening in 6 months? Church launch? New building? Let me know how we can help when the time comes."

---

### 13. "What if Pews goes out of business?"

**Root Concern**: Business continuity. Longevity.

**Response**:

"Super valid concern—especially with a newer company.

Here's the reality:

1. **Pews is profitable.** We're not a VC-funded startup burning cash. We're a small, sustainable business. We don't need 10,000 customers to survive—we need 200. And we're already past that.

2. **You own your data.** If Pews ever shuts down (which we don't plan to), we'll give you 90 days' notice and export all your data in standard formats (CSV, JSON, SQL). You can import it into any other tool.

3. **Open-source contingency.** We're exploring open-sourcing Pews if we ever wind down. That means anyone could host it themselves. But again—not planning on it.

4. **Compare to the alternatives:**
   - Planning Center: Been around since 2006. Solid. But also expensive and bloated.
   - Breeze: Small team, like us. But they've been around since 2012.
   - Pews: New, but built on rock-solid infrastructure (Supabase, Vercel). If *we* go away, those platforms aren't going anywhere.

Here's my ask: **Judge us on the product, not on how long we've been around.** Try Pews for a month. If it works, great. If we shut down (again, not planning to), you'll have 90 days to migrate. That's way more notice than most software companies give."

**Follow-Up**:  
"What would give you confidence in our longevity? Financial transparency? Customer growth numbers? Something else?"

---

### 14. "We're worried about data migration. What if it goes wrong?"

**Root Concern**: Data loss or corruption during transition.

**Response**:

"Critical concern—getting data migration right is make-or-break.

Here's how we handle it:

1. **We do the migration, not you.** You don't touch anything. We export from your current system (Planning Center, Breeze, etc.), map the fields, and import into Pews.

2. **Test migration first.** We do a test import into a staging environment. You review it, catch any issues, we fix them. Then we do the final migration.

3. **Run parallel for a week.** Keep using your old system while Pews runs alongside. Double-check everything. Once you're confident, switch over.

4. **We've done this before.** We've migrated churches from PCO, Breeze, Elvanto, CCB, FellowshipOne—you name it. We know the gotchas.

5. **Backups.** Your old data stays in your old system (you don't delete anything until you're 100% sure). And Pews keeps daily backups, so even if something goes sideways, we can roll back.

That said—**no migration is perfect.** You might find a few phone numbers formatted weird or a household that didn't merge correctly. But we'll catch 99% of it upfront, and the 1% we fix together.

I won't sugarcoat it: Data migration is the hardest part of switching tools. But we've gotten pretty good at it."

**Follow-Up**:  
"What system are you migrating from? Let me walk you through exactly what that process looks like."

---

### 15. "We need multi-site support. Can Pews handle that?"

**Root Concern**: Complexity of managing multiple campuses.

**Response**:

"Yes—Pews supports multi-site churches.

Here's how it works:

1. **One database, multiple campuses.** Each campus has its own check-in stations, giving reports, attendance tracking, etc. But your admin dashboard shows **consolidated data across all sites**.

2. **Campus-specific permissions.** Your campus pastor in [Location B] can see their campus's data but not [Location A]'s. Your executive pastor can see everything.

3. **Unified reporting.** Want to see total giving across all campuses? One report. Want to compare attendance trends between sites? One dashboard.

4. **Separate or shared groups.** You decide: Are small groups campus-specific, or church-wide?

5. **Same $100/mo price.** Multi-site doesn't cost extra (unlike Planning Center, where you pay per location).

That said—if you're running **10+ campuses** with complex org structures, we should have a deeper conversation. Pews handles 2-5 campuses beautifully. Beyond that, we might need to customize."

**Follow-Up**:  
"How many campuses do you have? What's your structure like (autonomous vs. centralized)?"

---

### 16. "We already paid for [X tool] for the year. Can't we wait?"

*See Objection #8 above (locked-in contract).*

---

### 17. "What if our internet goes down on Sunday?"

**Root Concern**: Single point of failure (internet dependency).

**Response**:

"Great question—internet outages are rare, but they happen.

Here's our offline strategy:

1. **Offline check-in mode** (coming Q1 2026): You can check kids in even without internet. Data syncs when you're back online.

2. **Mobile hotspot backup**: Most churches keep a mobile hotspot (phone or dedicated device) as backup. If your main internet dies, switch to hotspot for 20 minutes while you troubleshoot.

3. **Paper fallback**: Old-school pen-and-paper roster. Not ideal, but it works. You manually enter data into Pews later.

That said—**every cloud tool has this limitation** (Planning Center, Breeze, Tithe.ly). If you want true offline functionality, you'd need an on-premise system hosted on a local server—but then *you* have to maintain it, back it up, and fix it when it breaks. Most churches prefer the trade-off of cloud reliability (99.9% uptime) vs. the hassle of self-hosting.

Bottom line: Internet outages are inconvenient, but rare. And when they happen, we have workarounds."

**Follow-Up**:  
"How often does your internet go down? If it's frequent, let's talk about backup options."

---

### 18. "We're a small church (under 50 people). Is Pews overkill?"

**Root Concern**: Paying for features they don't need.

**Response**:

"I appreciate the concern—you don't want to overpay for bloat.

Here's the thing: **Pews scales down just as well as it scales up.**

Even at 50 people, you probably need:
- A member directory (so you're not hunting through spreadsheets)
- Attendance tracking (for trends and follow-up)
- Online giving (because nobody carries cash anymore)
- Email/SMS (to communicate with your people)
- Kids check-in (if you have families)

Those aren't 'enterprise features'—they're basics. And Pews gives you all of them for $100/mo.

Compare to alternatives:
- **Spreadsheets**: Free, but manual and error-prone
- **Breeze**: $72/mo—not that much cheaper, and less capable
- **Planning Center**: $100+ once you add modules

So Pews isn't overkill—**it's the right tool at the right price**, whether you have 50 members or 500.

And here's the best part: When you grow to 100, 200, 500 people, **the system grows with you**. You're not re-learning a new tool every time you hit a size threshold."

**Follow-Up**:  
"What are you using now to manage your 50 people? Let me show you how much easier Pews makes it."

---

### 19. "We're worried about vendor lock-in."

**Root Concern**: Can't leave if we want to.

**Response**:

"Totally legitimate concern. Nobody wants to feel trapped.

Here's our stance on lock-in:

1. **No long-term contracts.** Month-to-month. Cancel anytime.

2. **You own your data.** Not us. You. We're just hosting it for you.

3. **Export anytime.** Click a button, download your entire database (CSV, JSON, SQL—your choice). No hoops, no 'contact sales,' no fees.

4. **No proprietary formats.** Your data is stored in standard formats that any other tool can import.

5. **API access** (if you're technical): You can pull data programmatically anytime you want.

The only 'lock-in' is that **Pews works so well you don't want to leave.** But if you do, we'll make it easy.

Compare that to Planning Center, where exporting data is a multi-step process and some data doesn't export cleanly. Or FellowshipOne, where churches have told us horror stories about trying to leave.

**We don't want prisoners. We want happy customers.**"

**Follow-Up**:  
"Have you had bad experiences with vendor lock-in before? Let me know what would make you feel safe."

---

### 20. "This sounds great, but we need to think about it."

**Root Concern**: Stalling. Uncertain. Need time.

**Response**:

"Totally fair—this is a big decision. Take the time you need.

Before you go, let me ask: **What's holding you back?**

Is it:
- Price? (Let's talk budget.)
- Features? (What's missing that you need?)
- Timing? (Are you not ready yet?)
- Trust? (What would make you confident in Pews?)
- Internal buy-in? (Do you need to convince someone else?)

I'm not trying to pressure you—I genuinely want to understand so I can help. If Pews isn't the right fit, that's okay. But if it *is* the right fit and something's just unclear, let's clear it up now.

What's your timeline for deciding? Can I check in with you next week?"

**Follow-Up**:  
"What would make this a 'hell yes' instead of a 'we need to think about it'?"

---

## When to Walk Away

Not every church is a good fit for Pews. Here are signs to **disqualify gracefully**:

1. **They want on-premise hosting.** (We're cloud-only.)
2. **They need complex multi-site with 10+ campuses.** (We're not there yet.)
3. **They need denominational reporting integrations.** (Coming, but not built yet.)
4. **They're shopping purely on price** and will churn the moment someone's $5/mo cheaper.
5. **They're unkind or abusive** to you. (Life's too short.)

**How to disqualify**:  
"Based on what you've told me, I don't think Pews is the right fit for you right now. Here's what I'd recommend instead: [suggest alternative]. If your needs change down the road, feel free to reach out."

---

## Final Tip: The Best Objection Handler is Confidence

If you believe Pews is the best tool for the job—and you genuinely want to help churches save money and run better—that comes through.

Don't be desperate. Don't be pushy. Be confident, helpful, and honest.

**You're not selling. You're solving a problem.**
