DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_username ON users(username);
CREATE UNIQUE INDEX idx_users_email ON users(email);

DROP INDEX IF EXISTS idx_clubs_siren;
DROP INDEX IF EXISTS idx_clubs_name;
DROP INDEX IF EXISTS idx_clubs_city;
DROP INDEX IF EXISTS idx_clubs_postal_code;
CREATE UNIQUE INDEX idx_clubs_siren ON clubs(siren);
CREATE INDEX idx_clubs_name ON clubs(LOWER(name));
CREATE INDEX idx_clubs_city ON clubs(LOWER(city));
CREATE INDEX idx_clubs_postal_code ON clubs(postal_code);

DROP INDEX IF EXISTS idx_members_user_id;
DROP INDEX IF EXISTS idx_members_club_id;
DROP INDEX IF EXISTS idx_members_user_club;
DROP INDEX IF EXISTS idx_members_firstname;
DROP INDEX IF EXISTS idx_members_lastname;
DROP INDEX IF EXISTS idx_members_name;
DROP INDEX IF EXISTS idx_members_birthdate;
CREATE INDEX idx_members_user_id ON members(user_id);
CREATE INDEX idx_members_club_id ON members(club_id);
CREATE INDEX idx_members_user_club ON members(user_id, club_id);
CREATE INDEX idx_members_firstname ON members(LOWER(firstname), club_id);
CREATE INDEX idx_members_lastname ON members(LOWER(lastname), club_id);
CREATE INDEX idx_members_name ON members(LOWER(firstname), LOWER(lastname), club_id);
CREATE INDEX idx_members_birthdate ON members(birthdate, club_id);

DROP INDEX IF EXISTS idx_roles_user_id;
DROP INDEX IF EXISTS idx_roles_club_id;
DROP INDEX IF EXISTS idx_roles_user_club;
CREATE INDEX idx_roles_user_id ON roles(user_id);
CREATE INDEX idx_roles_club_id ON roles(club_id);
CREATE INDEX idx_roles_user_club ON roles(user_id, club_id);

DROP INDEX IF EXISTS idx_logs_record_id;
DROP INDEX IF EXISTS idx_logs_changed_at;
DROP INDEX IF EXISTS idx_logs_table_name;
DROP INDEX IF EXISTS idx_logs_changed_by;
CREATE INDEX idx_logs_record_id ON logs(record_id);
CREATE INDEX idx_logs_changed_at ON logs(changed_at);
CREATE INDEX idx_logs_table_name ON logs(table_name);
CREATE INDEX idx_logs_changed_by ON logs(changed_by);
