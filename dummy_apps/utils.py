import requests

class R:
    @staticmethod
    def success(data):
        return {
            "code": 0,
            "msg": "success",
            "data": data
        }

    @staticmethod
    def error(code, msg):
        return {
            "code": code,
            "msg": msg,
            "data": None
        }

def simple_crawl(url: str):
    cookies = {
        '_ga': 'GA1.1.371080025.1728356883',
        '_c_WBKFRo': 'l9lxL8bgxh8LBycHjH44gRiG24B97xs330NCIcu2',
        'sensorsdata2015jssdkcross': '%7B%22distinct_id%22%3A%22cbba116dd08a48238a674e7ce3350637%22%2C%22first_id%22%3A%221926a1911d01d5e-0592b796ee5f188-16525637-1484784-1926a1911d12467%22%2C%22props%22%3A%7B%22%24latest_traffic_source_type%22%3A%22%E7%9B%B4%E6%8E%A5%E6%B5%81%E9%87%8F%22%2C%22%24latest_search_keyword%22%3A%22%E6%9C%AA%E5%8F%96%E5%88%B0%E5%80%BC_%E7%9B%B4%E6%8E%A5%E6%89%93%E5%BC%80%22%2C%22%24latest_referrer%22%3A%22%22%7D%2C%22identities%22%3A%22eyIkaWRlbnRpdHlfY29va2llX2lkIjoiMTkyNmExOTExZDAxZDVlLTA1OTJiNzk2ZWU1ZjE4OC0xNjUyNTYzNy0xNDg0Nzg0LTE5MjZhMTkxMWQxMjQ2NyIsIiRpZGVudGl0eV9sb2dpbl9pZCI6ImNiYmExMTZkZDA4YTQ4MjM4YTY3NGU3Y2UzMzUwNjM3In0%3D%22%2C%22history_login_id%22%3A%7B%22name%22%3A%22%24identity_login_id%22%2C%22value%22%3A%22cbba116dd08a48238a674e7ce3350637%22%7D%2C%22%24device_id%22%3A%221926a1911d01d5e-0592b796ee5f188-16525637-1484784-1926a1911d12467%22%7D',
        '_ga_QZXRQK8759': 'GS2.1.s1753779436$o3$g0$t1753779439$j57$l0$h0',
        '_ga_5DW4TZD93L': 'GS2.1.s1753779442$o1$g1$t1753779453$j49$l0$h0',
        'ahoy_visitor': 'ac6ccb70-932d-4b50-bac2-aa30d3d3e262',
        '_ga_R78HWX068N': 'GS2.1.s1763027025$o21$g0$t1763027025$j60$l0$h0',
    }

    headers = {
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
        'Accept-Language': 'zh,en-US;q=0.9,en;q=0.8,zh-CN;q=0.7,zh-TW;q=0.6',
        'Connection': 'keep-alive',
        'If-None-Match': 'W/"e46d30f5935a4ed8b7f346d093b32fd8"',
        'Sec-Fetch-Dest': 'document',
        'Sec-Fetch-Mode': 'navigate',
        'Sec-Fetch-Site': 'none',
        'Sec-Fetch-User': '?1',
        'Upgrade-Insecure-Requests': '1',
        'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36',
        'channel': 'local',
        'sec-ch-ua': '"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"',
        'sec-ch-ua-mobile': '?0',
        'sec-ch-ua-platform': '"macOS"',
        # 'Cookie': '_ga=GA1.1.371080025.1728356883; _c_WBKFRo=l9lxL8bgxh8LBycHjH44gRiG24B97xs330NCIcu2; sensorsdata2015jssdkcross=%7B%22distinct_id%22%3A%22cbba116dd08a48238a674e7ce3350637%22%2C%22first_id%22%3A%221926a1911d01d5e-0592b796ee5f188-16525637-1484784-1926a1911d12467%22%2C%22props%22%3A%7B%22%24latest_traffic_source_type%22%3A%22%E7%9B%B4%E6%8E%A5%E6%B5%81%E9%87%8F%22%2C%22%24latest_search_keyword%22%3A%22%E6%9C%AA%E5%8F%96%E5%88%B0%E5%80%BC_%E7%9B%B4%E6%8E%A5%E6%89%93%E5%BC%80%22%2C%22%24latest_referrer%22%3A%22%22%7D%2C%22identities%22%3A%22eyIkaWRlbnRpdHlfY29va2llX2lkIjoiMTkyNmExOTExZDAxZDVlLTA1OTJiNzk2ZWU1ZjE4OC0xNjUyNTYzNy0xNDg0Nzg0LTE5MjZhMTkxMWQxMjQ2NyIsIiRpZGVudGl0eV9sb2dpbl9pZCI6ImNiYmExMTZkZDA4YTQ4MjM4YTY3NGU3Y2UzMzUwNjM3In0%3D%22%2C%22history_login_id%22%3A%7B%22name%22%3A%22%24identity_login_id%22%2C%22value%22%3A%22cbba116dd08a48238a674e7ce3350637%22%7D%2C%22%24device_id%22%3A%221926a1911d01d5e-0592b796ee5f188-16525637-1484784-1926a1911d12467%22%7D; _ga_QZXRQK8759=GS2.1.s1753779436$o3$g0$t1753779439$j57$l0$h0; _ga_5DW4TZD93L=GS2.1.s1753779442$o1$g1$t1753779453$j49$l0$h0; ahoy_visitor=ac6ccb70-932d-4b50-bac2-aa30d3d3e262; _ga_R78HWX068N=GS2.1.s1763027025$o21$g0$t1763027025$j60$l0$h0',
    }
    response = requests.get(url, headers=headers)
    if response.status_code == 200:
        return response.text
    else:
        return None
