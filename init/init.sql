DO $$
DECLARE

BEGIN
    -- Idea Statuses
    INSERT INTO "status" (title, is_deleted, created_at) VALUES ('Создана', false, NOW()) ON CONFLICT (title) DO NOTHING;
    INSERT INTO "status" (title, is_deleted, created_at) VALUES ('В работе', false, NOW()) ON CONFLICT (title) DO NOTHING;
    INSERT INTO "status" (title, is_deleted, created_at) VALUES ('Реализована', false, NOW()) ON CONFLICT (title) DO NOTHING;
    INSERT INTO "status" (title, is_deleted, created_at) VALUES ('Отклонена', false, NOW()) ON CONFLICT (title) DO NOTHING;
END $$;