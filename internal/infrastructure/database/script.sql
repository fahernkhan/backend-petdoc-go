CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Tambahkan trigger ke tabel doctors
CREATE TRIGGER update_doctors_updated_at
BEFORE UPDATE ON doctors
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- INDEXING
-- Pada tabel doctors
CREATE INDEX idx_doctors_user_id ON doctors(user_id);
CREATE INDEX idx_doctors_updated_at ON doctors(updated_at);

-- Pada tabel users
CREATE INDEX idx_users_role ON users(role);

CREATE INDEX idx_doctors_updated_at ON doctors(updated_at);

-- Index in consultations
CREATE INDEX idx_consultations_doctor_time ON consultations (doctor_id, start_time, end_time);
CREATE INDEX idx_consultations_user_time ON consultations (user_id, start_time, end_time);
CREATE INDEX idx_consultations_date ON consultations (consultation_date);