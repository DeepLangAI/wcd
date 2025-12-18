import uvicorn
from pydantic import BaseModel
import requests
from fastapi import FastAPI
from utils import R, simple_crawl
from typing import Optional

import argparse

app = FastAPI()

class CrawlRequest(BaseModel):
    url: str
    force_browser: Optional[bool] = False

@app.post("/crawl")
async def crawler(req: CrawlRequest):
    url = req.url
    html = simple_crawl(url)
    if html:
        data = {
            "url": url,
            "html": html,
        }
        return R.success(data)
    else:
        return R.error(1, f"Failed to crawl {url}")


def build_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", type=str, default="0.0.0.0")
    parser.add_argument("--port", type=int, default=18100)
    parser.add_argument("--reload", action="store_true", default=False)
    return parser.parse_args()

if __name__ == "__main__":
    args = build_args()
    uvicorn.run('server_crawler:app', host=args.host, port=args.port, reload=args.reload)
