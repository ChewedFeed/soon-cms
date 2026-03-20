CREATE TABLE IF NOT EXISTS project_milestones (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL DEFAULT 'feature',
    status VARCHAR(50) NOT NULL DEFAULT 'planned',
    target_date DATE,
    completed_date DATE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO project_milestones (
    service_id,
    title,
    description,
    category,
    status,
    target_date,
    completed_date,
    sort_order,
    created_at,
    updated_at
)
SELECT
    service_id,
    name,
    NULL,
    CASE WHEN release_date IS NOT NULL THEN 'release' ELSE 'feature' END,
    CASE WHEN completed THEN 'completed' ELSE 'planned' END,
    target_date,
    release_date,
    sort_order,
    NOW(),
    NOW()
FROM project_roadmap
WHERE NOT EXISTS (
    SELECT 1
    FROM project_milestones pm
    WHERE pm.service_id = project_roadmap.service_id
      AND pm.title = project_roadmap.name
      AND COALESCE(pm.target_date, DATE '1900-01-01') = COALESCE(project_roadmap.target_date, DATE '1900-01-01')
      AND pm.sort_order = project_roadmap.sort_order
);

INSERT INTO project_milestones (
    service_id,
    title,
    description,
    category,
    status,
    target_date,
    completed_date,
    sort_order,
    created_at,
    updated_at
)
SELECT
    service_id,
    COALESCE(NULLIF(task, ''), 'Launch task #' || launch_task_id),
    NULL,
    'other',
    CASE WHEN completed THEN 'completed' ELSE 'planned' END,
    NULL,
    NULL,
    COALESCE(launch_task_id, 0),
    created_at,
    NOW()
FROM launch_task
WHERE NOT EXISTS (
    SELECT 1
    FROM project_milestones pm
    WHERE pm.service_id = launch_task.service_id
      AND pm.title = COALESCE(NULLIF(launch_task.task, ''), 'Launch task #' || launch_task.launch_task_id)
      AND pm.category = 'other'
      AND pm.sort_order = COALESCE(launch_task.launch_task_id, 0)
);

DROP TABLE IF EXISTS project_roadmap;
DROP TABLE IF EXISTS launch_task;
