# Get user settings

## Response settings object

| name                  | type     | data type | description | example                        |
|-----------------------|----------|-----------|-------------|--------------------------------|
| businessDays          | required | []int     | N/A         | [1, 3]                         |
| businessStartTime     | required | string    | N/A         | "09:00"                        |
| businessEndTime       | required | string    | N/A         | "19:00"                        |
| firstDayOfWeek        | required | int       | N/A         | 1                              |
| hideNonWorkingDays    | required | boolean   | N/A         | true                           |
| isOpenCalendarLeftBar | required | boolean   | N/A         | true                           |

## Example cURL

```javascript
  curl
'http://localhost:8065/plugins/com.dmkir.calendar/settings'
 ```

## Example response

 ```json
{
  "data": {
    "businessStartTime": "09:00",
    "businessEndTime": "19:00",
    "isOpenCalendarLeftBar": true,
    "firstDayOfWeek": 1,
    "businessDays": [
      1,
      2,
      3,
      4,
      5
    ],
    "hideNonWorkingDays": true
  }
}
```

