# Token_Transfer_API

GraphQL API with tansfer mutation, connects to PostgreSQL database.

API handles race conditions. Wallet balances cannot go negative.

## How to run

### App

To build and start containers first make sure to open 
Docker Desktop. Then open main project directory 
in termianal and use:
```
docker compose -f docker-compose.yml up --build
```

After local changes restart the container. 
First make sure to open Docker Desktop. Then use:
```
docker compose restart app
```

### Tests

To run tests make sure to start containers first. 
Then open main project directory and use:
```
docker compose exec app go test -v ./tests
```

## Database

To open PostgreSQL database make sure to build and run containers first. 
Then open project main directory in terminal and use:
```
docker compose exec db psql -U myuser -d mydb
```

## GraphQL playground

To use GraphQL playground run the app and open 
`http://localhost:8080`.
On the left panel create mutation, for example:
```
mutation {
  transfer(
    from_address: "address_1", 
    to_address: "address_2", 
    amount: 5
  )
}
```
For sender wallet make sure to use address that exist in database. 
If receiver wallet does not exist it will be created 
with initial balance equal 0.
If transfer completes successfully the playground will 
show updated balance of sender wallet, for example:
```
{
  "data": {
    "transfer": 95
  }
}
```