import os
import sys
from datetime import datetime, timedelta, timezone

import jwt
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="Zero Trust IdP", version="1.0.0")

JWT_SECRET = os.environ.get("JWT_SECRET")
if not JWT_SECRET:
    print("FATAL: JWT_SECRET environment variable is required", file=sys.stderr)
    sys.exit(1)

ALGORITHM = "HS256"
TOKEN_EXPIRY_MINUTES = 30

USERS = {
    "admin": {"password": "admin123", "role": "admin", "name": "Alice Admin"},
    "guest": {"password": "guest123", "role": "guest", "name": "Bob Guest"},
}


class LoginRequest(BaseModel):
    username: str
    password: str


class TokenResponse(BaseModel):
    access_token: str
    token_type: str
    expires_in: int
    role: str


@app.post("/login", response_model=TokenResponse)
def login(req: LoginRequest):
    user = USERS.get(req.username)
    if not user or user["password"] != req.password:
        raise HTTPException(status_code=401, detail="Invalid username or password")

    now = datetime.now(timezone.utc)
    payload = {
        "sub": req.username,
        "role": user["role"],
        "name": user["name"],
        "iat": now,
        "exp": now + timedelta(minutes=TOKEN_EXPIRY_MINUTES),
    }

    token = jwt.encode(payload, JWT_SECRET, algorithm=ALGORITHM)

    return TokenResponse(
        access_token=token,
        token_type="bearer",
        expires_in=TOKEN_EXPIRY_MINUTES * 60,
        role=user["role"],
    )


@app.get("/health")
def health():
    return {"status": "ok", "service": "idp"}
