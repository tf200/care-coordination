-- Drop tables in reverse order of creation (respecting foreign key dependencies)
-- Most dependent tables first, then their dependencies

DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
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
DROP TYPE IF EXISTS incident_severity_enum;
DROP TYPE IF EXISTS incident_type_enum;
DROP TYPE IF EXISTS incident_status_enum;
DROP TYPE IF EXISTS discharge_status_enum;
DROP TYPE IF EXISTS discharge_reason_enum;
DROP TYPE IF EXISTS waiting_list_priority_enum;
DROP TYPE IF EXISTS client_status_enum;
DROP TYPE IF EXISTS intake_status_enum;
DROP TYPE IF EXISTS registration_status_enum;
DROP TYPE IF EXISTS care_type_enum;
DROP TYPE IF EXISTS gender_enum;
