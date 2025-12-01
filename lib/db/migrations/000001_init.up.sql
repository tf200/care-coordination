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


CREATE TABLE  employees (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    bsn TEXT UNIQUE NOT NULL,
    date_of_birth DATE NOT NULL,
    phone_number TEXT NOT NULL,
    gender gender_enum NOT NULL,
    role TEXT NOT NULL
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

-- organization information
    org_name TEXT NOT NULL,
    org_contact_person TEXT NOT NULL,
    org_phone_number TEXT NOT NULL,
    org_email TEXT NOT NULL,
-- registration for 
    care_type care_type_enum NOT NULL,
    coordinator_id TEXT NOT NULL REFERENCES employees(id),
    registration_date TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    registration_reason TEXT NOT NULL,
    additional_notes TEXT,
    status registration_status_enum DEFAULT 'pending',
    attachment_ids TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
)
