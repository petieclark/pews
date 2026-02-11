-- Seed default drip campaigns
-- This script should be run after a tenant is created
-- Replace YOUR_TENANT_ID with actual tenant ID

-- New Visitor Welcome Campaign
INSERT INTO drip_campaigns (id, tenant_id, name, trigger_event, is_active)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', 'YOUR_TENANT_ID', 'New Visitor Welcome', 'first_visit', true);

-- New Visitor Welcome Steps
INSERT INTO drip_steps (campaign_id, step_order, delay_days, action_type, subject, body) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 1, 0, 'email', 'Welcome to Our Church!', 
     'Hi {{first_name}},\n\nThank you for visiting us! We''re so glad you chose to worship with us. If you have any questions or would like to learn more about our church, please don''t hesitate to reach out.\n\nBlessings,\nThe Church Team'),
    
    ('550e8400-e29b-41d4-a716-446655440000', 2, 3, 'follow_up', 'Follow-up Call Reminder', 
     'Call {{first_name}} {{last_name}} to check in and see how their first visit was.'),
    
    ('550e8400-e29b-41d4-a716-446655440000', 3, 7, 'email', 'Explore Our Small Groups', 
     'Hi {{first_name}},\n\nWe hope you''ve had a great week! We wanted to let you know about our small groups - a great way to connect with others and grow in your faith.\n\nCheck out our available groups here: [GROUPS_LINK]\n\nBlessings,\nThe Church Team'),
    
    ('550e8400-e29b-41d4-a716-446655440000', 4, 14, 'email', 'Discover Ways to Serve', 
     'Hi {{first_name}},\n\nWe believe everyone has unique gifts and talents to share. We''d love to help you discover ways to serve and make a difference in our church and community.\n\nLearn more about serving opportunities: [SERVE_LINK]\n\nBlessings,\nThe Church Team');

-- New Member Onboarding Campaign
INSERT INTO drip_campaigns (id, tenant_id, name, trigger_event, is_active)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'YOUR_TENANT_ID', 'New Member Onboarding', 'new_member', true);

-- New Member Onboarding Steps
INSERT INTO drip_steps (campaign_id, step_order, delay_days, action_type, subject, body) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 1, 0, 'email', 'Welcome to the Family!', 
     'Hi {{first_name}},\n\nWelcome to the church family! We''re thrilled to have you officially join us. As a member, you''re now part of something bigger - a community dedicated to loving God and serving others.\n\nBlessings,\nThe Church Team'),
    
    ('550e8400-e29b-41d4-a716-446655440001', 2, 7, 'email', 'Getting Started: New Members Class', 
     'Hi {{first_name}},\n\nWe''d love to help you get connected! Our New Members Class is designed to help you learn more about our church, our beliefs, and how you can get involved.\n\nRegister for the next class: [CLASS_LINK]\n\nBlessings,\nThe Church Team'),
    
    ('550e8400-e29b-41d4-a716-446655440001', 3, 30, 'email', 'How Are You Doing?', 
     'Hi {{first_name}},\n\nIt''s been a month since you joined our church family! We wanted to check in and see how you''re doing. Have you been able to connect with others? Is there anything we can help you with?\n\nFeel free to reply to this email or give us a call.\n\nBlessings,\nThe Church Team');

-- Connection Card Follow-up Campaign
INSERT INTO drip_campaigns (id, tenant_id, name, trigger_event, is_active)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440002', 'YOUR_TENANT_ID', 'Connection Card Follow-up', 'connection_card', true);

-- Connection Card Follow-up Steps
INSERT INTO drip_steps (campaign_id, step_order, delay_days, action_type, subject, body) VALUES
    ('550e8400-e29b-41d4-a716-446655440002', 1, 0, 'email', 'Thanks for Connecting!', 
     'Hi {{first_name}},\n\nThank you for filling out a connection card! We''re excited to connect with you and help you feel at home in our church community.\n\nIf you indicated any prayer requests or interests, someone from our team will be in touch soon.\n\nBlessings,\nThe Church Team'),
    
    ('550e8400-e29b-41d4-a716-446655440002', 2, 1, 'follow_up', 'Review Connection Card', 
     'Review connection card submission from {{first_name}} {{last_name}}. Follow up on any prayer requests or expressed interests.'),
    
    ('550e8400-e29b-41d4-a716-446655440002', 3, 7, 'email', 'Stay Connected', 
     'Hi {{first_name}},\n\nWe hope you''ve had a wonderful week! We wanted to make sure you know about all the ways you can stay connected with our church:\n\n- Join us for Sunday services\n- Follow us on social media\n- Sign up for our weekly newsletter\n\nWe look forward to seeing you again soon!\n\nBlessings,\nThe Church Team');
