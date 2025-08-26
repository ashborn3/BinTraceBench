import requests
import json

BASE_URL = "http://localhost:8080"

def upload_file(endpoint: str, binary_path: str, params=None):
    print(f"🔹 Testing {endpoint} with multipart upload")
    with open(binary_path, "rb") as f:
        files = {'file': (binary_path.split('/')[-1], f)}
        res = requests.post(f"{BASE_URL}{endpoint}", files=files, params=params)
    print("Status:", res.status_code)
    try:
        print("Response:", json.dumps(res.json(), indent=2))
    except:
        print("Raw Response:", res.text[:500])

def test_all(binary_path):
    upload_file("/analyze", binary_path)
    upload_file("/analyze", binary_path, params={"dynamic": "true"})
    upload_file("/bench", binary_path)

if __name__ == "__main__":
    test_all("/bin/ls")
