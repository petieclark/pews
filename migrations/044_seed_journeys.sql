-- Seed: Welcome Visitor journey (connection_card trigger, 4 steps)
-- and New Member journey (new_member trigger, 3 steps)
-- These use inline config (subject/body) rather than templates.

-- Note: These are tenant-agnostic seeds. They'll be created per-tenant via the app.
-- This migration adds them to the seed_data table for the onboarding flow to pick up.

-- Create a seed_journeys table to store journey templates that get cloned per tenant on onboarding
CREATE TABLE IF NOT EXISTS seed_journey_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger_type VARCHAR(50) NOT NULL,
    steps JSONB NOT NULL DEFAULT '[]'
);

INSERT INTO seed_journey_templates (name, description, trigger_type, steps) VALUES
(
    'Welcome Visitor',
    'Automated welcome sequence for new visitors who fill out a connection card',
    'connection_card',
    '[
        {
            "position": 1,
            "step_type": "send_email",
            "delay_days": 0,
            "delay_hours": 0,
            "config": {
                "subject": "Welcome to {{church_name}}! We''re glad you visited.",
                "body": "<p>Hi {{first_name}},</p><p>Thank you so much for visiting {{church_name}}! We are thrilled you joined us and hope you felt welcome.</p><p>If you have any questions about our church or want to learn more, don''t hesitate to reach out. We''d love to help you get connected!</p><p>See you soon,<br>The {{church_name}} Team</p>"
            }
        },
        {
            "position": 2,
            "step_type": "send_email",
            "delay_days": 3,
            "delay_hours": 0,
            "config": {
                "subject": "Here are some ways to get connected at {{church_name}}",
                "body": "<p>Hi {{first_name}},</p><p>We wanted to share a few ways you can get more connected at {{church_name}}:</p><ul><li><strong>Small Groups</strong> — Meet others and grow in faith together</li><li><strong>Upcoming Events</strong> — Check out what''s happening this month</li><li><strong>Volunteer Teams</strong> — Use your gifts to serve others</li></ul><p>Getting plugged in is the best way to make {{church_name}} feel like home. Let us know how we can help!</p><p>Blessings,<br>The {{church_name}} Team</p>"
            }
        },
        {
            "position": 3,
            "step_type": "send_email",
            "delay_days": 7,
            "delay_hours": 0,
            "config": {
                "subject": "We''d love to see you again this Sunday!",
                "body": "<p>Hi {{first_name}},</p><p>Just a quick note to say we''d love to see you again this weekend at {{church_name}}!</p><p>Whether it''s your second visit or you''re still thinking about it, know that there''s always a seat saved for you.</p><p>Hope to see you soon!<br>The {{church_name}} Team</p>"
            }
        },
        {
            "position": 4,
            "step_type": "send_email",
            "delay_days": 14,
            "delay_hours": 0,
            "config": {
                "subject": "Have you thought about joining a small group?",
                "body": "<p>Hi {{first_name}},</p><p>One of the best ways to build lasting friendships and grow in your faith is through a small group.</p><p>Our groups meet throughout the week at various times and locations. There''s a group for everyone — whether you''re into Bible study, serving the community, or just hanging out with great people.</p><p>Interested? Reply to this email and we''ll help you find the perfect fit!</p><p>With love,<br>The {{church_name}} Team</p>"
            }
        }
    ]'::jsonb
),
(
    'New Member Welcome',
    'Onboarding sequence for people who become official members',
    'new_member',
    '[
        {
            "position": 1,
            "step_type": "send_email",
            "delay_days": 0,
            "delay_hours": 0,
            "config": {
                "subject": "Welcome to the {{church_name}} family!",
                "body": "<p>Hi {{first_name}},</p><p>Congratulations on becoming an official member of {{church_name}}! We are so excited to have you as part of our church family.</p><p>As a member, you''re not just attending — you''re part of something bigger. We can''t wait to see how God uses you here.</p><p>Welcome home!<br>The {{church_name}} Team</p>"
            }
        },
        {
            "position": 2,
            "step_type": "send_email",
            "delay_days": 7,
            "delay_hours": 0,
            "config": {
                "subject": "Your next steps as a {{church_name}} member",
                "body": "<p>Hi {{first_name}},</p><p>Now that you''re a member, here are some great next steps:</p><ul><li><strong>Join a small group</strong> — Build deep relationships</li><li><strong>Discover your gifts</strong> — Take our spiritual gifts assessment</li><li><strong>Serve on a team</strong> — There are so many ways to get involved</li></ul><p>If you need help with any of these, just reply to this email!</p><p>Blessings,<br>The {{church_name}} Team</p>"
            }
        },
        {
            "position": 3,
            "step_type": "send_email",
            "delay_days": 14,
            "delay_hours": 0,
            "config": {
                "subject": "How are you settling in, {{first_name}}?",
                "body": "<p>Hi {{first_name}},</p><p>It''s been a couple of weeks since you joined {{church_name}}, and we just wanted to check in.</p><p>Do you have any questions? Need help finding a group or team? We''re here for you — just hit reply and let us know.</p><p>We''re glad you''re here!<br>The {{church_name}} Team</p>"
            }
        }
    ]'::jsonb
);
