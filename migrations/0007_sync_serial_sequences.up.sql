SELECT setval('services_id_seq', COALESCE((SELECT MAX(id) FROM services), 1), true);
SELECT setval('project_links_id_seq', COALESCE((SELECT MAX(id) FROM project_links), 1), true);
SELECT setval('project_roadmap_id_seq', COALESCE((SELECT MAX(id) FROM project_roadmap), 1), true);
SELECT setval('project_milestones_id_seq', COALESCE((SELECT MAX(id) FROM project_milestones), 1), true);
SELECT setval('launch_task_launch_task_id_seq', COALESCE((SELECT MAX(launch_task_id) FROM launch_task), 1), true);
SELECT setval('admin_users_id_seq', COALESCE((SELECT MAX(id) FROM admin_users), 1), true);
SELECT setval('admin_sessions_id_seq', COALESCE((SELECT MAX(id) FROM admin_sessions), 1), true);
