import requests
import os

class Client:
    def __init__(self):
        self.base_url = "http://localhost:8080"
        self.token = None
        
    def register(self, username, password, email):
        res = requests.post(f"{self.base_url}/auth/register", json={
            "username": username, "password": password, "email": email
        })
        return res.status_code == 201
    
    def login(self, username, password):
        res = requests.post(f"{self.base_url}/auth/login", json={
            "username": username, "password": password
        })
        if res.status_code == 200:
            self.token = res.json().get("token")
            return True
        return False
    
    def upload_file(self, endpoint, binary_path, params=None):
        with open(binary_path, "rb") as f:
            files = {'file': (binary_path.split('/')[-1], f)}
            res = requests.post(
                f"{self.base_url}{endpoint}", 
                files=files, 
                params=params,
                headers={"Authorization": f"Bearer {self.token}"}
            )
        return res.status_code, res.text
    
    def get_results(self, endpoint):
        res = requests.get(f"{self.base_url}{endpoint}", 
                          headers={"Authorization": f"Bearer {self.token}"})
        return res.status_code, res.text

def test():
    client = Client()
    client.register("testuser", "testpass123", "test@example.com")
    
    if not client.login("testuser", "testpass123"):
        print("Login failed")
        return
    
    binary = "/bin/ls" if os.path.exists("/bin/ls") else "/usr/bin/ls"
    
    print("Static analysis:", client.upload_file("/analyze", binary)[0])
    print("Dynamic analysis:", client.upload_file("/analyze", binary, {"dynamic": "true"})[0])
    print("Benchmark:", client.upload_file("/bench", binary)[0])
    print("Benchmark with trace:", client.upload_file("/bench", binary, {"trace": "true"})[0])

if __name__ == "__main__":
    test()
