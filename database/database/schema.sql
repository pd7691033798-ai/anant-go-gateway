-- 1. Users Table (Core Student & Account Profile)
CREATE TABLE IF NOT EXISTS users (
    phone VARCHAR(15) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    grade INT NOT NULL,
    state VARCHAR(50) DEFAULT 'Rajasthan',
    district VARCHAR(50) DEFAULT 'Sri Ganganagar',
    preferred_dialect VARCHAR(50) DEFAULT 'BAGRI_PUNJABI_FUSION',
    custom_interest_topic VARCHAR(100) DEFAULT 'रोबोटिक्स और कार इंजन',
    plan_tier VARCHAR(20) DEFAULT 'DEMO',
    plan_expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '7 days',
    consecutive_paid_months INT DEFAULT 1,
    streak_count INT DEFAULT 0,
    last_scan_date DATE DEFAULT CURRENT_DATE,
    consecutive_missed_days INT DEFAULT 0,
    primary_device_hash VARCHAR(100),
    current_location_city VARCHAR(50),
    sharing_suspicion_score INT DEFAULT 0,
    detected_grade_drift_count INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. State Academic Calendars
CREATE TABLE IF NOT EXISTS state_academic_calendars (
    id SERIAL PRIMARY KEY,
    state VARCHAR(50) NOT NULL,
    district VARCHAR(50) DEFAULT 'ALL',
    holiday_type VARCHAR(30) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

-- 3. Student Exam Schedules
CREATE TABLE IF NOT EXISTS student_exam_schedules (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(15) REFERENCES users(phone),
    exam_type VARCHAR(30) NOT NULL,
    subject VARCHAR(50),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Submission Logs (OCR & Homework Hashes)
CREATE TABLE IF NOT EXISTS submission_logs (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(15) REFERENCES users(phone),
    image_hash VARCHAR(64) NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Student Holiday Assignments (Vacation Mode)
CREATE TABLE IF NOT EXISTS student_holiday_assignments (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(15) REFERENCES users(phone) UNIQUE,
    total_assigned_tasks INT DEFAULT 0,
    completed_tasks INT DEFAULT 0,
    allocated_vacation_days INT NOT NULL,
    daily_task_quota INT DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Parent Feedback Tickets & Dialect Voice Logs
CREATE TABLE IF NOT EXISTS parent_feedback_tickets (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(15) REFERENCES users(phone),
    state VARCHAR(50),
    district VARCHAR(50),
    detected_dialect VARCHAR(30),
    raw_parent_message TEXT NOT NULL,
    sentiment_category VARCHAR(30),
    urgency_score INT DEFAULT 1,
    should_escalate BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 7. Multi-User Violation Logs (Account Sharing Guard)
CREATE TABLE IF NOT EXISTS multi_user_violations (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(15) REFERENCES users(phone),
    detected_issue VARCHAR(50),
    confidence_score FLOAT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 8. Parent Accounts (Family & Family Unlimited Master Accounts)
CREATE TABLE IF NOT EXISTS parent_accounts (
    parent_uid VARCHAR(64) PRIMARY KEY,
    parent_name VARCHAR(100) NOT NULL,
    primary_phone VARCHAR(20) NOT NULL UNIQUE,
    family_surname VARCHAR(50),
    active_device_id VARCHAR(128),
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    plan_tier VARCHAR(32) DEFAULT 'BASIC'
);

-- 9. Family Children (60-Day Anti-Churn Child Lock)
CREATE TABLE IF NOT EXISTS family_children (
    id VARCHAR(64) PRIMARY KEY,
    parent_uid VARCHAR(64) REFERENCES parent_accounts(parent_uid) ON DELETE CASCADE,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    grade INT NOT NULL,
    school_name VARCHAR(150) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    locked_till TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Indexing for Fast Query Performance & High Concurrency
CREATE INDEX IF NOT EXISTS idx_family_children_parent ON family_children(parent_uid);
CREATE INDEX IF NOT EXISTS idx_parent_primary_phone ON parent_accounts(primary_phone);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_submission_logs_phone ON submission_logs(student_phone);

