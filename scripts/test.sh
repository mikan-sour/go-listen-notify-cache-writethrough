export $(grep -v '^#' .env | xargs)

RecordId=$(PGPASSWORD=$DB_PASSWORD psql -U $DB_USERNAME $DB_NAME -h $DB_HOST -p $DB_PORT -qtAX -c 'INSERT INTO records (created_date) values(now()) RETURNING id')
redis-cli -h $CACHE_HOST -a $CACHE_PASSWORD --no-auth-warning get $RecordId