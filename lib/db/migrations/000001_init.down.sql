-- Drop tables in reverse order of creation (respecting foreign key dependencies)
-- Most dependent tables first, then their dependencies

-- First, drop all RLS policies that depend on user_roles
DROP POLICY IF EXISTS coordinator_progress_logs ON goal_progress_logs;
DROP POLICY IF EXISTS admin_all_progress_logs ON goal_progress_logs;
DROP POLICY IF EXISTS coordinator_evaluations ON client_evaluations;
DROP POLICY IF EXISTS admin_all_evaluations ON client_evaluations;
DROP POLICY IF EXISTS coordinator_goals ON client_goals;
DROP POLICY IF EXISTS admin_all_goals ON client_goals;
DROP POLICY IF EXISTS coordinator_own_clients ON clients;
DROP POLICY IF EXISTS admin_all ON clients;

-- Drop calendar-related RLS policies
DROP POLICY IF EXISTS admin_all_appointments ON appointments;
DROP POLICY IF EXISTS coordinator_appointments ON appointments;
DROP POLICY IF EXISTS admin_all_participants ON appointment_participants;
DROP POLICY IF EXISTS coordinator_participants ON appointment_participants;
DROP POLICY IF EXISTS admin_all_reminders ON reminders;
DROP POLICY IF EXISTS user_own_reminders ON reminders;

-- Drop notification RLS policy
DROP POLICY IF EXISTS user_own_notifications ON notifications;

-- Drop notifications table
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_priority_enum;
DROP TYPE IF EXISTS notification_type_enum;

-- Then drop the tables
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS appointment_external_mappings;
DROP TABLE IF EXISTS calendar_integrations;
DROP TABLE IF EXISTS reminders;
DROP TABLE IF EXISTS appointment_participants;
DROP TABLE IF EXISTS appointments;
DROP TYPE IF EXISTS participant_type_enum;
DROP TYPE IF EXISTS appointment_type_enum;
DROP TYPE IF EXISTS appointment_status_enum;

DROP TABLE IF EXISTS goal_progress_logs;
DROP TABLE IF EXISTS client_evaluations;

DROP TABLE IF EXISTS client_goals;
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS client_location_transfers;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS intake_forms;
DROP TABLE IF EXISTS registration_forms;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS referring_orgs;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS attachments;

-- Drop enums
DROP TYPE IF EXISTS incident_severity_enum CASCADE;
DROP TYPE IF EXISTS incident_type_enum CASCADE;
DROP TYPE IF EXISTS incident_status_enum CASCADE;
DROP TYPE IF EXISTS goal_progress_status CASCADE;
DROP TYPE IF EXISTS discharge_status_enum CASCADE;
DROP TYPE IF EXISTS discharge_reason_enum CASCADE;
DROP TYPE IF EXISTS waiting_list_priority_enum CASCADE;
DROP TYPE IF EXISTS client_status_enum CASCADE;
DROP TYPE IF EXISTS intake_status_enum CASCADE;
DROP TYPE IF EXISTS registration_status_enum CASCADE;
DROP TYPE IF EXISTS care_type_enum CASCADE;
DROP TYPE IF EXISTS gender_enum CASCADE;
DROP TYPE IF EXISTS location_transfer_status_enum CASCADE;
DROP TYPE IF EXISTS contract_type_enum CASCADE;
DROP TYPE IF EXISTS evaluation_status_enum CASCADE;
