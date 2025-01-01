from pydantic_settings import BaseSettings
import os

class Settings(BaseSettings):
    token: str

    class Config:
        env_file = ".env"

# Check if .env file exists, if not, create it with a default token
env_file_path = ".env"
if not os.path.exists(env_file_path):
    with open(env_file_path, "w") as file:
        file.write("TOKEN=default_secret_token_here\n")

settings = Settings()
