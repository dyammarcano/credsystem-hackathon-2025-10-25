import requests
import json
response = requests.get(
  url="https://openrouter.ai/api/v1/key",
  headers={
    "Authorization": f"Bearer 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"
  }
)
print(json.dumps(response.json(), indent=2))
