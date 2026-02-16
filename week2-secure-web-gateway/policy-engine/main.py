"""
FastAPI Policy Engine for Secure Web Gateway
Serves blocklist policies to the Go proxy server
"""

from fastapi import FastAPI
from fastapi.responses import JSONResponse
from typing import List
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Cisco SWG Policy Engine",
    description="Policy management API for Secure Web Gateway",
    version="1.0.0"
)

# In-memory blocklist (in production, this would come from a database)
BLOCKED_DOMAINS = [
    "facebook.com",
    "tiktok.com",
    "twitter.com",
    "instagram.com",
    "reddit.com",
    "youtube.com",
    "gambling.com",
    "bet365.com",
    "pokerstars.com"
]


@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "service": "Cisco SWG Policy Engine",
        "status": "running",
        "version": "1.0.0"
    }


@app.get("/policy")
async def get_policy():
    """
    Returns the current blocklist policy
    
    Returns:
        JSON object with blocked domains list
    """
    logger.info(f"Policy requested - returning {len(BLOCKED_DOMAINS)} blocked domains")
    
    return JSONResponse(
        content={
            "blocked": BLOCKED_DOMAINS,
            "total": len(BLOCKED_DOMAINS),
            "last_updated": "2026-02-16T00:00:00Z"
        }
    )


@app.post("/policy/add")
async def add_domain(domain: str):
    """
    Add a domain to the blocklist
    
    Args:
        domain: Domain name to block (e.g., "example.com")
    """
    domain = domain.lower().strip()
    
    if domain in BLOCKED_DOMAINS:
        return {
            "status": "already_exists",
            "domain": domain,
            "message": f"{domain} is already in the blocklist"
        }
    
    BLOCKED_DOMAINS.append(domain)
    logger.info(f"Added domain to blocklist: {domain}")
    
    return {
        "status": "added",
        "domain": domain,
        "total_blocked": len(BLOCKED_DOMAINS)
    }


@app.delete("/policy/remove")
async def remove_domain(domain: str):
    """
    Remove a domain from the blocklist
    
    Args:
        domain: Domain name to unblock (e.g., "example.com")
    """
    domain = domain.lower().strip()
    
    if domain not in BLOCKED_DOMAINS:
        return {
            "status": "not_found",
            "domain": domain,
            "message": f"{domain} is not in the blocklist"
        }
    
    BLOCKED_DOMAINS.remove(domain)
    logger.info(f"Removed domain from blocklist: {domain}")
    
    return {
        "status": "removed",
        "domain": domain,
        "total_blocked": len(BLOCKED_DOMAINS)
    }


@app.get("/policy/domains")
async def list_domains():
    """
    List all blocked domains
    
    Returns:
        List of all blocked domains with metadata
    """
    return {
        "domains": sorted(BLOCKED_DOMAINS),
        "total": len(BLOCKED_DOMAINS)
    }


@app.get("/health")
async def health_check():
    """Kubernetes/container health check endpoint"""
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn
    
    logger.info("Starting Cisco SWG Policy Engine...")
    logger.info(f"Loaded {len(BLOCKED_DOMAINS)} blocked domains")
    
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8000,
        log_level="info"
    )
