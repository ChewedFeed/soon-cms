ALTER TABLE services
ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'planned';

UPDATE services
SET status = CASE
    WHEN live = true THEN 'active'
    ELSE 'planned'
END
WHERE status = 'planned';
