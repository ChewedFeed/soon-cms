CREATE TABLE IF NOT EXISTS project_roadmap (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_date DATE,
    release_date DATE,
    completed BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0
);
