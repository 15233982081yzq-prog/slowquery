wrk.method = "GET"
wrk.headers["Content-Type"] = "application/json"
baseUrl = "http://10.129.120.132:30083/rds/smart/v1/api/non-live/platform/database/hosts"


local databases = {"test1","test2","test3"}
local auth = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImI3OTBlYjE5MmY2YTJkMTZjNDQwZWE3NGRlODU4ODg1NDc5NjZmZDEiLCJ0eXAiOiJKV1QifQ.eyJlbWFpbCI6ImppYW4uYmlhbkBzaG9wZWUuY29tIiwiZXhwIjoxNjkzNzkzNTAyLCJpc3MiOiJodHRwczovL3NwYWNlLnRlc3Quc2hvcGVlLmlvL3YxL2NlcnRzIiwic3ViIjoiMzE5NjExMzMifQ.GkYn889_YLY5bgUZ0ctrE-gzDJ7vrqYf106fdQWPlAA0TaGqneMV2SOMpSpInXra4bnIDHh3lwPdwfmsPnLa1d3uld5ZqRW830UFllNpiuvnTp3iKdT78WuJqYatg1akVg2Eo--rX-NE18K1vqX2cgIV5_B1W0tfcHNco-OyZHLrl_zhs6mU3tTk-pbjq5CCCpiGqJYq5Hg81LkTMKy1pYZbH24OLLn3HUMlIBDqrliwkWyB9SE8OSDpcw7xFzWl9ELy8l8ezkmXNuq8vliWQVFTFeK509ip_r9wnXr4QpaARoVCvRctr9yhEJBsJFQpEv0ChkhDlV72DThTZ_IAjxTYzKXabIUAVSvypfA9xbIEn5O9n06PGsYBKHaYHauKhjxQRPTbZKQ54txaa4bNB8LZncMwEFPzD7NFdYU0Xoo5PxkiU39fnOqc6nj7k9XHkjqjkw_L-_qf9KKoJ181jklwipgihDyeImilvY4V776tB3HLyOQKwgQnQr77_voi6B7WZDRWtSb-9LnCucnXoS0YwiXaI8h2C4jhNJfC2CN_JsYVox0SIQNzcgveG9l4Y2-PsPOsdsY0Jrl0eYaYeQKkGvXctxM7ffraU7NEjh-Abx4EvuP1C5YSA5x5XY0yftfCiUWCFcRxKeMhEB8L6aonPLGr_lZrxylROz1rYqg"
local env = "pfb"
local params = {}

function request()
    wrk.headers["Authorization"] = "Bearer " .. auth
    params = build_query_map()
    queryStr = build_query_string(params)
    local full_url = baseUrl .. "?" .. queryStr-- 去除最后一个"&"

    -- print("Generated Database_hosts URL: " .. full_url)  -- 打印生成的URL
    return wrk.format(nil, full_url)
end



function build_query_map()
    local start_timestamp = build_start_timestamp()
    local end_timestamp = random_time_range(start_timestamp)

    params["db_name"] = random_database(databases)
    params["db_env"] = env
    params["start_time"] = start_timestamp
    params["end_time"] = end_timestamp

    return params
end

-- 随机选择数据库函数
function random_database(databases)
    local random_index = math.random(#databases)
    return databases[random_index]
end

-- 请求参数拼装函数
function build_query_string(params)
    local query_string = ""
    for key, value in pairs(params) do
        query_string = query_string .. key .. "=" .. value .. "&"
    end
    return query_string:sub(1, -2) -- 去除最后一个"&"
end


-- 获取当前日期和时间
local current_time = os.time()

-- 获取当前日期的年、月、日部分
local current_date_table = os.date("*t", current_time)
local current_year = current_date_table.year
local current_month = current_date_table.month
local current_day = current_date_table.day

-- 构建指定时间的日期时间表
local target_time_table = {
    year = current_year,
    month = current_month,
    day = current_day,
    hour = 0,
    min = 0,
    sec = 1
}

function build_start_timestamp()
    return os.time(target_time_table)
end

-- 生成12-23小时之间的随机时间戳函数
function random_time_range(base_timestamp)
    local offset = math.random(1 * 3600, 23 * 3600) -- 1-23小时转换为秒
    return base_timestamp + offset
end
