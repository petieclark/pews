# Pews Database Scripts

## seed-demo.sql

Comprehensive seed data script for the demo church instance (demo.pews.app).

### What it creates

- **Tenant:** Grace Community Church (slug: `demo-church`)
- **People:** 33 people across 8 realistic households
- **Tags:** 5 tags (Volunteer, Youth, Worship Team, Small Group Leader, New Member)
- **Groups:** 5 groups with members (small groups, youth, worship team, prayer team)
- **Songs:** 16 popular worship songs with CCLI numbers and metadata
- **Services:** 10 services (4 past Sunday, 2 past Wednesday, 4 upcoming)
- **Giving:** 3 funds with 48 donations spread across 3 months
- **Check-ins:** 2 stations with 28 check-ins across 4 Sundays
- **Communication:** 3 campaigns and 1 automated journey template
- **Streams:** 3 streams (2 ended, 1 scheduled)

### Usage

```bash
# Apply seed data to database
cat scripts/seed-demo.sql | docker compose exec -T postgres psql -U pews -d pews
```

### Features

- **Idempotent:** Safe to run multiple times without duplicating data
- **Realistic:** Uses common church names, realistic donation amounts, varied member statuses
- **Complete:** Covers all major modules (People, Groups, Services, Giving, Check-ins, Communication, Streaming)
- **Time-aware:** Past services marked as completed, future services as planning/confirmed

### Verification

After running the seed script, verify data was inserted:

```bash
# Check tenant exists
docker compose exec postgres psql -U pews -d pews -c "SELECT name, slug FROM tenants WHERE slug = 'demo-church';"

# View counts
docker compose exec postgres psql -U pews -d pews -c "
SELECT 
  (SELECT COUNT(*) FROM people WHERE tenant_id = '00000000-0000-0000-0000-000000000001') as people,
  (SELECT COUNT(*) FROM households WHERE tenant_id = '00000000-0000-0000-0000-000000000001') as households,
  (SELECT COUNT(*) FROM groups WHERE tenant_id = '00000000-0000-0000-0000-000000000001') as groups,
  (SELECT COUNT(*) FROM songs WHERE tenant_id = '00000000-0000-0000-0000-000000000001') as songs,
  (SELECT COUNT(*) FROM services WHERE tenant_id = '00000000-0000-0000-0000-000000000001') as services,
  (SELECT COUNT(*) FROM donations WHERE tenant_id = '00000000-0000-0000-0000-000000000001') as donations;
"
```

### Notes

- Uses fixed UUIDs for deterministic seeding (tenant ID: `00000000-0000-0000-0000-000000000001`)
- All data is tenant-scoped for isolation
- Donations span December 2024 through February 2025
- Services include past completed events and upcoming planned events
- Check-ins cover the last 4 Sundays with realistic attendance patterns
