import requests
import os
import json

class Client:
    def __init__(self):
        self.base_url = "http://localhost:8080"
        self.token = None
        self.results = {}
        
    def save_response(self, name, response):
        try:
            data = response.json()
        except:
            data = {"raw_text": response.text, "status_code": response.status_code}
        
        self.results[name] = data
        with open(f"{name}.json", "w") as f:
            json.dump(data, f, indent=2)
        
    def register(self, username, password, email):
        res = requests.post(f"{self.base_url}/auth/register", json={
            "username": username, "password": password, "email": email
        })
        self.save_response("register", res)
        return res.status_code == 201
    
    def login(self, username, password):
        res = requests.post(f"{self.base_url}/auth/login", json={
            "username": username, "password": password
        })
        self.save_response("login", res)
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
        
        name = endpoint.replace("/", "") + ("_dynamic" if params and "dynamic" in params else "") + ("_trace" if params and "trace" in params else "")
        self.save_response(name, res)
        return res.status_code, res.text
    
    def get_results(self, endpoint):
        res = requests.get(f"{self.base_url}{endpoint}", 
                          headers={"Authorization": f"Bearer {self.token}"})
        name = endpoint.replace("/", "") + "_results"
        self.save_response(name, res)
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
    
    with open("all_responses.json", "w") as f:
        json.dump(client.results, f, indent=2)

if __name__ == "__main__":
    test()
