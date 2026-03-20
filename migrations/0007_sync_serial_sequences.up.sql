DO $$
BEGIN
    IF to_regclass('public.services_id_seq') IS NOT NULL THEN
        PERFORM setval('services_id_seq', COALESCE((SELECT MAX(id) FROM services), 1), true);
    END IF;

    IF to_regclass('public.project_links_id_seq') IS NOT NULL THEN
        PERFORM setval('project_links_id_seq', COALESCE((SELECT MAX(id) FROM project_links), 1), true);
    END IF;

    IF to_regclass('public.project_roadmap_id_seq') IS NOT NULL THEN
        PERFORM setval('project_roadmap_id_seq', COALESCE((SELECT MAX(id) FROM project_roadmap), 1), true);
    END IF;

    IF to_regclass('public.project_milestones_id_seq') IS NOT NULL THEN
        PERFORM setval('project_milestones_id_seq', COALESCE((SELECT MAX(id) FROM project_milestones), 1), true);
    END IF;

    IF to_regclass('public.launch_task_launch_task_id_seq') IS NOT NULL THEN
        PERFORM setval('launch_task_launch_task_id_seq', COALESCE((SELECT MAX(launch_task_id) FROM launch_task), 1), true);
    END IF;

    IF to_regclass('public.admin_users_id_seq') IS NOT NULL THEN
        PERFORM setval('admin_users_id_seq', COALESCE((SELECT MAX(id) FROM admin_users), 1), true);
    END IF;

    IF to_regclass('public.admin_sessions_id_seq') IS NOT NULL THEN
        PERFORM setval('admin_sessions_id_seq', COALESCE((SELECT MAX(id) FROM admin_sessions), 1), true);
    END IF;
END $$;
