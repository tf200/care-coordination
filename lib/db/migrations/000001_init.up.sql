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
    registration_date TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    registration_reason TEXT NOT NULL,
    additional_notes TEXT,
    status registration_status_enum DEFAULT 'pending',
    attachment_ids TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);


CREATE TYPE client_status_enum AS ENUM ('waiting_list', 'in_care', 'discharged');
CREATE TYPE waiting_list_priority_enum AS ENUM ('low', 'normal', 'high');
CREATE TABLE clients (
    id TEXT PRIMARY KEY,
    -- Client personal information (from registration)
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    bsn TEXT UNIQUE NOT NULL,
    date_of_birth DATE NOT NULL,
    phone_number TEXT,
    gender gender_enum NOT NULL,
    
    -- Registration and intake references
    registration_form_id TEXT NOT NULL UNIQUE REFERENCES registration_forms(id),
    intake_form_id TEXT NOT NULL UNIQUE REFERENCES intake_forms(id),
    
    -- Care information
    care_type care_type_enum NOT NULL,
    accomodation_weeks INTEGER NULL,
    ambulant_care_hours_per_week INTEGER NULL,
    referring_org_id TEXT REFERENCES referring_orgs(id),
    
    -- status Management
    status client_status_enum NOT NULL DEFAULT 'waiting_list',
    waiting_list_priority waiting_list_priority_enum NOT NULL DEFAULT 'normal',

    -- In care management
    care_start_date DATE NULL,
    care_end_date DATE NULL,

    
    -- Assigned location (null while on waiting list)
    assigned_location_id TEXT REFERENCES locations(id),

    
    -- Care team
    coordinator_id TEXT REFERENCES employees(id),
    
    -- Additional information
    family_situation TEXT,
    limitations TEXT,
    focus_areas TEXT,
    goals TEXT,
    notes TEXT,
    
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP
);



