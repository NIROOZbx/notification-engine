-- System Workspace
INSERT INTO workspaces (id, name, slug, plan_id)
VALUES (
        '00000000-0000-0000-0000-000000000001',
        '__system__',
        '__system__',
        (
            SELECT id
            FROM plans
            WHERE name = 'Enterprise'
            LIMIT 1
        )
    ) ON CONFLICT (id) DO NOTHING;
-- System Environment
INSERT INTO environments (id, workspace_id, name)
VALUES (
        '00000000-0000-0000-0000-000000000002',
        '00000000-0000-0000-0000-000000000001',
        'production'
    ) ON CONFLICT (id) DO NOTHING;
-- System Layout (Premium Design)
INSERT INTO layouts (workspace_id, name, html, is_default)
VALUES (
        '00000000-0000-0000-0000-000000000001',
        'System Layout',
        '<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background-color: #fafafa; font-family: -apple-system, BlinkMacSystemFont, ''Segoe UI'', Roboto, Helvetica, Arial, sans-serif;">
    <table width="100%" border="0" cellspacing="0" cellpadding="0" style="background-color: #fafafa;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table width="600" border="0" cellspacing="0" cellpadding="0" style="background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);">
                    <!-- Header -->
                    <tr>
                        <td align="center" style="padding: 40px 40px 20px 40px;">
                            <div style="font-size: 24px; font-weight: 800; letter-spacing: -0.025em; color: #000000;">NIROOZ</div>
                        </td>
                    </tr>
                    <!-- Content -->
                    <tr>
                        <td style="padding: 0 40px 40px 40px;">
                            {{content}}
                        </td>
                    </tr>
                    <!-- Footer -->
                    <tr>
                        <td align="center" style="padding: 32px 40px; background-color: #f9fafb; border-top: 1px solid #f3f4f6;">
                            <p style="margin: 0; font-size: 12px; color: #6b7280; line-height: 1.5;">
                                &copy; 2026 NIROOZ Inc. All rights reserved.<br>
                                High-performance notification infrastructure for modern teams.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>',
        true
    ) ON CONFLICT DO NOTHING;
-- Verification Template
INSERT INTO templates (
        workspace_id,
        environment_id,
        name,
        description,
        event_type,
        status
    )
VALUES (
        '00000000-0000-0000-0000-000000000001',
        '00000000-0000-0000-0000-000000000002',
        'Email Verification',
        'System template for user email verification',
        'user.verification',
        'live'
    ) ON CONFLICT (workspace_id, environment_id, event_type) DO NOTHING;
-- Verification Email Channel
INSERT INTO template_channels (template_id, channel, content)
VALUES (
        (
            SELECT id
            FROM templates
            WHERE event_type = 'user.verification'
                AND workspace_id = '00000000-0000-0000-0000-000000000001'
            LIMIT 1
        ), 'email', '{"subject": "Verify your email address", "body": "<div style=\"text-align: center;\"><h1 style=\"margin: 0 0 16px 0; font-size: 24px; font-weight: 700; color: #111827;\">Verify your email</h1><p style=\"margin: 0 0 32px 0; font-size: 16px; line-height: 1.6; color: #4b5563;\">Welcome! Please click the button below to verify your email address and complete your registration.</p><a href=\"{{verification_url}}\" style=\"display: inline-block; padding: 14px 32px; background-color: #000000; color: #ffffff; font-size: 14px; font-weight: 600; text-decoration: none; border-radius: 8px;\">Verify Email Address</a><p style=\"margin: 32px 0 0 0; font-size: 14px; color: #9ca3af;\">If you didn''t request this, you can safely ignore this email.</p></div>"}'
    ) ON CONFLICT (template_id, channel) DO NOTHING;
-- 80% Usage Warning Template
INSERT INTO templates (
        workspace_id,
        environment_id,
        name,
        description,
        event_type,
        status
    )
VALUES (
        '00000000-0000-0000-0000-000000000001',
        '00000000-0000-0000-0000-000000000002',
        '80% Usage Warning',
        'System alert at 80% limit',
        'subscription_limit_reached_80',
        'live'
    ) ON CONFLICT (workspace_id, environment_id, event_type) DO NOTHING;
INSERT INTO template_channels (template_id, channel, content)
VALUES (
        (
            SELECT id
            FROM templates
            WHERE event_type = 'subscription_limit_reached_80'
                AND workspace_id = '00000000-0000-0000-0000-000000000001'
            LIMIT 1
        ), 'email', '{"subject": "Action Required: 80% of notification limit reached", "body": "<div><h1 style=\"margin: 0 0 16px 0; font-size: 24px; font-weight: 700; color: #111827;\">Usage Warning</h1><p style=\"margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #4b5563;\">Your workspace has used <strong>{{current}}</strong> notifications on the <strong>{{channel}}</strong> channel. You are approaching your monthly limit.</p><div style=\"padding: 16px; background-color: #fffbeb; border-left: 4px solid #f59e0b; border-radius: 4px; margin-bottom: 32px;\"><p style=\"margin: 0; font-size: 14px; color: #92400e;\">Your usage resets on <strong>{{reset_at}}</strong>. Consider upgrading to avoid service interruption.</p></div><a href=\"{{upgrade_url}}\" style=\"display: inline-block; padding: 14px 32px; background-color: #000000; color: #ffffff; font-size: 14px; font-weight: 600; text-decoration: none; border-radius: 8px;\">View Plans & Upgrade</a></div>"}'
    ) ON CONFLICT (template_id, channel) DO NOTHING;
-- 100% Limit Reached Template
INSERT INTO templates (
        workspace_id,
        environment_id,
        name,
        description,
        event_type,
        status
    )
VALUES (
        '00000000-0000-0000-0000-000000000001',
        '00000000-0000-0000-0000-000000000002',
        'Limit Reached',
        'System alert at 100% limit',
        'subscription_limit_reached_100',
        'live'
    ) ON CONFLICT (workspace_id, environment_id, event_type) DO NOTHING;
INSERT INTO template_channels (template_id, channel, content)
VALUES (
        (
            SELECT id
            FROM templates
            WHERE event_type = 'subscription_limit_reached_100'
                AND workspace_id = '00000000-0000-0000-0000-000000000001'
            LIMIT 1
        ), 'email', '{"subject": "Urgent: Notification limit reached", "body": "<div><h1 style=\"margin: 0 0 16px 0; font-size: 24px; font-weight: 700; color: #111827;\">Limit Reached</h1><p style=\"margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #4b5563;\">Your workspace has reached its monthly limit on the <strong>{{channel}}</strong> channel with <strong>{{current}}</strong> notifications sent.</p><div style=\"padding: 16px; background-color: #fef2f2; border-left: 4px solid #ef4444; border-radius: 4px; margin-bottom: 32px;\"><p style=\"margin: 0; font-size: 14px; color: #b91c1c;\">New notifications will be blocked until your usage resets on <strong>{{reset_at}}</strong>. Upgrade now to resume service instantly.</p></div><a href=\"{{upgrade_url}}\" style=\"display: inline-block; padding: 14px 32px; background-color: #000000; color: #ffffff; font-size: 14px; font-weight: 600; text-decoration: none; border-radius: 8px;\">Upgrade Plan Now</a></div>"}'
    ) ON CONFLICT (template_id, channel) DO NOTHING;
-- Expiry Reminder Template
INSERT INTO templates (
        workspace_id,
        environment_id,
        name,
        description,
        event_type,
        status
    )
VALUES (
        '00000000-0000-0000-0000-000000000001',
        '00000000-0000-0000-0000-000000000002',
        'Expiry Reminder',
        'System alert before plan expiry',
        'subscription_expiry_reminder',
        'live'
    ) ON CONFLICT (workspace_id, environment_id, event_type) DO NOTHING;
INSERT INTO template_channels (template_id, channel, content)
VALUES (
        (
            SELECT id
            FROM templates
            WHERE event_type = 'subscription_expiry_reminder'
                AND workspace_id = '00000000-0000-0000-0000-000000000001'
            LIMIT 1
        ), 'email', '{"subject": "Your subscription is expiring soon", "body": "<div><h1 style=\"margin: 0 0 16px 0; font-size: 24px; font-weight: 700; color: #111827;\">Subscription Expiring</h1><p style=\"margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #4b5563;\">Your subscription for workspace is scheduled to expire on <strong>{{expiry_date}}</strong>.</p><p style=\"margin: 0 0 32px 0; font-size: 16px; line-height: 1.6; color: #4b5563;\">Renew your subscription today to ensure uninterrupted access to all premium features.</p><a href=\"{{renew_url}}\" style=\"display: inline-block; padding: 14px 32px; background-color: #000000; color: #ffffff; font-size: 14px; font-weight: 600; text-decoration: none; border-radius: 8px;\">Renew Subscription</a></div>"}'
    ) ON CONFLICT (template_id, channel) DO NOTHING;