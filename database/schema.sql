-- ============================================================================
-- अनंत अभ्यास (ANANT ABHYAS) - 360° पैन-इंडिया प्रोडक्शन डेटाबेस स्कीमा
-- ============================================================================

-- 1. Users Table (Core Student & Account Profile + PIN Security + UPI Grace Flow)
CREATE TABLE IF NOT EXISTS users (
    phone VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    grade INT NOT NULL,
    state VARCHAR(50),                         -- ऑनबोर्डिंग द्वारा निर्धारित (कोई हार्डकोडेड राज्य नहीं)
    district VARCHAR(50),                      -- ऑनबोर्डिंग द्वारा निर्धारित (कोई हार्डकोडेड जिला नहीं)
    preferred_dialect VARCHAR(50),             -- शुद्ध डायनामिक (यूजर चयन/डिटेक्शन पर आधारित, कोई भाषा पक्षपात नहीं)
    custom_interest_topic VARCHAR(100),        -- बच्चे की व्यक्तिगत रुचि (डायनामिक AI प्रोफाइलिंग)
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- 🔐 ज़ीरो-ट्रस्ट मास्टर पिन सुरक्षा
    pin_hash VARCHAR(255),
    pin_salt VARCHAR(64),
    pin_locked_until TIMESTAMP WITH TIME ZONE,
    failed_attempts INT DEFAULT 0,
    last_failed_at TIMESTAMP WITH TIME ZONE,
    daily_bypass_count INT DEFAULT 0,
    bypass_reset_at TIMESTAMP WITH TIME ZONE,

    -- 💳 UPI ऑटो-पे 48-घंटे ग्रेस पीरियड व मैंडेट सुरक्षा
    payment_grace_until TIMESTAMP WITH TIME ZONE,
    last_mandate_status VARCHAR(30) DEFAULT 'ACTIVE'
);

-- 2. State Academic Calendars (पैन-इंडिया 28 राज्यों व केंद्र शासित प्रदेशों का अवकाश कैलेंडर)
CREATE TABLE IF NOT EXISTS state_academic_calendars (
    id SERIAL PRIMARY KEY,
    state VARCHAR(50) NOT NULL,
    district VARCHAR(50) DEFAULT 'ALL',
    holiday_type VARCHAR(30) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    CONSTRAINT uq_state_holiday UNIQUE (state, district, holiday_type, start_date)
);

-- 3. Student Exam Schedules (नवोदय, सैनिक स्कूल, बोर्ड व स्थानीय परीक्षाएं)
CREATE TABLE IF NOT EXISTS student_exam_schedules (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(20) REFERENCES users(phone) ON DELETE CASCADE,
    exam_type VARCHAR(30) NOT NULL,
    subject VARCHAR(50),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Submission Logs (OCR, Homework Hashes & DPDPA Retention)
CREATE TABLE IF NOT EXISTS submission_logs (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(20) REFERENCES users(phone) ON DELETE CASCADE,
    image_hash VARCHAR(64) NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_student_submission UNIQUE (student_phone, image_hash)
);

-- 5. Student Holiday Assignments (Vacation Mode Quota)
CREATE TABLE IF NOT EXISTS student_holiday_assignments (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(20) REFERENCES users(phone) ON DELETE CASCADE UNIQUE,
    total_assigned_tasks INT DEFAULT 0,
    completed_tasks INT DEFAULT 0,
    allocated_vacation_days INT NOT NULL,
    daily_task_quota INT DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Parent Feedback Tickets & Dialect Voice Logs
CREATE TABLE IF NOT EXISTS parent_feedback_tickets (
    id SERIAL PRIMARY KEY,
    student_phone VARCHAR(20) REFERENCES users(phone) ON DELETE CASCADE,
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
    student_phone VARCHAR(20) REFERENCES users(phone) ON DELETE CASCADE,
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

-- 9. Family Children (60-Day Anti-Churn Child Lock & Student Profile Mapping)
CREATE TABLE IF NOT EXISTS family_children (
    id VARCHAR(64) PRIMARY KEY,
    parent_uid VARCHAR(64) REFERENCES parent_accounts(parent_uid) ON DELETE CASCADE,
    student_phone VARCHAR(20) REFERENCES users(phone) ON DELETE SET NULL, -- सीधे छात्र प्रोफाइल से लिंक
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    grade INT NOT NULL,
    school_name VARCHAR(150) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    locked_till TIMESTAMP WITH TIME ZONE NOT NULL
);

-- 10. Security Audit Logs (पिन सुरक्षा, ऑडिट और पैरेंटल ओवरराइड ट्रैकिंग)
CREATE TABLE IF NOT EXISTS security_audit_logs (
    id SERIAL PRIMARY KEY,
    parent_phone VARCHAR(20) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- इंडेक्सिंग (High Concurrency, DPDPA Purge & Query Acceleration)
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_parent_primary_phone ON parent_accounts(primary_phone);
CREATE INDEX IF NOT EXISTS idx_family_children_parent ON family_children(parent_uid);
CREATE INDEX IF NOT EXISTS idx_family_children_student ON family_children(student_phone);
CREATE INDEX IF NOT EXISTS idx_submission_logs_phone ON submission_logs(student_phone);
CREATE INDEX IF NOT EXISTS idx_security_audit_phone ON security_audit_logs(parent_phone);

-- 30-दिवसीय DPDPA 2023 ऑटो-डिलीशन को तेज करने के लिए टाइमस्टैम्प इंडेक्स
CREATE INDEX IF NOT EXISTS idx_submission_logs_retention ON submission_logs(submitted_at);
