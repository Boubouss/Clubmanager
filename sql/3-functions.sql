CREATE OR REPLACE FUNCTION update_updated_at() 
RETURNS TRIGGER AS $$
BEGIN 
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION log_changes()
RETURNS TRIGGER AS $$
DECLARE
    log_record RECORD;
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO logs (
            table_name, record_id, action, new_data, changed_by, ip_addr, user_agent
        ) VALUES (
            TG_TABLE_NAME,
            NEW.id,
            'INSERT',
            to_jsonb(NEW.*),
            NULLIF(current_setting('app.current_user_id'), '')::UUID,
            NULLIF(current_setting('app.client_ip'), '')::INET,
            NULLIF(current_setting('app.user_agent'), '')::TEXT
        );
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO logs (
            table_name, record_id, action, old_data, new_data, changed_by, ip_addr, user_agent
        ) VALUES (
            TG_TABLE_NAME,
            NEW.id,
            'UPDATE',
            to_jsonb(OLD.*),
            to_jsonb(NEW.*),
            NULLIF(current_setting('app.current_user_id'), '')::UUID,
            NULLIF(current_setting('app.client_ip'), '')::INET,
            NULLIF(current_setting('app.user_agent'), '')::TEXT
        );
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO logs (
            table_name, record_id, action, old_data, changed_by, ip_addr, user_agent
        ) VALUES (
            TG_TABLE_NAME,
            OLD.id,
            'DELETE',
            to_jsonb(OLD.*),
            NULLIF(current_setting('app.current_user_id'), '')::UUID,
            NULLIF(current_setting('app.client_ip'), '')::INET,
            NULLIF(current_setting('app.user_agent'), '')::TEXT
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
