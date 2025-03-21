-- Table users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255),
    gender VARCHAR(10),
    username VARCHAR(255) UNIQUE,
    date_of_birth DATE,
    role VARCHAR(50) DEFAULT 'user',
    provider_id VARCHAR(255),
    provider_name VARCHAR(50),
    otp_secret VARCHAR(255),       
    otp_enabled BOOLEAN DEFAULT FALSE, 
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Table doctors
CREATE TABLE doctors (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Relasi ke tabel users
    full_name VARCHAR(255) NOT NULL,
    last_education VARCHAR(255) NOT NULL,
    specialist_at VARCHAR(255) NOT NULL,
    profile_image VARCHAR(255),
    birth_date DATE,
    hospital_name VARCHAR(255),
    years_of_experience INT,
    price_per_hour DECIMAL(10, 2) NOT NULL,
    gmeet_link VARCHAR(255) NOT NULL,
    working_days JSONB, -- Contoh: ["Senin", "Rabu", "Jumat"]
    working_hours JSONB, -- Contoh: { "start": "09:00", "end": "17:00" }
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Table consultations
CREATE TABLE consultations (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    doctor_id INT REFERENCES doctors(id) ON DELETE CASCADE,
    pet_type VARCHAR(255) NOT NULL,
    pet_name VARCHAR(255) NOT NULL,
    pet_age INT NOT NULL,
    disease_description TEXT NOT NULL,
    consultation_date DATE NOT NULL, -- Tanggal konsultasi
    start_time TIMESTAMPTZ NOT NULL, -- Waktu mulai konsultasi (dengan zona waktu)
    end_time TIMESTAMPTZ NOT NULL, -- Waktu selesai konsultasi (dengan zona waktu)
    payment_proof TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);