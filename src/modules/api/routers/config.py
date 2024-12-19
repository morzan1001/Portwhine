#!/usr/bin/env python3
from fastapi import APIRouter

router = APIRouter()

@router.get("/test")
async def read_test():
    return {"message": "This is a test endpoint"}