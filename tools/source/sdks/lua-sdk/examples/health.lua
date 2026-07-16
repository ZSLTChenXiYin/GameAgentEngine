local Client = require("../src/client")
local client = Client.new("http://127.0.0.1:8080", "dev-key")
print(client:health_path())
