# Creating new event

## Parameters

| name       | type     | data type | description | example                         |
|------------|----------|-----------|-------------|---------------------------------|
| title      | required | string    | N/A         | new event                       |
| start      | required | datetime  | N/A         | 2023-01-28T00:30:00Z            |
| end        | required | datetime  | N/A         | 2023-01-28T01:00:00Z            |
| attendees  | optional | []string  | N/A         | ["sh9d5kji7tf49echstq79dm36r",] |
| channel    | optional | string    | N/A         | 516netffp7dgxx6denw6tbk9br      |
| recurrence | optional | string    | N/A         | ""                              |
| team       | optional | string    | N/A         | 516netffp7dgxx6denw6tbk9br      |
| alert      | optional | string    | N/A         | 5_minutes_before                |
| visibility | optional | string    | N/A         | private                         |
| color      | optional | string    | N/A         | #D0D0D0                         |

## Response Event Object

| name       | type     | data type | description | example                                |
|------------|----------|-----------|-------------|----------------------------------------|
| id         | required | string    | N/A         | "a8639bf2-9467-44b9-b797-7bf1004d2ffc" |
| title      | required | string    | N/A         | new event                              |
| start      | required | datetime  | N/A         | 2023-01-28T00:30:00Z                   |
| end        | required | datetime  | N/A         | 2023-01-28T01:00:00Z                   |
| attendees  | optional | []string  | N/A         | ["sh9d5kji7tf49echstq79dm36r",]        |
| channel    | optional | string    | N/A         | 516netffp7dgxx6denw6tbk9br             |
| recurrence | optional | string    | N/A         | ""                                     |
| created    | required | datetime  | N/A         | 2023-01-28T20:09:40.829475047Z         |
| owner      | required | string    | N/A         | sh9d5kji7tf49echstq79dm36r             |
| team       | optional | string    | N/A         | 516netffp7dgxx6denw6tbk9br             |
| alert      | optional | string    | N/A         | 5_minutes_before                       |
| visibility | optional | string    | N/A         | private                                |
| color      | optional | string    | N/A         | #D0D0D0                                |

## Example cURL

```javascript
  curl
'http://localhost:8065/plugins/com.dmkir.calendar/events' \
 --data - raw
'{"title":"new event","start":"2023-01-28T00:30:00Z","end":"2023-01-28T01:00:00Z","attendees":["sh9d5kji7tf49echstq79dm36r"],"channel":"516netffp7dgxx6denw6tbk9br","recurrence":""}'
--compressed
 ```

## Example response

 ```json
{
  "data": {
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
}
```

