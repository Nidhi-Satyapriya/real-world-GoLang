"""
Collector API - Device Posture Report Receiver
Receives and processes device health reports from the Go agent
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional
import uvicorn
import json

app = FastAPI(
    title="Device Posture Collector API",
    description="Receives and processes device health reports",
    version="1.0.0"
)

# In-memory storage for reports (in production, use a database)
reports_db = []


class DeviceStatus(BaseModel):
    """Device status model matching the Go agent's data structure"""
    hostname: str = Field(..., description="Device hostname")
    ip: str = Field(..., description="Device IP address")
    disk_usage: float = Field(..., description="Disk usage percentage", ge=0, le=100)
    status: str = Field(..., description="Health status (HEALTHY/UNHEALTHY)")
    timestamp: datetime = Field(..., description="Report timestamp")
    message: Optional[str] = Field(None, description="Additional status message")


@app.get("/")
def root():
    """Root endpoint with API information"""
    return {
        "service": "Device Posture Collector API",
        "version": "1.0.0",
        "endpoints": {
            "/report": "POST - Submit device status report",
            "/reports": "GET - View all received reports",
            "/reports/unhealthy": "GET - View unhealthy devices",
            "/health": "GET - API health check"
        }
    }


@app.post("/report")
def receive_report(data: DeviceStatus):
    """
    Receive and process device status report
    
    Args:
        data: DeviceStatus object containing device health information
    
    Returns:
        Confirmation message with any alerts
    """
    # Store the report
    report_dict = data.model_dump()
    reports_db.append(report_dict)
    
    # Check if device is unhealthy
    if data.status == "UNHEALTHY":
        alert_msg = (
            f"🚨 ALERT: Device {data.hostname} is CRITICAL!\n"
            f"   IP: {data.ip}\n"
            f"   Disk Usage: {data.disk_usage:.2f}%\n"
            f"   Message: {data.message}\n"
            f"   Timestamp: {data.timestamp}"
        )
        print(alert_msg)
        
        return {
            "msg": "Report received - UNHEALTHY device detected",
            "alert": True,
            "device": data.hostname,
            "action_required": "Immediate attention needed"
        }
    
    # Healthy device
    print(f"✓ Report received from {data.hostname} ({data.ip}) - Status: {data.status}")
    
    return {
        "msg": "Report received successfully",
        "alert": False,
        "device": data.hostname,
        "status": data.status
    }


@app.get("/reports")
def get_all_reports(limit: int = 50):
    """
    Get all received reports
    
    Args:
        limit: Maximum number of reports to return (default: 50)
    
    Returns:
        List of device reports
    """
    return {
        "total_reports": len(reports_db),
        "reports": reports_db[-limit:] if limit > 0 else reports_db
    }


@app.get("/reports/unhealthy")
def get_unhealthy_reports():
    """
    Get all reports from unhealthy devices
    
    Returns:
        List of unhealthy device reports
    """
    unhealthy = [r for r in reports_db if r.get("status") == "UNHEALTHY"]
    return {
        "total_unhealthy": len(unhealthy),
        "unhealthy_devices": unhealthy
    }


@app.get("/reports/{hostname}")
def get_device_reports(hostname: str, limit: int = 10):
    """
    Get reports for a specific device
    
    Args:
        hostname: Device hostname to filter by
        limit: Maximum number of reports to return
    
    Returns:
        List of reports for the specified device
    """
    device_reports = [r for r in reports_db if r.get("hostname") == hostname]
    
    if not device_reports:
        raise HTTPException(status_code=404, detail=f"No reports found for device: {hostname}")
    
    return {
        "hostname": hostname,
        "total_reports": len(device_reports),
        "reports": device_reports[-limit:]
    }


@app.get("/health")
def health_check():
    """API health check endpoint"""
    return {
        "status": "healthy",
        "service": "collector-api",
        "timestamp": datetime.now().isoformat(),
        "reports_received": len(reports_db)
    }


@app.delete("/reports")
def clear_reports():
    """Clear all stored reports (use with caution)"""
    count = len(reports_db)
    reports_db.clear()
    return {
        "msg": f"Cleared {count} reports",
        "reports_remaining": len(reports_db)
    }


if __name__ == "__main__":
    print("""
╔═══════════════════════════════════════════════════╗
║     📡 COLLECTOR API v1.0 📡                     ║
║     Device Posture Report Receiver               ║
╚═══════════════════════════════════════════════════╝

🚀 Starting FastAPI server...
📍 Listening on: http://localhost:8000
📝 API Docs: http://localhost:8000/docs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
""")
    
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8000,
        log_level="info"
    )
