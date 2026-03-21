UPDATE services
SET progress = 0
WHERE progress IS NULL;

ALTER TABLE services
    ALTER COLUMN progress SET DEFAULT 0;
