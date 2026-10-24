DO $$
BEGIN
    BEGIN
        DROP EXTENSION IF EXISTS pgcrypto;
    EXCEPTION
        WHEN dependent_objects_still_exist THEN
            NULL;
    END;
END;
$$;