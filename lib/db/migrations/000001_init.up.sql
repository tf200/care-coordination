CREATE TABLE attachments (
    id TEXT PRIMARY KEY,
    filekey TEXT NOT NULL,
    content_type TEXT NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE gender_enum AS ENUM ('male', 'female', 'other');


CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_id ON users(id);


CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    token_family TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    user_agent TEXT  NULL,  
    ip_address TEXT  NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);



CREATE TABLE locations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    address TEXT NOT NULL,
    capacity INTEGER NOT NULL,
    occupied INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE referring_orgs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    contact_person TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);



CREATE TABLE  employees (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    bsn TEXT UNIQUE NOT NULL,
    date_of_birth DATE NOT NULL,
    phone_number TEXT NOT NULL,
    gender gender_enum NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);
 

CREATE TYPE care_type_enum AS ENUM ('protected_living', 'semi_independent_living', 'independent_assisted_living', 'ambulatory_care');
CREATE TYPE registration_status_enum AS ENUM ('pending', 'approved', 'rejected', 'in_review');

 CREATE TABLE registration_forms (
 -- client information
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    bsn TEXT UNIQUE NOT NULL,
    date_of_birth DATE NOT NULL,
    gender gender_enum NOT NULL,

-- organization information
    reffering_org_id TEXT REFERENCES referring_orgs(id),
-- registration for 
    care_type care_type_enum NOT NULL,
    registration_date DATE  DEFAULT CURRENT_DATE,
    registration_reason TEXT NOT NULL,
    additional_notes TEXT,
    status registration_status_enum DEFAULT 'pending',
    attachment_ids TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE
);


CREATE TYPE intake_status_enum AS ENUM ('completed', 'pending');
CREATE TABLE intake_forms (
    id TEXT PRIMARY KEY,
    registration_form_id TEXT NOT NULL UNIQUE REFERENCES registration_forms(id),
    intake_date DATE NOT NULL,
    intake_Time TIME NOT NULL,
    location_id TEXT NOT NULL REFERENCES locations(id),
    coordinator_id TEXT NOT NULL REFERENCES employees(id),
    family_situation TEXT,
    main_provider TEXT,
    limitations TEXT,
    focus_areas TEXT,
    goals TEXT,
    notes TEXT,
    status intake_status_enum NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);


CREATE TYPE client_status_enum AS ENUM ('waiting_list', 'in_care', 'discharged');
CREATE TYPE waiting_list_priority_enum AS ENUM ('low', 'normal', 'high');
CREATE TYPE discharge_reason_enum AS ENUM (
    'treatment_completed', 
    'terminated_by_mutual_agreement', 
    'terminated_by_client',
    'terminated_by_provider',
    'terminated_due_to_external_factors',
    'other'
    );
CREATE TYPE discharge_status_enum AS ENUM ('in_progress', 'completed');
CREATE TABLE clients (
    id TEXT PRIMARY KEY,
    -- Client personal information (from registration)
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    bsn TEXT  NOT NULL,
    date_of_birth DATE NOT NULL,
    phone_number TEXT,  
    gender gender_enum NOT NULL,
    
    -- Registration and intake references
    registration_form_id TEXT NOT NULL UNIQUE REFERENCES registration_forms(id),
    intake_form_id TEXT NOT NULL UNIQUE REFERENCES intake_forms(id),
    
    -- Care information
    care_type care_type_enum NOT NULL,
    ambulatory_weekly_hours INTEGER NULL,
    referring_org_id TEXT REFERENCES referring_orgs(id),
    
    -- status Management
    status client_status_enum NOT NULL DEFAULT 'waiting_list',
    waiting_list_priority waiting_list_priority_enum NOT NULL DEFAULT 'normal',

    -- In care management
    care_start_date DATE NULL,
    care_end_date DATE NULL,

    -- Discharge information
    discharge_date DATE NULL,
    closing_report TEXT NULL,
    evaluation_report TEXT NULL,
    reason_for_discharge discharge_reason_enum NULL,
    discharge_attachment_ids TEXT[] NULL,
    discharge_status discharge_status_enum NULL,

    
    -- Assigned location (null while on waiting list)
    assigned_location_id TEXT NOT NULL REFERENCES locations(id),

    
    -- Care team
    coordinator_id TEXT NOT NULL REFERENCES employees(id),
    
    -- Additional information
    family_situation TEXT,
    limitations TEXT,
    focus_areas TEXT,
    goals TEXT,
    notes TEXT,
    
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint: ambulatory_weekly_hours only allowed for ambulatory_care
    CONSTRAINT chk_ambulatory_hours CHECK (
        (care_type = 'ambulatory_care' AND ambulatory_weekly_hours IS NOT NULL AND ambulatory_weekly_hours > 0) OR
        (care_type != 'ambulatory_care' AND ambulatory_weekly_hours IS NULL)
    ),

    -- Status integrity constraints
    -- 1. Waiting list clients should not have care/discharge dates
    CONSTRAINT chk_waiting_list_fields CHECK (
        status != 'waiting_list' OR (
            care_start_date IS NULL AND 
            care_end_date IS NULL AND
            discharge_date IS NULL AND 
            discharge_status IS NULL AND
            reason_for_discharge IS NULL
        )
    ),

    -- 2. In-care clients must have care_start_date, no discharge info yet
    CONSTRAINT chk_in_care_fields CHECK (
        status != 'in_care' OR (
            care_start_date IS NOT NULL AND 
            discharge_date IS NULL AND
            discharge_status IS NULL
        )
    ),

    -- 3. Discharged clients must have required discharge information
    CONSTRAINT chk_discharged_fields CHECK (
        status != 'discharged' OR (
            care_start_date IS NOT NULL AND
            discharge_date IS NOT NULL AND 
            reason_for_discharge IS NOT NULL AND
            discharge_status IS NOT NULL
        )
    ),

    -- 4. Discharge date should be on or after care start date
    CONSTRAINT chk_discharge_after_care_start CHECK (
        discharge_date IS NULL OR care_start_date IS NULL OR discharge_date >= care_start_date
    ),

    -- 5. Care end date (planned) should be after care start date
    CONSTRAINT chk_care_end_after_start CHECK (
        care_end_date IS NULL OR care_start_date IS NULL OR care_end_date >= care_start_date
    )
);



CREATE TABLE client_location_transfers (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL REFERENCES clients(id),
    from_location_id TEXT REFERENCES locations(id),
    to_location_id TEXT NOT NULL REFERENCES locations(id),
    current_coordinator_id TEXT NOT NULL REFERENCES employees(id),
    new_coordinator_id TEXT NOT NULL REFERENCES employees(id),
    transfer_date TIMESTAMP NOT NULL,
    reason TEXT,
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);



CREATE TYPE incident_status_enum AS ENUM ('pending', 'under_investigation', 'completed');
CREATE TYPE incident_type_enum AS ENUM ('aggression', 'medical_emergency', 'safety_concern', 'unwanted_behavior', 'other');
CREATE TYPE incident_severity_enum AS ENUM ('minor', 'moderate', 'severe');
CREATE TABLE incidents (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL REFERENCES clients(id),
    incident_date DATE NOT NULL,
    incident_time TIME NOT NULL,
    incident_type incident_type_enum NOT NULL,
    incident_severity incident_severity_enum NOT NULL,
    location_id TEXT NOT NULL REFERENCES locations(id),
    coordinator_id TEXT NOT NULL REFERENCES employees(id),
    incident_description TEXT NOT NULL,
    action_taken TEXT NOT NULL,
    other_parties TEXT,
    status incident_status_enum NOT NULL,
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);