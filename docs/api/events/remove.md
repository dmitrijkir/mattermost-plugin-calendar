# Remove event by id

## Parameters

| name       | type     | data type | description | example                                 |
|------------|----------|-----------|-------------|-----------------------------------------|
| eventId    | required | string    | N/A         | "a8639bf2-9467-44b9-b797-7bf1004d2ffc"  |


## Response Event Object

| name       | type       | data type | description | example |
|------------|------------|-----------|-------------|--------|
| success    | required   | bool      | N/A         | true   |


## Example cURL

```javascript
  curl --request DELETE 'http://localhost:8065/plugins/com.dmkir.calendar/event?eventId=316c4857-def9-4fe9-afd1-7b13308d65a7'
 ```


## Example response

 ```json
{
  "data": {
    "success": true
  }
}
```

