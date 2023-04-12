# Mail IT
## Send emails with a POST Request

This is an application that sends emails :)

To request an email send a `POST` request to `/email` with the following body structure

```
{
  "to": [
    {
      "name": "June",
      "address": "june@doe.com"
    }
  ],
  "cc": [
    {
      "name": "James",
      "address": "james@doe.com"
    }
  ],
  "subject": "Subject",
  "body": "email body",
  "attachments": [
    "https://go.dev/images/learn/go-programming-blueprints.png"
  ]
}
```