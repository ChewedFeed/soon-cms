CREATE TABLE IF NOT EXISTS project_roadmap (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_date DATE,
    release_date DATE,
    completed BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS launch_task (
    launch_task_id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed BOOLEAN DEFAULT false,
    task TEXT
);

INSERT INTO project_roadmap (service_id, name, target_date, release_date, completed, sort_order)
SELECT
    service_id,
    title,
    target_date,
    completed_date,
    status = 'completed',
    sort_order
FROM project_milestones
WHERE category <> 'other';

INSERT INTO launch_task (service_id, created_at, completed, task)
SELECT
    service_id,
    created_at,
    status = 'completed',
    title
FROM project_milestones
WHERE category = 'other';

DROP TABLE IF EXISTS project_milestones;
