# Push service mock implementation

URL: `/push`<br>
method: `POST`<br>
payload encoding: `application/json`

```json
{
    "id": 1231,
    "push_service": "marketing",
    "title": "20% discount",
    "text": "Only today discount",
    "data": {
        "campaign_id": "38"
    },
    "ttl": 0,
    "tokens": [
        "abcmdsfdsfdsfkj435435.324453",
        "anbvf7vf665a76vnmrwero43/435345.45"
    ],
    "validate_only": false
}

```

## Push v2

URL: `/push/v2`<br>
method: `POST`<br>
payload encoding: `application/json`

```json
{
    "id": 213123,
    "push_source": "marketing",
    "raw_msg": {},
    "tokens": [
        "abcmdsfdsfdsfkj435435.324453",
        "anbvf7vf665a76vnmrwero43/435345.45"
    ],
    "validate_only": false
}
```

Comments:
- `data` can be any valid json
- `data` should not contain token list for minimize payload size