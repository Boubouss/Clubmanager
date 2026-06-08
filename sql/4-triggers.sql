DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP TRIGGER IF EXISTS trigger_clubs_updated_at ON clubs;
DROP TRIGGER IF EXISTS trigger_members_updated_at ON members;
DROP TRIGGER IF EXISTS trigger_roles_updated_at ON roles;

CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_clubs_updated_at
BEFORE UPDATE ON clubs
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_members_updated_at
BEFORE UPDATE ON members
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_roles_updated_at
BEFORE UPDATE ON roles
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trigger_licences_updated_at ON licences;
CREATE TRIGGER trigger_licences_updated_at
BEFORE UPDATE ON licences
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trigger_log_licences ON licences;
CREATE TRIGGER trigger_log_licences
AFTER INSERT OR UPDATE OR DELETE ON licences
FOR EACH ROW
EXECUTE FUNCTION log_changes();

DROP TRIGGER IF EXISTS trigger_events_updated_at ON events;
CREATE TRIGGER trigger_events_updated_at
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trigger_log_events ON events;
CREATE TRIGGER trigger_log_events
AFTER INSERT OR UPDATE OR DELETE ON events
FOR EACH ROW
EXECUTE FUNCTION log_changes();

DROP TRIGGER IF EXISTS trigger_log_users ON users;
DROP TRIGGER IF EXISTS trigger_log_members ON members;
DROP TRIGGER IF EXISTS trigger_log_clubs ON clubs;
DROP TRIGGER IF EXISTS trigger_log_roles ON roles;

CREATE TRIGGER trigger_log_users
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW
EXECUTE FUNCTION log_changes();

CREATE TRIGGER trigger_log_members
AFTER INSERT OR UPDATE OR DELETE ON members
FOR EACH ROW
EXECUTE FUNCTION log_changes();

CREATE TRIGGER trigger_log_clubs
AFTER INSERT OR UPDATE OR DELETE ON clubs
FOR EACH ROW
EXECUTE FUNCTION log_changes();

CREATE TRIGGER trigger_log_roles
AFTER INSERT OR UPDATE OR DELETE ON roles
FOR EACH ROW
EXECUTE FUNCTION log_changes();
