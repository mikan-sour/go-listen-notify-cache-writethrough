# Using Postgres Listen & Notify for Cache Writethrough

## Utilizing Postgres native pub/sub features 
Write to the cache with the DB and a listener instead of your API, improving latency for your client

## To run
1. set up your env
> make env
2. run the containers for Postgres and Redis
> make up
3. run the app
> make run
4. do an insert
> make one

The above command will insert a record into the Postgres DB and print the value of the key set in the cache
You can also open a postgres client and insert yourself.

## Tests
### Unit
> make test

### Integration
> make e2e
