# Get events

## Parameters

| name  | type     | data type | description | where       | example             |
|-------|----------|-----------|-------------|-------------|---------------------|
| start | required | datetime  | N/A         | Querystring | 2023-01-30T00:00:00 |
| end   | required | datetime  | N/A         | Querystring | 2023-02-06T00:00:00 |

## Response list of event object

| name       | type     | data type | description | example                                |
|------------|----------|-----------|-------------|----------------------------------------|
| id         | required | string    | N/A         | "a8639bf2-9467-44b9-b797-7bf1004d2ffc" |
| title      | required | string    | N/A         | new event                              |
| start      | required | datetime  | N/A         | 2023-01-28T00:30:00Z                   |
| end        | required | datetime  | N/A         | 2023-01-28T01:00:00Z                   |
| attendees  | optional | []string  | N/A         | ["sh9d5kji7tf49echstq79dm36r",]        |
| channel    | optional | string    | N/A         | 516netffp7dgxx6denw6tbk9br             |
| recurrence | required | string    | N/A         | ""                                     |
| created    | required | datetime  | N/A         | 2023-01-28T20:09:40.829475047Z         |
| owner      | required | string    | N/A         | sh9d5kji7tf49echstq79dm36r             |
| color      | optional | string    | N/A         | #D0D0D0                                |
| team       | optional | string    | N/A         | 516netffp7dgxx6denw6tbk9br             |
| alert      | optional | string    | N/A         | 5_minutes_before                       |
| visibility | optional | string    | N/A         | private                                |

## Example cURL

```javascript
  curl
'http://localhost:8065/plugins/com.dmkir.calendar/events?start=2023-01-30T00%3A00%3A00&end=2023-02-06T00%3A00%3A00'
 ```

## Example response

 ```json
{
  "data": [
    {
      "id": "a8639bf2-9467-44b9-b797-7bf1004d2ffc",
      "title": "new event",
      "start": "2023-01-27T20:30:00Z",
      "end": "2023-01-27T21:00:00Z",
      "attendees": [
        "sh9d5kji7tf49echstq79dm36r"
      ],
      "created": "2023-01-28T20:09:40.829475047Z",
      "owner": "sh9d5kji7tf49echstq79dm36r",
      "channel": "516netffp7dgxx6denw6tbk9br",
      "recurrence": ""
    }
  ]
}
```

