import requests
import json
import os

BASE_URL = "http://localhost:8080"

class BinTraceBenchClient:
    def __init__(self, base_url=BASE_URL):
        self.base_url = base_url
        self.token = None
        
    def register(self, username, password, email):
        """Register a new user"""
        print(f"🔑 Registering user: {username}")
        data = {
            "username": username,
            "password": password,
            "email": email
        }
        res = requests.post(f"{self.base_url}/auth/register", json=data)
        print("Status:", res.status_code)
        try:
            response_data = res.json()
            print("Response:", json.dumps(response_data, indent=2))
            return res.status_code == 201
        except:
            print("Raw Response:", res.text[:500])
            return False
    
    def login(self, username, password):
        """Login and store token"""
        print(f"🔑 Logging in user: {username}")
        data = {
            "username": username,
            "password": password
        }
        res = requests.post(f"{self.base_url}/auth/login", json=data)
        print("Status:", res.status_code)
        try:
            response_data = res.json()
            print("Response:", json.dumps(response_data, indent=2))
            if res.status_code == 200:
                self.token = response_data.get("token")
                print(f"✅ Logged in successfully. Token: {self.token[:20]}...")
                return True
        except:
            print("Raw Response:", res.text[:500])
        return False
    
    def get_headers(self):
        """Get headers with authentication"""
        if not self.token:
            raise Exception("Not authenticated. Call login() first.")
        return {"Authorization": f"Bearer {self.token}"}
    
    def upload_file(self, endpoint: str, binary_path: str, params=None):
        """Upload file to endpoint with authentication"""
        print(f"🔹 Testing {endpoint} with multipart upload")
        try:
            with open(binary_path, "rb") as f:
                files = {'file': (binary_path.split('/')[-1], f)}
                res = requests.post(
                    f"{self.base_url}{endpoint}", 
                    files=files, 
                    params=params,
                    headers=self.get_headers()
                )
            print("Status:", res.status_code)
            try:
                response_data = res.json()
                print("Response:", json.dumps(response_data, indent=2))
                return response_data
            except:
                print("Raw Response:", res.text[:500])
        except Exception as e:
            print(f"Error: {e}")
    
    def get_results(self, endpoint: str):
        """Get stored results"""
        print(f"📋 Getting results from {endpoint}")
        try:
            res = requests.get(f"{self.base_url}{endpoint}", headers=self.get_headers())
            print("Status:", res.status_code)
            try:
                response_data = res.json()
                print("Response:", json.dumps(response_data, indent=2))
                return response_data
            except:
                print("Raw Response:", res.text[:500])
        except Exception as e:
            print(f"Error: {e}")
    
    def me(self):
        """Get current user info"""
        print("👤 Getting current user info")
        try:
            res = requests.get(f"{self.base_url}/auth/me", headers=self.get_headers())
            print("Status:", res.status_code)
            try:
                response_data = res.json()
                print("Response:", json.dumps(response_data, indent=2))
                return response_data
            except:
                print("Raw Response:", res.text[:500])
        except Exception as e:
            print(f"Error: {e}")

def test_all(binary_path):
    client = BinTraceBenchClient()
    
    # Test user registration and login
    username = "testuser"
    password = "testpass123"
    email = "test@example.com"
    
    print("=" * 60)
    print("🧪 Testing Authentication")
    print("=" * 60)
    
    # Register user (might fail if user already exists)
    client.register(username, password, email)
    
    # Login
    if not client.login(username, password):
        print("❌ Login failed, cannot continue with authenticated tests")
        return
    
    # Get user info
    client.me()
    
    print("\n" + "=" * 60)
    print("🧪 Testing Binary Analysis")
    print("=" * 60)
    
    # Test static analysis
    client.upload_file("/analyze", binary_path)
    
    # Test dynamic analysis
    client.upload_file("/analyze", binary_path, params={"dynamic": "true"})
    
    # Get analysis results
    client.get_results("/analyze")
    
    print("\n" + "=" * 60)
    print("🧪 Testing Benchmarking")
    print("=" * 60)
    
    # Test benchmark without trace
    client.upload_file("/bench", binary_path)
    
    # Test benchmark with trace
    client.upload_file("/bench", binary_path, params={"trace": "true"})
    
    # Get benchmark results
    client.get_results("/bench")

if __name__ == "__main__":
    # Use a simple binary that should exist on most systems
    test_binary = "/bin/ls"
    if not os.path.exists(test_binary):
        test_binary = "/usr/bin/ls"
    if not os.path.exists(test_binary):
        print("❌ Could not find test binary (/bin/ls or /usr/bin/ls)")
        exit(1)
    
    print(f"🔬 Testing with binary: {test_binary}")
    test_all(test_binary)
