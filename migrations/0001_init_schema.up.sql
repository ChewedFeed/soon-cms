CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    search_name VARCHAR(255),
    description TEXT,
    full_description TEXT,
    url TEXT,
    non_url TEXT,
    alternatives TEXT,
    launch_year INTEGER DEFAULT 0,
    launch_month INTEGER DEFAULT 0,
    launch_day INTEGER DEFAULT 0,
    progress INTEGER DEFAULT 0,
    icon VARCHAR(255),
    uptime VARCHAR(255),
    live BOOLEAN DEFAULT false,
    started BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS launch_task (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    completed BOOLEAN DEFAULT false
);
