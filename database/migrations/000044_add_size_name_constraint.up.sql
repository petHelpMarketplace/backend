-- Create a function to validate animal size
CREATE OR REPLACE FUNCTION validate_animal_size()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.size IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM animal_sizes WHERE name = NEW.size
    ) THEN
        RAISE EXCEPTION 'Invalid animal size: %', NEW.size;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to validate size before insert or update
CREATE TRIGGER validate_animal_size_trigger
    BEFORE INSERT OR UPDATE ON animals
    FOR EACH ROW
    EXECUTE FUNCTION validate_animal_size(); 