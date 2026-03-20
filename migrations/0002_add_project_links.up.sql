CREATE TABLE IF NOT EXISTS project_links (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    link_type VARCHAR(50) NOT NULL,
    url TEXT NOT NULL,
    label VARCHAR(255)
);

INSERT INTO project_links (service_id, link_type, url, label)
SELECT id, 'main', url, name
FROM services
WHERE url IS NOT NULL AND url != '';
