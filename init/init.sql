CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
    IF NOT EXISTS records (
        id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
        created_date timestamp without time zone NOT NULL DEFAULT (
            current_timestamp AT TIME ZONE 'UTC'
        )
    );
CREATE OR REPLACE FUNCTION NOTIFY_ROW_INSERTED() RETURNS 
TRIGGER AS 
	$$ DECLARE BEGIN PERFORM pg_notify(
	    CAST('row_inserted' AS text),
	    row_to_json(NEW):: text
	);
	RETURN NEW;
END; 
$$ LANGUAGE plpgsql;

CREATE TRIGGER NOTIFY_ROW_INSERTED 
	AFTER
	INSERT ON records FOR EACH ROW
	EXECUTE
	    PROCEDURE notify_row_inserted();

LISTEN row_inserted;


